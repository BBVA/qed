// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package history

// stats to measure the performance of the tree and
// the associated caches and strategies
type stats struct {
	unfreezing     int
	unfreezingHits int
	freezing       int
	leafHashes     int
	internalHashes int
}
