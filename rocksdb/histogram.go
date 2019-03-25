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
import (
	"C"
)

type HistogramData struct {
	c *C.rocksdb_histogram_data_t
}

// NewHistogramData constructs a HistogramData object.
func NewHistogramData() *HistogramData {
	return &HistogramData{c: C.rocksdb_histogram_create_data()}
}

// GetAverage returns the average value.
func (d *HistogramData) GetAverage() float64 {
	return float64(C.rocksdb_histogram_get_average(d.c))
}

// GetMedian returns the median value.
func (d *HistogramData) GetMedian() float64 {
	return float64(C.rocksdb_histogram_get_median(d.c))
}

// GetPercentile95 returns the value of the percentile 95.
func (d *HistogramData) GetPercentile95() float64 {
	return float64(C.rocksdb_histogram_get_percentile95(d.c))
}

// GetPercentile99 returns the value of the percentile 99.
func (d *HistogramData) GetPercentile99() float64 {
	return float64(C.rocksdb_histogram_get_percentile99(d.c))
}

// GetStandardDeviation returns the value of the standard deviation.
func (d *HistogramData) GetStandardDeviation() float64 {
	return float64(C.rocksdb_histogram_get_stdev(d.c))
}

// GetMax returns the max value.
func (d *HistogramData) GetMax() float64 {
	return float64(C.rocksdb_histogram_get_max(d.c))
}

// GetCount returns the total number of measure.
func (d *HistogramData) GetCount() uint64 {
	return uint64(C.rocksdb_histogram_get_count(d.c))
}

// GetSum returns the sum of all measures.
func (d *HistogramData) GetSum() uint64 {
	return uint64(C.rocksdb_histogram_get_sum(d.c))
}

// GetMin returns the min value.
func (d *HistogramData) GetMin() float64 {
	return float64(C.rocksdb_histogram_get_min(d.c))
}

// Destroy deallocates the HistogramData object.
func (d *HistogramData) Destroy() {
	C.rocksdb_histogram_data_destroy(d.c)
	d.c = nil
}
