/*
   Copyright 2018-2019 Banco Bilbao Vizcaya Argentaria, S.A.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package rocks

import (
	"sync"

	"github.com/bbva/qed/rocksdb"
	"github.com/prometheus/client_golang/prometheus"
)

// The following package variables act as singletons, to be shared by all Store
// instantiations. This allows multiple stores to be instantiated within the
// same process.
var (
	rms *rocksDBMetrics
	mmu sync.RWMutex
)

// PrometheusCollectors satisfies the prom.PrometheusCollector interface.
func PrometheusCollectors() []prometheus.Collector {
	mmu.RLock()
	defer mmu.RUnlock()

	var collectors []prometheus.Collector
	if rms != nil {
		collectors = append(collectors, rms.blockCacheMetrics.PrometheusCollectors()...)
		collectors = append(collectors, rms.ioMetrics.PrometheusCollectors()...)
	}
	return collectors
}

// namespace is the leading part of all published metrics for the Storage service.
const namespace = "qed_storage"

const blockCacheSubsystem = "block" // sub-system associated with metrics for block cache.
const ioSubsystem = "io"            // sub-system associated with metrics for I/O.

type rocksDBMetrics struct {
	*blockCacheMetrics
	*ioMetrics
}

func newRocksDBMetrics(stats *rocksdb.Statistics) *rocksDBMetrics {
	return &rocksDBMetrics{
		blockCacheMetrics: newBlockCacheMetrics(stats),
		ioMetrics:         newIOMetrics(stats),
	}
}

// blockCacheMetrics are a set of metrics concerned with the block cache.
type blockCacheMetrics struct {
	BlockCacheMiss prometheus.GaugeFunc
	BlockCacheHit  prometheus.GaugeFunc
}

// newBlockCacheMetrics initialises the prometheus metris for block cache.
func newBlockCacheMetrics(stats *rocksdb.Statistics) *blockCacheMetrics {
	return &blockCacheMetrics{
		BlockCacheMiss: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_miss",
				Help:      "Total block cache misses.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheMiss),
		),
		BlockCacheHit: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_hit",
				Help:      "Total block cache hits.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheHit),
		),
	}
}

// PrometheusCollectors satisfies the prom.PrometheusCollector interface.
func (m *blockCacheMetrics) PrometheusCollectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.BlockCacheMiss,
		m.BlockCacheHit,
	}
}

type ioMetrics struct {
	BytesRead    prometheus.GaugeFunc
	BytesWritten prometheus.GaugeFunc
}

func newIOMetrics(stats *rocksdb.Statistics) *ioMetrics {
	return &ioMetrics{
		BytesRead: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: ioSubsystem,
				Name:      "bytes_read",
				Help:      "Number of uncompressed bytes read.",
			},
			extractMetric(stats, rocksdb.TickerBytesRead),
		),
		BytesWritten: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: ioSubsystem,
				Name:      "bytes_written",
				Help:      "Number of uncompressed bytes written.",
			},
			extractMetric(stats, rocksdb.TickerBytesWritten),
		),
	}
}

// PrometheusCollectors satisfies the prom.PrometheusCollector interface.
func (m *ioMetrics) PrometheusCollectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.BytesRead,
		m.BytesWritten,
	}
}

func extractMetric(stats *rocksdb.Statistics, ticker rocksdb.TickerType) func() float64 {
	return func() float64 {
		return float64(stats.GetTickerCount(ticker))
	}
}
