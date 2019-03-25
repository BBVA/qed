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

// TickerType is the logical mapping of tickers defined in rocksdb::Tickers.
type TickerType uint32

const (
	// TickerBytesWritten is the number of uncompressed bytes issued by db.Put(),
	// db.Delete(), db.Merge(), and db.Write().
	TickerBytesWritten = TickerType(C.BYTES_WRITTEN)
	// TickerBytesRead is the number of uncompressed bytes read from db.Get().
	// It could be either from memtables, cache, or table files.
	// For the number of logical bytes read from db.MultiGet(),
	// please use NumberMultiGetBytesRead.
	TickerBytesRead = TickerType(C.BYTES_READ)
)
