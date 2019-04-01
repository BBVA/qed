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
	"github.com/bbva/qed/rocksdb"
	"github.com/prometheus/client_golang/prometheus"
)

// namespace is the leading part of all published metrics for the Storage service.
const namespace = "qed_storage"

const blockCacheSubsystem = "block"  // sub-system associated with metrics for block cache.
const filterSubsystem = "filter"     // sub-system associated with metrics for bloom filters.
const memtableSubsystem = "memtable" // sub-system associated with metrics for memtable.
const getSubsystem = "get"           // sub-system associated with metrics for gets.
const ioSubsystem = "io"             // sub-system associated with metrics for I/O.
const compressSybsystem = "compress" // sub-system associated with metrics for compression.

type rocksDBMetrics struct {
	*blockCacheMetrics
	*bloomFilterMetrics
	*memtableMetrics
	*getsMetrics
	*ioMetrics
	*compressMetrics
}

func newRocksDBMetrics(stats *rocksdb.Statistics) *rocksDBMetrics {
	return &rocksDBMetrics{
		blockCacheMetrics:  newBlockCacheMetrics(stats),
		bloomFilterMetrics: newBloomFilterMetrics(stats),
		memtableMetrics:    newMemtableMetrics(stats),
		getsMetrics:        newGetsMetrics(stats),
		ioMetrics:          newIOMetrics(stats),
		compressMetrics:    newCompressMetrics(stats),
	}
}

// collectors satisfies the prom.PrometheusCollector interface.
func (m *rocksDBMetrics) collectors() []prometheus.Collector {
	var collectors []prometheus.Collector
	collectors = append(collectors, m.blockCacheMetrics.collectors()...)
	collectors = append(collectors, m.bloomFilterMetrics.collectors()...)
	collectors = append(collectors, m.memtableMetrics.collectors()...)
	collectors = append(collectors, m.getsMetrics.collectors()...)
	collectors = append(collectors, m.ioMetrics.collectors()...)
	collectors = append(collectors, m.compressMetrics.collectors()...)
	return collectors
}

// blockCacheMetrics are a set of metrics concerned with the block cache.
type blockCacheMetrics struct {
	Miss              prometheus.GaugeFunc
	Hit               prometheus.GaugeFunc
	Add               prometheus.GaugeFunc
	AddFailures       prometheus.GaugeFunc
	IndexMiss         prometheus.GaugeFunc
	IndexHit          prometheus.GaugeFunc
	IndexAdd          prometheus.GaugeFunc
	IndexBytesInsert  prometheus.GaugeFunc
	IndexBytesEvict   prometheus.GaugeFunc
	FilterMiss        prometheus.GaugeFunc
	FilterHit         prometheus.GaugeFunc
	FilterAdd         prometheus.GaugeFunc
	FilterBytesInsert prometheus.GaugeFunc
	FilterBytesEvict  prometheus.GaugeFunc
	DataMiss          prometheus.GaugeFunc
	DataHit           prometheus.GaugeFunc
	DataAdd           prometheus.GaugeFunc
	DataBytesInsert   prometheus.GaugeFunc
	BytesRead         prometheus.GaugeFunc
	BytesWrite        prometheus.GaugeFunc
}

// newBlockCacheMetrics initialises the prometheus metris for block cache.
func newBlockCacheMetrics(stats *rocksdb.Statistics) *blockCacheMetrics {
	return &blockCacheMetrics{
		Miss: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_miss",
				Help:      "Block cache misses.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheMiss),
		),
		Hit: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_hit",
				Help:      "Block cache hits.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheHit),
		),
		Add: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_add",
				Help:      "Block cache adds.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheAdd),
		),
		AddFailures: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_add_failures",
				Help:      "Block cache add failures.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheAddFailures),
		),
		IndexMiss: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_index_miss",
				Help:      "Block cache index misses.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheIndexMiss),
		),
		IndexHit: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_index_hit",
				Help:      "Block cache index hits.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheIndexHit),
		),
		IndexAdd: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_index_add",
				Help:      "Block cache index adds.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheIndexAdd),
		),
		IndexBytesInsert: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_index_bytes_insert",
				Help:      "Block cache index bytes inserted.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheIndexBytesInsert),
		),
		IndexBytesEvict: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_index_bytes_evict",
				Help:      "Block cache index bytes evicted.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheIndexBytesEvict),
		),
		FilterMiss: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_filter_miss",
				Help:      "Block cache filter misses.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheFilterMiss),
		),
		FilterHit: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_filter_hit",
				Help:      "Block cache filter hits.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheFilterHit),
		),
		FilterAdd: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_filter_add",
				Help:      "Block cache filter adds.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheFilterAdd),
		),
		FilterBytesInsert: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_filter_bytes_insert",
				Help:      "Block cache filter bytes inserted.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheFilterBytesInsert),
		),
		FilterBytesEvict: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_filter_bytes_evict",
				Help:      "Block cache filter bytes evicted.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheFilterBytesEvict),
		),
		DataMiss: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_data_miss",
				Help:      "Block cache data misses.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheDataMiss),
		),
		DataHit: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_data_hit",
				Help:      "Block cache data hits.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheDataHit),
		),
		DataAdd: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_data_add",
				Help:      "Block cache data adds.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheDataAdd),
		),
		DataBytesInsert: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_data_bytes_insert",
				Help:      "Block cache data bytes inserted.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheDataBytesInsert),
		),
		BytesRead: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_bytes_read",
				Help:      "Block cache bytes read.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheBytesRead),
		),
		BytesWrite: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_bytes_write",
				Help:      "Block cache bytes written.",
			},
			extractMetric(stats, rocksdb.TickerBlockCacheBytesWrite),
		),
	}
}

