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

package raftrocks

import (
	"fmt"
	"strconv"

	"github.com/bbva/qed/rocksdb"
	"github.com/prometheus/client_golang/prometheus"
)

// namespace is the leading part of all published metrics for the Storage service.
const namespace = "qed_wal"

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
	tables []*perTableMetrics
}

func newRocksDBMetrics(store *RocksDBStore) *rocksDBMetrics {
	tables := make([]*perTableMetrics, 0)
	tables = append(tables, newPerTableMetrics(stableTable, store))
	tables = append(tables, newPerTableMetrics(logTable, store))
	return &rocksDBMetrics{
		blockCacheMetrics:  newBlockCacheMetrics(store.stats, store.blockCache),
		bloomFilterMetrics: newBloomFilterMetrics(store.stats),
		memtableMetrics:    newMemtableMetrics(store.stats),
		getsMetrics:        newGetsMetrics(store.stats),
		ioMetrics:          newIOMetrics(store.stats),
		compressMetrics:    newCompressMetrics(store.stats),
		tables:             tables,
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
	for _, table := range m.tables {
		collectors = append(collectors, table.collectors()...)
	}
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
	Usage             prometheus.GaugeFunc
	PinnedUsage       prometheus.GaugeFunc
}

// newBlockCacheMetrics initialises the prometheus metris for block cache.
func newBlockCacheMetrics(stats *rocksdb.Statistics, cache *rocksdb.Cache) *blockCacheMetrics {
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
		Usage: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_memory_usage",
				Help:      "Block cache memory usage.",
			},
			func() float64 {
				return float64(cache.GetUsage())
			},
		),
		PinnedUsage: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: blockCacheSubsystem,
				Name:      "cache_pinned_memory_usage",
				Help:      "Block cache pinned memory usage.",
			},
			func() float64 {
				return float64(cache.GetPinnedUsage())
			},
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
		m.Usage,
		m.PinnedUsage,
	}
}

// bloomFilterMetrics are a set of metrics concerned with bloom filters.
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
				Help:      "Number of times bloom fullfilter did not avoid reads and data actually exist.",
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

// memtableMetrics are a set of metrics concerned with rocksDB memtable.
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

// getsMetrics are a set of metrics concerned with rocksDB SST Level0 (L0)
// and upper levels.
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

// ioMetrics are a set of metrics concerned with IO.
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

// compressMetrics are a set of metrics concerned with rocksDB data compression.
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

// perTableMetrics are a set of metrics concerned with rocksDB tables.
type perTableMetrics struct {
	NumFilesAtLevelN                []prometheus.GaugeFunc
	NumImmutableMemtables           prometheus.GaugeFunc
	NumImmutableMemtablesFlushed    prometheus.GaugeFunc
	NumRunningFlushes               prometheus.GaugeFunc
	NumRunningCompactions           prometheus.GaugeFunc
	CurrentSizeActiveMemtable       prometheus.GaugeFunc
	CurrentSizeAllMemtables         prometheus.GaugeFunc
	SizeAllMemtables                prometheus.GaugeFunc
	NumEntriesActiveMemtable        prometheus.GaugeFunc
	NumEntriesImmutableMemtables    prometheus.GaugeFunc
	EstimatedNumKeys                prometheus.GaugeFunc
	EstimateTableReadersMem         prometheus.GaugeFunc
	NumLiveVersions                 prometheus.GaugeFunc
	EstimatedLiveDataSize           prometheus.GaugeFunc
	TotalSSTFilesSize               prometheus.GaugeFunc
	TotalLiveSSTFilesSize           prometheus.GaugeFunc
	EstimatedPendingCompactionBytes prometheus.GaugeFunc
	ActualDelayedWriteRate          prometheus.GaugeFunc
	BlockCacheUsage                 prometheus.GaugeFunc
	BlockCachePinnedUsage           prometheus.GaugeFunc
}

