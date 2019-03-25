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

// HistogramType is the logical mapping of histograms defined in rocksdb:Histogram.
type HistogramType uint32

const (
	// HistogramBytesPerRead is value size distribution in read operations.
	HistogramBytesPerRead = HistogramType(C.BYTES_PER_READ)
)
