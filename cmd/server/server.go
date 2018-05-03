// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"verifiabledata/api/apihttp"
	"verifiabledata/balloon"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/history"
	"verifiabledata/balloon/hyper"
	"verifiabledata/balloon/storage"
	"verifiabledata/balloon/storage/badger"
	"verifiabledata/balloon/storage/bolt"
	"verifiabledata/balloon/storage/cache"
	"verifiabledata/log"
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
	flag.StringVar(&logLevel, "log", "silent", "Choose between log levels: silent, error, info and debug")
	flag.Parse()

	var frozen, leaves storage.Store
	var l log.Logger

	switch storageName {
	case "badger":
		frozen = badger.NewBadgerStorage(fmt.Sprintf("%s/frozen.db", dbPath))
		leaves = badger.NewBadgerStorage(fmt.Sprintf("%s/leaves.db", dbPath))
	case "bolt":
		frozen = bolt.NewBoltStorage(fmt.Sprintf("%s/frozen.db", dbPath), "frozen")
		leaves = bolt.NewBoltStorage(fmt.Sprintf("%s/leaves.db", dbPath), "leaves")
	default:
		fmt.Print("Please select a valid storage backend")
		os.Exit(-1)
	}

	switch logLevel {
	case "silent":
		l = log.NewSilent()
	case "error":
		l = log.NewError(os.Stdout, "Server: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	case "info":
		l = log.NewInfo(os.Stdout, "Server: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	case "debug":
		l = log.NewDebug(os.Stdout, "Server: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
	default:
		fmt.Print("Please select a valid log level")
		os.Exit(-1)
	}

	cache := cache.NewSimpleCache(cacheSize)
	hasher := hashing.Sha256Hasher
	history := history.NewTree(frozen, hasher, l)
	hyper := hyper.NewTree(dbPath, cache, leaves, hasher, l)
	balloon := balloon.NewHyperBalloon(hasher, history, hyper, l)

	// start profiler
	go func() {
		l.Info(http.ListenAndServe("localhost:6060", nil))
	}()

	router := apihttp.NewApiHttp(balloon)

	s := &http.Server{
		Addr:    httpEndpoint,
		Handler: logHandler(router, l),
	}

	err := s.ListenAndServe()
	if err != nil {
		l.Errorf("Can't start HTTP Server: ", err)
	}
}

type statusWriter struct {
	http.ResponseWriter
	status int
	length int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	w.length = len(b)
	return w.ResponseWriter.Write(b)
}

// WriteLog Logs the Http Status for a request into fileHandler and returns a httphandler function which is a wrapper to log the requests.
func logHandler(handle http.Handler, l log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		start := time.Now()
		writer := statusWriter{w, 0, 0}
		handle.ServeHTTP(&writer, request)
		latency := time.Now().Sub(start)

		l.Debugf("Request: lat %d %+v", latency, request)
		if writer.status >= 400 {
			l.Infof("Bad Request: %d %+v", latency, request)
		}
	}
}
