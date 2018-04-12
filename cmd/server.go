// Copyright 2018 BBVA. All rights reserved.
// Use of this source code is governed by a Apache 2 License
// that can be found in the LICENSE file

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"verifiabledata/api/apihttp"
	"verifiabledata/balloon"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/storage"
	"verifiabledata/balloon/storage/badger"
	"verifiabledata/balloon/storage/bolt"
	"verifiabledata/balloon/storage/cache"
)

var (
	httpEndpoint = *flag.String("http_endpoint", ":8080", "Endpoint for REST requests on (host:port)")
	path         = *flag.String("path", "/tmp/balloon.db", "Set default storage path.")
	cach         = *flag.Int("cache", 5000000, "Initialize and reserve custom cache size.")
	stor         = *flag.String("storage", "badger", "Choose between different storage backends. Eg badge|bolt")
)

func main() {

	flag.Parse()

	var frozen, leaves storage.Store

	switch stor {
	case "badger":
		frozen = badger.NewBadgerStorage(fmt.Sprintf("%s/frozen.db", path))
		leaves = badger.NewBadgerStorage(fmt.Sprintf("%s/leaves.db", path))
	case "bolt":
		frozen = bolt.NewBoltStorage(fmt.Sprintf("%s/frozen.db", path), "forzen")
		leaves = bolt.NewBoltStorage(fmt.Sprintf("%s/leaves.db", path), "leaves")
	default:
		fmt.Print("Please select a valid storage backend")
	}

	cache := cache.NewSimpleCache(cach)

	balloon := balloon.NewHyperBalloon(path, hashing.Sha256Hasher, frozen, leaves, cache)

	err := http.ListenAndServe(httpEndpoint, apihttp.New(balloon))
	if err != nil {
		log.Fatalln("Can't start HTTP Server: ", err)
	}
}
