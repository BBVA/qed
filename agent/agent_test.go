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
	"verifiabledata/balloon/history"
	"verifiabledata/balloon/hyper"
	"verifiabledata/balloon/storage/badger"
	"verifiabledata/balloon/storage/cache"
)

var testAgent *Agent

func init() {
	// Instantiate a server with defaults.
	go (func() {
		dbPath := "/tmp/testAdd"
		os.MkdirAll(dbPath, os.FileMode(0755))

		frozen := badger.NewBadgerStorage(fmt.Sprintf("%s/frozen.db", dbPath))
		leaves := badger.NewBadgerStorage(fmt.Sprintf("%s/leaves.db", dbPath))
		cache := cache.NewSimpleCache(5000000)
		hasher := hashing.Sha256Hasher
		history := history.NewTree(frozen, history.LeafHasherF(hasher), history.InteriorHasherF(hasher))
		hyper := hyper.NewTree(dbPath, 30, cache, leaves, hasher, hyper.LeafHasherF(hasher), hyper.InteriorHasherF(hasher))
		balloon := balloon.NewHyperBalloon(hasher, history, hyper)

		err := http.ListenAndServe(":8079", apihttp.NewApiHttp(balloon))
		if err != nil {
			log.Fatalln("Can't start HTTP Server: ", err)
		}
	})()

	testAgent, _ = NewAgent("http://localhost:8079")
}

func TestAdd(t *testing.T) {
	testAgent.Add("Ping Pong")
}

func TestMembership(t *testing.T) {
	event := "King Pong"
	testAgent.Add(event)

	record := testAgent.storage[string(testAgent.hasher([]byte(event)))]

	testAgent.MembershipProof(record.event, record.commitment.Version)

}

func TestVerify(t *testing.T) {

	event := "Donkey Pong"
	testAgent.Add(event)

	record := testAgent.storage[string(testAgent.hasher([]byte(event)))]

	testAgent.MembershipProof(record.event, record.commitment.Version)

	if !testAgent.Verify(record.proof, record.commitment, record.event) {
		t.Error("Can't verify the Membership Proof")
	}

}

func BenchmarkAdd(b *testing.B) {
	for n := 0; n < b.N; n++ {
		testAgent.Add(string(n))
	}
}
