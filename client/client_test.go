// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package client

import (
	"os"
	"testing"

	"qed/log"
	"qed/server"
)

var (
	client *HttpClient
	path   string
)

func resetPath() {
	os.RemoveAll(path)
	os.MkdirAll(path, os.FileMode(0755))
}

func init() {
	path = "/var/tmp/balloonClientTest"
	resetPath()

	server := server.NewServer(":8079", path, "my-awesome-api-key", uint64(50000), "badger")

	go (func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Errorf("Can't start HTTP Server: ", err)
		}
	})()

	client = NewHttpClient("http://localhost:8079", "my-awesome-api-key")
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

func benchmarkAdd(i int, b *testing.B) {
	resetPath()
	b.ResetTimer()
	for n := 0; n < i; n++ {
		client.Add(string(n))
	}
}

func BenchmarkAdd10(b *testing.B)       { benchmarkAdd(10, b) }
func BenchmarkAdd100(b *testing.B)      { benchmarkAdd(100, b) }
func BenchmarkAdd1000(b *testing.B)     { benchmarkAdd(1000, b) }
func BenchmarkAdd10000(b *testing.B)    { benchmarkAdd(10000, b) }
func BenchmarkAdd100000(b *testing.B)   { benchmarkAdd(100000, b) }
func BenchmarkAdd10000000(b *testing.B) { benchmarkAdd(10000000, b) }

func BenchmarkVerify(b *testing.B) {
	b.ResetTimer()
	b.N = 100000
	for n := 0; n < b.N; n++ {
		snapshot, _ := client.Add(string(n))
		proof, _ := client.Membership(snapshot.Event, snapshot.Version)
		if !client.Verify(proof, snapshot) {
			b.Fatalf("correct  at %d should be true", n)
		}
	}
}
