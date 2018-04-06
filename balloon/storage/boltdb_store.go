package storage

import (
	"bytes"
	"fmt"
	"log"

	bolt "github.com/coreos/bbolt"
)

type BoltStorage struct {
	db     *bolt.DB
	bucket []byte
}

func (s *BoltStorage) Add(key []byte, value []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)
		err := b.Put(key, value)
		return err
	})
}

func (s *BoltStorage) Get(key []byte) ([]byte, error) {
	var value []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)
		v := b.Get(key)
		if v == nil {
			return fmt.Errorf("Unknown key %d", key)
		}
		value = make([]byte, len(v))
		copy(value, v)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (s *BoltStorage) GetRange(start, end []byte) LeavesSlice {
	var leaves LeavesSlice

	s.db.View(func(tx *bolt.Tx) error {
		cursor := tx.Bucket(s.bucket).Cursor()

		for k, _ := cursor.Seek(start); k != nil && bytes.Compare(k, end) <= 0; k, _ = cursor.Next() {
			leaves = append(leaves, k)
		}

		return nil
	})

	return leaves
}

func NewBoltStorage(path, bucketName string) *BoltStorage {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	// create bucket
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
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
