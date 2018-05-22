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

// Package bolt implements the storage engine interface for
// github.com/coreos/bbolt
package bolt

import (
	"bytes"
	"errors"

	b "github.com/coreos/bbolt"

	"github.com/bbva/qed/balloon/storage"
	"github.com/bbva/qed/log"
)

type BoltStorage struct {
	db     *b.DB
	bucket []byte
}

func (s BoltStorage) Add(key []byte, value []byte) error {
	return s.db.Update(func(tx *b.Tx) error {
		b := tx.Bucket(s.bucket)
		err := b.Put(key, value)
		return err
	})
}

func (s BoltStorage) Get(key []byte) ([]byte, error) {
	var value []byte
	err := s.db.View(func(tx *b.Tx) error {
		b := tx.Bucket(s.bucket)
		v := b.Get(key)
		if v == nil {
			value = make([]byte, 0)
		} else {
			value = make([]byte, len(v))
			copy(value, v)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s BoltStorage) GetRange(start, end []byte) storage.LeavesSlice {
	var leaves storage.LeavesSlice

	s.db.View(func(tx *b.Tx) error {
		cursor := tx.Bucket(s.bucket).Cursor()

		for k, _ := cursor.Seek(start); k != nil && bytes.Compare(k, end) <= 0; k, _ = cursor.Next() {
			leaves = append(leaves, k)
		}

		return nil
	})

	return leaves
}

func (s BoltStorage) Delete([]byte) error { return errors.New("not implemented") }

func (s BoltStorage) Close() error {
	return s.db.Close()
}

func NewBoltStorage(path, bucketName string) *BoltStorage {
	db, err := b.Open(path, 0600, nil)
	if err != nil {
		log.Error(err)
	}

	// create bucket
	db.Update(func(tx *b.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			log.Errorf("create bucket: %s", err)
		}
		return nil
	})

	// start stats collection
	// go func() {
	// 	// Grab the initial stats.
	// 	prev := db.Stats()

	// 	for {
	// 		// Wait for 10s.
	// 		time.Sleep(10 * time.Second)

	// 		// Grab the current stats and diff them.
	// 		stats := db.Stats()
	// 		diff := stats.Sub(&prev)

	// 		// Encode stats to JSON and print to STDOUT.
	// 		json.NewEncoder(os.Stdout).Encode(diff)

	// 		// Save stats for the next loop.
	// 		prev = stats
	// 	}
	// }()

	return &BoltStorage{
		db,
		[]byte(bucketName),
	}

}
