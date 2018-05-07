// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package client

import (
	"os"
	"testing"

	"verifiabledata/server"
)

var (
	client *HttpClient
)

func init() {
	path := "/tmp/balloonClientTest"
	os.MkdirAll(path, os.FileMode(0755))

	server, logger := server.NewServer(":8079", path, uint64(50000), "badger", "silent")

	go (func() {
		err := server.ListenAndServe()
		if err != nil {
			logger.Errorf("Can't start HTTP Server: ", err)
		}
	})()

	client = NewHttpClient("http://localhost:8079", "my-awesome-api-key", logger)
}

func TestAdd(t *testing.T) {
	client.Add("Hola mundo!")
}

func TestMembership(t *testing.T) {
	snapshot, err := client.Add("Hola mundo!")
	if err != nil {
		t.Fatal(err)
	}

	proof, err := client.Membership(snapshot.Event, snapshot.Version)
	if err != nil {
		t.Fatal(err)
	}

	if !proof.Exists {
		t.Fatal("It should exist")
	}

}

func TestVerify(t *testing.T) {
	snapshot, err := client.Add("Hola mundo!")
	if err != nil {
		t.Fatal(err)
	}

	proof, err := client.Membership(snapshot.Event, snapshot.Version)
	if err != nil {
		t.Fatal(err)
	}

	correct := client.Verify(proof, snapshot)

	if !correct {
		t.Fatal("correct should be true")
	}
}

func BenchmarkAdd(b *testing.B) {
	b.N = 10000
	for n := 0; n < b.N; n++ {
		client.Add(string(n))
	}
}

func BenchmarkAddAndVerify(b *testing.B) {
	b.ResetTimer()
	b.N = 10000
	for n := 0; n < b.N; n++ {
		snapshot, _ := client.Add(string(n))
		proof, _ := client.Membership(snapshot.Event, snapshot.Version)
		correct := client.Verify(proof, snapshot)
		if !correct {
			b.Fatal("correct should be true")
		}
	}
}