// collectors satisfies the prom.PrometheusCollector interface.
func (m *blockCacheMetrics) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.Miss,
		m.Hit,
		m.Add,
		m.AddFailures,
		m.IndexMiss,
		m.IndexHit,
		m.IndexAdd,
		m.IndexBytesInsert,
		m.IndexBytesEvict,
		m.FilterMiss,
		m.FilterHit,
		m.FilterAdd,
		m.FilterBytesInsert,
		m.FilterBytesEvict,
		m.DataMiss,
		m.DataHit,
		m.DataAdd,
		m.DataBytesInsert,
		m.BytesRead,
		m.BytesWrite,
	}
}

type bloomFilterMetrics struct {
	Useful           prometheus.GaugeFunc
	FullPositive     prometheus.GaugeFunc
	FullTruePositive prometheus.GaugeFunc
}

func newBloomFilterMetrics(stats *rocksdb.Statistics) *bloomFilterMetrics {
	return &bloomFilterMetrics{
		Useful: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: filterSubsystem,
				Name:      "useful",
				Help:      "Number of times bloom filter avoided reads.",
			},
			extractMetric(stats, rocksdb.TickerBloomFilterUseful),
		),
		FullPositive: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: filterSubsystem,
				Name:      "full_positive",
				Help:      "Number of times bloom fullfilter did not avoid reads.",
			},
			extractMetric(stats, rocksdb.TickerBloomFilterFullPositive),
		),
		FullTruePositive: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: filterSubsystem,
				Name:      "full_true_positive",
				Help:      "Number of times bloom full filter did not avoid reads.",
			},
			extractMetric(stats, rocksdb.TickerBloomFilterFullTruePositive),
		),
	}
}

// collectors satisfies the prom.PrometheusCollector interface.
func (m *bloomFilterMetrics) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.Useful,
		m.FullPositive,
		m.FullTruePositive,
	}
}

type memtableMetrics struct {
	Hit  prometheus.GaugeFunc
	Miss prometheus.GaugeFunc
}

func newMemtableMetrics(stats *rocksdb.Statistics) *memtableMetrics {
	return &memtableMetrics{
		Hit: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: memtableSubsystem,
				Name:      "hit",
				Help:      "Number of memtable hits.",
			},
			extractMetric(stats, rocksdb.TickerMemtableHit),
		),
		Miss: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: memtableSubsystem,
				Name:      "miss",
				Help:      "Number of memtable misses.",
			},
			extractMetric(stats, rocksdb.TickerMemtableMiss),
		),
	}
}

// collectors satisfies the prom.PrometheusCollector interface.
func (m *memtableMetrics) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.Hit,
		m.Miss,
	}
}

type getsMetrics struct {
	HitL0      prometheus.GaugeFunc
	HitL1      prometheus.GaugeFunc
	HitL2AndUp prometheus.GaugeFunc
}

func newGetsMetrics(stats *rocksdb.Statistics) *getsMetrics {
	return &getsMetrics{
		HitL0: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: getSubsystem,
				Name:      "hits_l0",
				Help:      "Number of Get() queries server by L0.",
			},
			extractMetric(stats, rocksdb.TickerGetHitL0),
		),
		HitL1: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: getSubsystem,
				Name:      "hits_l1",
				Help:      "Number of Get() queries server by L1.",
			},
			extractMetric(stats, rocksdb.TickerGetHitL1),
		),
		HitL2AndUp: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: getSubsystem,
				Name:      "hits_l2_up",
				Help:      "Number of Get() queries server by L2 and up.",
			},
			extractMetric(stats, rocksdb.TickerGetHitL2AndUp),
		),
	}
}

// collectors satisfies the prom.PrometheusCollector interface.
func (m *getsMetrics) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.HitL0,
		m.HitL1,
		m.HitL2AndUp,
	}
}

