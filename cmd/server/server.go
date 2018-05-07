// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package main

import (
	"flag"

	"verifiabledata/balloon/storage"
	"verifiabledata/server"
)

var (
	logLevel, httpEndpoint, dbPath, storageName string
	cacheSize                                   uint64
)

func main() {
	// We use the TypeVar flag syntax becouse balloon requires parameters as *type
	flag.StringVar(&httpEndpoint, "http_endpoint", ":8080", "Endpoint for REST requests on (host:port)")
	flag.StringVar(&dbPath, "path", "/tmp/balloon.db", "Set default storage path.")
	flag.Uint64Var(&cacheSize, "cache", storage.SIZE25, "Initialize and reserve custom cache size.")
	flag.StringVar(&storageName, "storage", "badger", "Choose between different storage backends. Eg badge|bolt")
	flag.StringVar(&logLevel, "log", "error", "Choose between log levels: silent, error, info and debug")
	flag.Parse()

	s, logger := server.NewServer(httpEndpoint, dbPath, cacheSize, storageName, logLevel)

	err := s.ListenAndServe()
	if err != nil {
		logger.Errorf("Can't start HTTP Server: ", err)
	}
}
