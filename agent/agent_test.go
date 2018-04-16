// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package agent

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	"verifiabledata/api/apihttp"
	"verifiabledata/balloon"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage/badger"
	"verifiabledata/balloon/storage/cache"
)

func dummyServer() {

	dbPath := "/tmp/testAdd"
	frozen := badger.NewBadgerStorage(fmt.Sprintf("%s/frozen.db", dbPath))
	leaves := badger.NewBadgerStorage(fmt.Sprintf("%s/leaves.db", dbPath))
	cache := cache.NewSimpleCache(5000000)
	balloon := balloon.NewHyperBalloon(dbPath, hashing.Sha256Hasher, frozen, leaves, cache)

	err := http.ListenAndServe(":8080", apihttp.New(balloon))
	if err != nil {
		log.Fatalln("Can't start HTTP Server: ", err)
	}
}

var flag bool

func TestAdd(t *testing.T) {
	if !flag {
		go dummyServer()
		flag = true
	}

	testAgent, _ := NewAgent("http://localhost:8080")
	testAgent.Add("Hola mundo!")

}

func BenchmarkAdd(b *testing.B) {

	if !flag {
		go dummyServer()
		flag = true
	}

	testAgent, _ := NewAgent("http://localhost:8080")

	for n := 0; n < b.N; n++ {
		testAgent.Add(string(n))
	}

}
