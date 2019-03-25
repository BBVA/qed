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

package rocksdb

// #include "extended.h"
// #include <stdlib.h>
import (
	"C"
)

type TickerType uint32

const (
	// The number of uncompressed bytes issued by db.Put(), db.Delete(),
	// db.Merge(), and db.Write().
	TickerBytesWritten = TickerType(C.BYTES_WRITTEN)
	// TickerBytesRead is the number of uncompressed bytes read from db.Get().
	// It could be either from memtables, cache, or table files.
	// For the number of logical bytes read from db.MultiGet(),
	// please use NumberMultiGetBytesRead.
	TickerBytesRead = TickerType(C.BYTES_READ)
)

type HistogramType uint32

const (
	HistogramBytesPerRead = HistogramType(C.BYTES_PER_READ)
)

type StatsLevel uint32

const (
	// All collect all stats, including measuring duration of mutex operations.
	// If getting time is expensive on the platform to run, it can
	// reduce scalability to more threads, especially for writes.
	LevelAll = StatsLevel(C.ALL)
)

type Statistics struct {
	c *C.rocksdb_statistics_t
}

func NewStatistics() *Statistics {
	return &Statistics{c: C.rocksdb_create_statistics()}
}

func (s *Statistics) GetTickerCount(tickerType TickerType) uint64 {
	return uint64(C.rocksdb_statistics_get_ticker_count(
		s.c,
		C.rocksdb_tickers_t(tickerType),
	))
}

func (s *Statistics) GetAndResetTickerCount(tickerType TickerType) uint64 {
	return uint64(C.rocksdb_statistics_get_and_reset_ticker_count(
		s.c,
		C.rocksdb_tickers_t(tickerType),
	))
}

func (s *Statistics) GetHistogramData(histogramType HistogramType) *HistogramData {
	data := NewHistogramData()
	C.rocksdb_statistics_histogram_data(
		s.c,
		C.rocksdb_histograms_t(histogramType),
		data.c,
	)
	return data
}

func (s *Statistics) Reset() {
	C.rocksdb_statistics_reset(s.c)
}

func (s *Statistics) StatsLevel() StatsLevel {
	return StatsLevel(C.rocksdb_statistics_stats_level(s.c))
}

func (s *Statistics) SetStatsLevel(statsLevel StatsLevel) {
	C.rocksdb_statistics_set_stats_level(
		s.c, C.rocksdb_stats_level_t(statsLevel))
}

func (s *Statistics) Destroy() {
	C.rocksdb_statistics_destroy(s.c)
	s.c = nil
}

type HistogramData struct {
	c *C.rocksdb_histogram_data_t
}

func NewHistogramData() *HistogramData {
	return &HistogramData{c: C.rocksdb_histogram_create_data()}
}

func (d *HistogramData) GetAverage() float64 {
	return float64(C.rocksdb_histogram_get_average(d.c))
}

func (d *HistogramData) GetMedian() float64 {
	return float64(C.rocksdb_histogram_get_median(d.c))
}

func (d *HistogramData) GetPercentile95() float64 {
	return float64(C.rocksdb_histogram_get_percentile95(d.c))
}

func (d *HistogramData) GetPercentile99() float64 {
	return float64(C.rocksdb_histogram_get_percentile99(d.c))
}

func (d *HistogramData) GetStandardDeviation() float64 {
	return float64(C.rocksdb_histogram_get_stdev(d.c))
}

func (d *HistogramData) GetMax() float64 {
	return float64(C.rocksdb_histogram_get_max(d.c))
}

func (d *HistogramData) GetCount() uint64 {
	return uint64(C.rocksdb_histogram_get_count(d.c))
}

func (d *HistogramData) GetSum() uint64 {
	return uint64(C.rocksdb_histogram_get_sum(d.c))
}

func (d *HistogramData) GetMin() float64 {
	return float64(C.rocksdb_histogram_get_min(d.c))
}

func (d *HistogramData) Destroy() {
	C.rocksdb_histogram_data_destroy(d.c)
	d.c = nil
}
