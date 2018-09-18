/*
   Copyright 2018 Banco Bilbao Vizcaya Argentaria, S.A.

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

package storage

import (
	"fmt"
	"os"

	"github.com/stretchr/testify/require"

	bd "github.com/bbva/qed/storage/badger"
	bp "github.com/bbva/qed/storage/bplus"
)

func OpenBPlusTreeStore() (*bp.BPlusTreeStore, func()) {
	store := bp.NewBPlusTreeStore()
	return store, func() {
		store.Close()
	}
}

func OpenBadgerStore(t require.TestingT, path string) (*bd.BadgerStore, func()) {
	store, err := bd.NewBadgerStore(path)
	if err != nil {
		t.Errorf("Error opening badger store: %v", err)
		t.FailNow()
	}
	return store, func() {
		store.Close()
		deleteFile(path)
	}
}

func deleteFile(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Printf("Unable to remove db file %s", err)
	}
}