type ioMetrics struct {
	KeysWritten       prometheus.GaugeFunc
	KeysRead          prometheus.GaugeFunc
	KeysUpdated       prometheus.GaugeFunc
	BytesRead         prometheus.GaugeFunc
	BytesWritten      prometheus.GaugeFunc
	StallMicros       prometheus.GaugeFunc
	WALFileSynced     prometheus.GaugeFunc
	WALFileBytes      prometheus.GaugeFunc
	CompactReadBytes  prometheus.GaugeFunc
	CompactWriteBytes prometheus.GaugeFunc
	FlushWriteBytes   prometheus.GaugeFunc
}

func newIOMetrics(stats *rocksdb.Statistics) *ioMetrics {
	return &ioMetrics{
		KeysWritten: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: ioSubsystem,
				Name:      "keys_written",
				Help:      "Number of keys written via puts and writes.",
			},
			extractMetric(stats, rocksdb.TickerNumberKeysWritten),
		),
		KeysRead: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: ioSubsystem,
				Name:      "keys_read",
				Help:      "Number of keys read.",
			},
			extractMetric(stats, rocksdb.TickerNumberKeysRead),
		),
		KeysUpdated: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: ioSubsystem,
				Name:      "keys_updated",
				Help:      "Number of keys updated.",
			},
			extractMetric(stats, rocksdb.TickerNumberKeysUpdated),
		),
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
		StallMicros: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: ioSubsystem,
				Name:      "stall_micros",
				Help:      "Number of microseconds waiting for compaction or flush to finish.",
			},
			extractMetric(stats, rocksdb.TickerStallMicros),
		),
		WALFileSynced: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: ioSubsystem,
				Name:      "wal_files_synced",
				Help:      "Number of times WAL sync is done.",
			},
			extractMetric(stats, rocksdb.TickerWALFileSynced),
		),
		WALFileBytes: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: ioSubsystem,
				Name:      "wal_file_bytes",
				Help:      "Number of bytes written to WAL.",
			},
			extractMetric(stats, rocksdb.TickerWALFileBytes),
		),
		CompactReadBytes: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: ioSubsystem,
				Name:      "compact_read_bytes",
				Help:      "Number of bytes read during compaction.",
			},
			extractMetric(stats, rocksdb.TickerCompactReadBytes),
		),
		CompactWriteBytes: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: ioSubsystem,
				Name:      "compact_write_bytes",
				Help:      "Number of bytes written during compaction.",
			},
			extractMetric(stats, rocksdb.TickerCompactWriteBytes),
		),
		FlushWriteBytes: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: ioSubsystem,
				Name:      "compact_flush_bytes",
				Help:      "Number of bytes written during flush.",
			},
			extractMetric(stats, rocksdb.TickerFlushWriteBytes),
		),
	}
}

// collectors satisfies the prom.PrometheusCollector interface.
func (m *ioMetrics) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.KeysWritten,
		m.KeysRead,
		m.KeysUpdated,
		m.BytesRead,
		m.BytesWritten,
		m.StallMicros,
		m.WALFileSynced,
		m.WALFileBytes,
		m.CompactReadBytes,
		m.CompactWriteBytes,
		m.FlushWriteBytes,
	}
}

type compressMetrics struct {
	NumberBlockCompressed    prometheus.GaugeFunc
	NumberBlockDecompressed  prometheus.GaugeFunc
	NumberBlockNotCompressed prometheus.GaugeFunc
}

func newCompressMetrics(stats *rocksdb.Statistics) *compressMetrics {
	return &compressMetrics{
		NumberBlockCompressed: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: compressSybsystem,
				Name:      "block_compressed",
				Help:      "Number of compressions executed",
			},
			extractMetric(stats, rocksdb.TickerNumberBlockCompressed),
		),
		NumberBlockDecompressed: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: compressSybsystem,
				Name:      "block_decompressed",
				Help:      "Number of decompressions executed",
			},
			extractMetric(stats, rocksdb.TickerNumberBlockDecompressed),
		),
		NumberBlockNotCompressed: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: compressSybsystem,
				Name:      "block_not_compressed",
				Help:      "Number of blocks not compressed.",
			},
			extractMetric(stats, rocksdb.TickerNumberBlockNotCompressed),
		),
	}
}

// collectors satisfies the prom.PrometheusCollector interface.
func (m *compressMetrics) collectors() []prometheus.Collector {
	return []prometheus.Collector{
		m.NumberBlockCompressed,
		m.NumberBlockDecompressed,
		m.NumberBlockNotCompressed,
	}
}

func extractMetric(stats *rocksdb.Statistics, ticker rocksdb.TickerType) func() float64 {
	return func() float64 {
		return float64(stats.GetAndResetTickerCount(ticker))
	}
}
