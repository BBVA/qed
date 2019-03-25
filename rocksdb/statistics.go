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

// StatsLevel is the level of Statistics to report.
type StatsLevel uint32

const (
	// LevelExceptHistogramOrTimers disables timer stats, and skip histogram stats.
	LevelExceptHistogramOrTimers = StatsLevel(C.EXCEPT_HISTOGRAM_OR_TIMERS)
	// LevelExceptTimers skips timer stats.
	LevelExceptTimers = StatsLevel(C.EXCEPT_TIMERS)
	// LevelExceptDetailedTimers collects all stats except time inside mutex lock
	// AND time spent on compression.
	LevelExceptDetailedTimers = StatsLevel(C.EXCEPT_DETAILED_TIMERS)
	// LevelExceptTimeForMutex collect all stats except the counters requiring to get time
	// inside the mutex lock.
	LevelExceptTimeForMutex = StatsLevel(C.EXCEPT_TIME_FOR_MUTEX)
	// LevelAll collects all stats, including measuring duration of mutex operations.
	// If getting time is expensive on the platform to run, it can
	// reduce scalability to more threads, especially for writes.
	LevelAll = StatsLevel(C.ALL)
)

// Statistics is used to analyze the performance of a db. Pointer for
// statistics object is managed by Option class.
type Statistics struct {
	c *C.rocksdb_statistics_t
}

// NewStatistics is the constructor for a Statistics struct.
func NewStatistics() *Statistics {
	return &Statistics{c: C.rocksdb_create_statistics()}
}

// GetTickerCount gets the count for a ticker.
func (s *Statistics) GetTickerCount(tickerType TickerType) uint64 {
	return uint64(C.rocksdb_statistics_get_ticker_count(
		s.c,
		C.rocksdb_tickers_t(tickerType),
	))
}

// GetAndResetTickerCount get the count for a ticker and reset the tickers count.
func (s *Statistics) GetAndResetTickerCount(tickerType TickerType) uint64 {
	return uint64(C.rocksdb_statistics_get_and_reset_ticker_count(
		s.c,
		C.rocksdb_tickers_t(tickerType),
	))
}

// GetHistogramData gets the histogram data for a particular histogram.
func (s *Statistics) GetHistogramData(histogramType HistogramType) *HistogramData {
	data := NewHistogramData()
	C.rocksdb_statistics_histogram_data(
		s.c,
		C.rocksdb_histograms_t(histogramType),
		data.c,
	)
	return data
}

// Reset resets all ticker and histogram stats.
func (s *Statistics) Reset() {
	C.rocksdb_statistics_reset(s.c)
}

// StatsLevel gets the current stats level.
func (s *Statistics) StatsLevel() StatsLevel {
	return StatsLevel(C.rocksdb_statistics_stats_level(s.c))
}

// SetStatsLevel sets the stats level.
func (s *Statistics) SetStatsLevel(statsLevel StatsLevel) {
	C.rocksdb_statistics_set_stats_level(
		s.c, C.rocksdb_stats_level_t(statsLevel))
}

// Destroy deallocates the Statistics object.
func (s *Statistics) Destroy() {
	C.rocksdb_statistics_destroy(s.c)
	s.c = nil
}