func newPerTableMetrics(table table, store *RocksDBStore) *perTableMetrics {
	m := &perTableMetrics{
		NumImmutableMemtables: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "num_immutable_memtables",
				Help:      "Number of immutable memtables that have not yet been flushed.",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.num-immutable-mem-table", store.cfHandles[table]))
			},
		),
		NumImmutableMemtablesFlushed: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "num_immutable_memtables_flushed",
				Help:      "Number of immutable memtables that have already been flushed.",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.num-immutable-mem-table-flushed", store.cfHandles[table]))
			},
		),
		NumRunningFlushes: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "num_running_flushes",
				Help:      "Number of currently running flushes.",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.num-running-flushes", store.cfHandles[table]))
			},
		),
		NumRunningCompactions: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "num_running_compactions",
				Help:      "Number of currently running compactions.",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.num-running-compactions", store.cfHandles[table]))
			},
		),
		CurrentSizeActiveMemtable: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "cur_size_active_memtable",
				Help:      "Approximate size of active memtable (bytes).",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.cur-size-active-mem-table", store.cfHandles[table]))
			},
		),
		CurrentSizeAllMemtables: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "cur_size_all_memtables",
				Help:      "Approximate size of active and unflushed immutable memtables (bytes).",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.cur-size-all-mem-tables", store.cfHandles[table]))
			},
		),
		SizeAllMemtables: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "size_all_memtables",
				Help:      "Approximate size of active, unflushed immutable, and pinned immutable memtables (bytes).",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.size-all-mem-tables", store.cfHandles[table]))
			},
		),
		NumEntriesActiveMemtable: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "num_entries_active_memtable",
				Help:      "Total number of entries in the active memtable.",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.num-entries-active-mem-table", store.cfHandles[table]))
			},
		),
		NumEntriesImmutableMemtables: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "num_entries_imm_memtables",
				Help:      "Total number of entries in the unflushed immutable memtables.",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.num-entries-imm-mem-tables", store.cfHandles[table]))
			},
		),
		EstimatedNumKeys: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "estimated_num_keys",
				Help:      "Estimated number of total keys in the active and unflushed immutable memtables and storage.",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.estimate-num-keys", store.cfHandles[table]))
			},
		),
		EstimateTableReadersMem: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "estimated_table_readers_mem",
				Help:      "Estimated memory used for reading SST tables, excluding memory used in block cache (e.g., filter and index blocks).",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.estimate-table-readers-mem", store.cfHandles[table]))
			},
		),
		NumLiveVersions: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "num_live_versions",
				Help:      "Number of live versions.",
				// A `Version` is an internal data structure. More live versions often mean more SST files
				// are held from being deleted, by iterators or unfinished compactions.
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.num-live-versions", store.cfHandles[table]))
			},
		),
		EstimatedLiveDataSize: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "estimated_live_data_size",
				Help:      "Estimate of the amount of live data (bytes).",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.estimate-live-data-size", store.cfHandles[table]))
			},
		),
		TotalSSTFilesSize: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "total_sst_files_size",
				Help:      "Total size (bytes) of all SST files.",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.total-sst-files-size", store.cfHandles[table]))
			},
		),
		TotalLiveSSTFilesSize: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "total_live_sst_files_size",
				Help:      "Total size (bytes) of all live SST files that belongs to theh last LSM tree.",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.live-sst-files-size", store.cfHandles[table]))
			},
		),
		EstimatedPendingCompactionBytes: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "estimated_pending_compaction_bytes",
				Help:      "Estimated total number of bytes compaction needs to rewrite to get all levels down to under target size.",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.estimate-pending-compaction-bytes", store.cfHandles[table]))
			},
		),
		ActualDelayedWriteRate: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "actual_delayed_write_rate",
				Help:      "Current actual delayed write rate. 0 means no delay.",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.actual-delayed-write-rate", store.cfHandles[table]))
			},
		),
		BlockCacheUsage: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "block_cache_usage",
				Help:      "Memory size (bytes) for the entries residing in block cache.",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.block-cache-usage", store.cfHandles[table]))
			},
		),
		BlockCachePinnedUsage: prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      "block_cache_pinned_usage",
				Help:      "Memory size (bytes) for the entries being pinned in block cache.",
			},
			func() float64 {
				return float64(store.db.GetUint64PropertyCF("rocksdb.block-cache-pinned-usage", store.cfHandles[table]))
			},
		),
	}
	numFileAtLevels := make([]prometheus.GaugeFunc, 0)
	for i := 0; i <= 5; i++ {
		propName := fmt.Sprintf("rocksdb.num-files-at-level%d", i)
		numFileAtLevels = append(numFileAtLevels, prometheus.NewGaugeFunc(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Subsystem: "cf_" + table.String(),
				Name:      fmt.Sprintf("num_files_at_level_%d", i),
				Help:      fmt.Sprintf("Number of files at level %d.", i),
			},
			func() float64 {
				sValue := store.db.GetPropertyCF(propName, store.cfHandles[table])
				if sValue != "" {
					value, _ := strconv.ParseFloat(sValue, 64)
					return value
				}
				return 0.0
			},
		))
	}
	m.NumFilesAtLevelN = numFileAtLevels
	return m
}

// collectors satisfies the prom.PrometheusCollector interface.
func (m *perTableMetrics) collectors() []prometheus.Collector {
	c := []prometheus.Collector{
		m.NumImmutableMemtables,
		m.NumImmutableMemtablesFlushed,
		m.NumRunningFlushes,
		m.NumRunningCompactions,
		m.CurrentSizeActiveMemtable,
		m.CurrentSizeAllMemtables,
		m.SizeAllMemtables,
		m.NumEntriesActiveMemtable,
		m.NumEntriesImmutableMemtables,
		m.EstimatedNumKeys,
		m.EstimateTableReadersMem,
		m.NumLiveVersions,
		m.EstimatedLiveDataSize,
		m.TotalSSTFilesSize,
		m.TotalLiveSSTFilesSize,
		m.EstimatedPendingCompactionBytes,
		m.ActualDelayedWriteRate,
		m.BlockCacheUsage,
		m.BlockCachePinnedUsage,
	}
	for _, metric := range m.NumFilesAtLevelN {
		c = append(c, metric)
	}
	return c
}

func extractMetric(stats *rocksdb.Statistics, ticker rocksdb.TickerType) func() float64 {
	return func() float64 {
		return float64(stats.GetAndResetTickerCount(ticker))
	}
}
