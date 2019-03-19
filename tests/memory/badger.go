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

package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/bbva/qed/log"
	"github.com/bbva/qed/storage"
	"github.com/bbva/qed/storage/badger"
	"github.com/bbva/qed/testutils/rand"
	b "github.com/dgraph-io/badger"
	bo "github.com/dgraph-io/badger/options"
)

func newBadgerTest(path string) (*badger.BadgerStore, error) {

	opts := b.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	opts.SyncWrites = false
	opts.ValueLogLoadingMode = bo.MemoryMap
	opts.TableLoadingMode = bo.FileIO

	opts.NumLevelZeroTables = 3
	opts.NumLevelZeroTablesStall = 6

	opts.NumCompactors = 1
	opts.MaxTableSize = .25 * 1073741824
	opts.NumMemtables = 3
	opts.ValueLogFileSize = 2 * 1073741824

	opts.SyncWrites = false

	storeOptions := &badger.Options{Path: path, BadgerOptions: &opts}
	return badger.NewBadgerStoreOpts(storeOptions)
}

func cleanup(db *b.DB) {

	_, lastVlogSize := db.Size()

	ticker := time.NewTicker(2 * time.Minute)
	const GB = int64(1 << 30)

	for {
		<-ticker.C
		_, currentVlogSize := db.Size()
		if currentVlogSize < lastVlogSize+GB {
			continue
		}

		// If size increased by 3.5 GB, then we run this 3 times.
		numTimes := (currentVlogSize - lastVlogSize) / GB
		for i := 0; i < int(numTimes); i++ {
			err := db.RunValueLogGC(0.5)
			if err != nil {
				log.Errorf("%s badgerOps unable to RunValueLogGC; %s", time.Now(), err)
			}
			log.Infof("%s CLEANUP RunValueLogGC completed iteration=%d", time.Now(), i)
		}
		_, lastVlogSize = db.Size()
	}

}

func main() {
	var counter uint64
	path := flag.String("p", "/var/tmp/memtest", "path to store database files")
	dur := flag.Duration("d", 10*time.Minute, "period of time to execute random insertions")
	pauseBeforeQuit := flag.Bool("P", false, "pause before end for debugging purposes")
	flag.Parse()

	b, _ := newBadgerTest(*path)

	// start profiler
	go func() {
		log.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	start := time.Now()
	reader := bufio.NewReader(os.Stdin)
	counter = 0

	for time.Now().Sub(start) < *dur {
		value := make([]byte, 8)
		binary.LittleEndian.PutUint64(value, 42)
		key := rand.Bytes(128)
		b.Mutate([]*storage.Mutation{{storage.IndexPrefix, key, value}})
		counter++
	}

	fmt.Println("Insertions:", counter)

	if *pauseBeforeQuit {
		fmt.Printf("Press some key to end..")
		reader.ReadString('\n')
	}
}
