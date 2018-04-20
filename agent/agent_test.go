// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package agent

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"verifiabledata/api/apihttp"
	"verifiabledata/balloon"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage/badger"
	"verifiabledata/balloon/storage/cache"
)

func init() {
	// Instantiate a server with defaults.
	// DISCUSS: it's filosophically required?
	go (func() {
		dbPath := "/tmp/testAdd"
		os.MkdirAll(dbPath, os.FileMode(0755))
		frozen := badger.NewBadgerStorage(fmt.Sprintf("%s/frozen.db", dbPath))
		leaves := badger.NewBadgerStorage(fmt.Sprintf("%s/leaves.db", dbPath))
		cache := cache.NewSimpleCache(5000000)
		balloon := balloon.NewHyperBalloon(dbPath, hashing.Sha256Hasher, frozen, leaves, cache)

		err := http.ListenAndServe(":8080", apihttp.NewApiHttp(balloon))
		if err != nil {
			log.Fatalln("Can't start HTTP Server: ", err)
		}
	})()
}

func TestAdd(t *testing.T) {
	testAgent, _ := NewAgent("http://localhost:8080")
	testAgent.Add("Hola mundo!")

}

func BenchmarkAdd(b *testing.B) {
	testAgent, _ := NewAgent("http://localhost:8080")

	for n := 0; n < b.N; n++ {
		testAgent.Add(string(n))
	}

}
