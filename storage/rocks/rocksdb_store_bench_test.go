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

package rocks

import (
	"context"
	"fmt"
	"log"
	rnd "math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bbva/qed/crypto/hashing"

	"github.com/bbva/qed/storage"
	"github.com/bbva/qed/testutils/rand"
	"github.com/bbva/qed/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func BenchmarkMutateOnlyHyper(b *testing.B) {
	store, closeF := openRocksDBStore(b)
	defer closeF()

	srvCloseF := startMetricsServer(store)
	defer srvCloseF()

	b.N = 10000000
	b.ResetTimer()

	hasher := hashing.NewFakeSha256Hasher()
	value := rand.Bytes(1024)
	for i := 0; i < b.N; i++ {
		key := util.Uint16AsBytes(uint16(rnd.Intn(10)))
		key = append(key, hasher.Do([]byte(fmt.Sprintf("test%d", rnd.Intn(10000))))...)
		store.Mutate([]*storage.Mutation{
			{
				Table: storage.HyperTable,
				Key:   key,
				Value: value,
			},
		})
	}

}

func BenchmarkQueryOnlyHyper(b *testing.B) {
	store, closeF := openRocksDBStore(b)
	defer closeF()

	N := 10000000
	b.N = N
	hasher := hashing.NewFakeSha256Hasher()

	srvCloseF := startMetricsServer(store)
	defer srvCloseF()

	// populate storage
	value := rand.Bytes(1024)
	for i := 0; i < b.N; i++ {
		key := []byte{0x0, 0x0}
		key = append(key, hasher.Do([]byte(fmt.Sprintf("test%d", rnd.Intn(10000))))...)
		store.Mutate([]*storage.Mutation{
			{
				Table: storage.HyperTable,
				Key:   key,
				Value: value,
			},
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := []byte{0x0, 0x0}
		key = append(key, hasher.Do([]byte(fmt.Sprintf("test%d", rnd.Intn(1000))))...)
		_, err := store.Get(storage.HyperTable, key)
		require.NoError(b, err)
	}

}

func BenchmarkMutateOnlyHistory(b *testing.B) {
	store, closeF := openRocksDBStore(b)
	defer closeF()

	srvCloseF := startMetricsServer(store)
	defer srvCloseF()

	b.N = 10000000
	b.ResetTimer()

	hasher := hashing.NewFakeSha256Hasher()
	for i := 0; i < b.N; i++ {
		key := util.Uint64AsBytes(uint64(i))
		key = append(key, []byte{0x0, 0x0}...)
		store.Mutate([]*storage.Mutation{
			{
				Table: storage.HistoryTable,
				Key:   key,
				Value: hasher.Do([]byte(fmt.Sprintf("test%d", i))),
			},
		})
	}

}

func BenchmarkQueryOnlyHistory(b *testing.B) {
	store, closeF := openRocksDBStore(b)
	defer closeF()

	N := 10000000
	b.N = N
	hasher := hashing.NewFakeSha256Hasher()

	srvCloseF := startMetricsServer(store)
	defer srvCloseF()

	// populate storage
	for i := 0; i < b.N; i++ {
		key := util.Uint64AsBytes(uint64(i))
		key = append(key, []byte{0x0, 0x0}...)
		store.Mutate([]*storage.Mutation{
			{
				Table: storage.HistoryTable,
				Key:   key,
				Value: hasher.Do([]byte(fmt.Sprintf("test%d", i))),
			},
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		index := rnd.Intn(N)
		key := util.Uint64AsBytes(uint64(index))
		key = append(key, []byte{0x0, 0x0}...)
		_, err := store.Get(storage.HistoryTable, key)
		require.NoError(b, err)
	}

}

func BenchmarkMutateOnlyFSMState(b *testing.B) {
	store, closeF := openRocksDBStore(b)
	defer closeF()

	srvCloseF := startMetricsServer(store)
	defer srvCloseF()

	b.N = 1000000
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.Mutate([]*storage.Mutation{
			{
				Table: storage.FSMStateTable,
				Key:   storage.FSMStateTableKey,
				Value: rand.Bytes(24),
			},
		})
	}

}

func BenchmarkQueryOnlyFSMState(b *testing.B) {
	store, closeF := openRocksDBStore(b)
	defer closeF()

	srvCloseF := startMetricsServer(store)
	defer srvCloseF()

	N := 1000000
	b.N = N

	// populate storage
	for i := 0; i < b.N; i++ {
		store.Mutate([]*storage.Mutation{
			{
				Table: storage.FSMStateTable,
				Key:   storage.FSMStateTableKey,
				Value: rand.Bytes(24),
			},
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := store.Get(storage.FSMStateTable, storage.FSMStateTableKey)
		require.NoError(b, err)
	}

}

func BenchmarkMutateAllTables(b *testing.B) {
	store, closeF := openRocksDBStore(b)
	defer closeF()

	srvCloseF := startMetricsServer(store)
	defer srvCloseF()

	hasher := hashing.NewFakeSha256Hasher()
	b.N = 1000000
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := hasher.Do([]byte(fmt.Sprintf("test%d", i)))
		index := util.Uint64AsBytes(uint64(i))
		historyKey := append(index, []byte{0x0, 0x0}...)
		hyperKey := []byte{0x0, 0x0}
		hyperKey = append(hyperKey, util.Uint64AsBytes(uint64(rnd.Intn(1000)))...)
		store.Mutate([]*storage.Mutation{
			{
				Table: storage.HyperTable,
				Key:   hyperKey,
				Value: rand.Bytes(1024),
			},
			{
				Table: storage.HistoryTable,
				Key:   historyKey,
				Value: key,
			},
			{
				Table: storage.FSMStateTable,
				Key:   storage.FSMStateTableKey,
				Value: rand.Bytes(24),
			},
		})
	}

}

func startMetricsServer(store *RocksDBStore) func() {
	reg := prometheus.NewRegistry()
	store.RegisterMetrics(reg)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	srv := &http.Server{Addr: ":2112", Handler: mux}
	go srv.ListenAndServe()
	closeF := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}
	return closeF
}
