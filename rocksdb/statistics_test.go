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

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStatsLevel(t *testing.T) {
	stats := NewStatistics()
	stats.SetStatsLevel(LevelAll)
	require.Equal(t, stats.StatsLevel(), LevelAll)
}

func TestStatsGetTickerCount(t *testing.T) {

	stats := NewStatistics()
	db, _ := newTestDB(t, "TestStatsGetTickerCount", func(opts *Options) {
		opts.SetStatistics(stats)
	})
	defer db.Close()
	defer stats.Destroy()

	key := []byte("some-key")
	val := []byte("some-value")

	require.NoError(t, db.Put(NewDefaultWriteOptions(), key, val))
	ro := NewDefaultReadOptions()
	defer ro.Destroy()
	for i := 0; i < 10; i++ {
		_, _ = db.Get(ro, key)
	}

	require.True(t, stats.GetTickerCount(TickerBytesRead) > 0)

}

func TestGetAndResetTickerCount(t *testing.T) {

	stats := NewStatistics()
	db, _ := newTestDB(t, "TestGetAndResetTickerCount", func(opts *Options) {
		opts.SetStatistics(stats)
	})
	defer db.Close()
	defer stats.Destroy()

	key := []byte("some-key")
	val := []byte("some-value")

	require.NoError(t, db.Put(NewDefaultWriteOptions(), key, val))
	ro := NewDefaultReadOptions()
	defer ro.Destroy()
	for i := 0; i < 10; i++ {
		_, _ = db.Get(ro, key)
	}

	read := stats.GetAndResetTickerCount(TickerBytesRead)
	require.True(t, read > 0)
	require.True(t, stats.GetAndResetTickerCount(TickerBytesRead) < read)

}

func TestGetHistogramData(t *testing.T) {

	t.Skip() // not working

	stats := NewStatistics()
	db, _ := newTestDB(t, "TestGetHistogramData", func(opts *Options) {
		opts.SetStatistics(stats)
	})
	defer db.Close()
	defer stats.Destroy()

	key := []byte("some-key")
	val := []byte("some-value")

	require.NoError(t, db.Put(NewDefaultWriteOptions(), key, val))
	ro := NewDefaultReadOptions()
	defer ro.Destroy()
	for i := 0; i < 10; i++ {
		_, _ = db.Get(ro, key)
	}

	histogramData := stats.GetHistogramData(HistogramBytesPerRead)
	defer histogramData.Destroy()
	require.NotNil(t, histogramData)
	require.True(t, histogramData.GetAverage() > 0)
	require.True(t, histogramData.GetMedian() > 0)
	require.True(t, histogramData.GetPercentile95() > 0)
	require.True(t, histogramData.GetPercentile99() > 0)
	require.True(t, histogramData.GetStandardDeviation() == 0.00)
	require.True(t, histogramData.GetMax() > 0)
	require.True(t, histogramData.GetCount() > 0)
	require.True(t, histogramData.GetSum() > 0)
}

func TestReset(t *testing.T) {
	stats := NewStatistics()
	db, _ := newTestDB(t, "TestReset", func(opts *Options) {
		opts.SetStatistics(stats)
	})
	defer db.Close()
	defer stats.Destroy()

	key := []byte("some-key")
	val := []byte("some-value")

	require.NoError(t, db.Put(NewDefaultWriteOptions(), key, val))
	ro := NewDefaultReadOptions()
	defer ro.Destroy()
	for i := 0; i < 10; i++ {
		_, _ = db.Get(ro, key)
	}

	read := stats.GetAndResetTickerCount(TickerBytesRead)
	require.True(t, read > 0)

	stats.Reset()

	readAfterReset := stats.GetAndResetTickerCount(TickerBytesRead)
	require.True(t, readAfterReset < read)
}
