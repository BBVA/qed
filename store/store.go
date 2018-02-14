// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

/*
	Package store implements a common API to access to multiple storage engines 
*/
package store

import (
	"verifiabledata/tree"
)

// Define the methods a storage engine must support in order to be used by the system
type Store interface {
	Add(tree.Node) error
	Get(tree.Position) (*tree.Node, error)
}
