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
	"net/http"
	"testing"

	"github.com/bbva/qed/storage"
	"github.com/bbva/qed/testutils/rand"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func BenchmarkMutateOnlyIndex(b *testing.B) {
	store, closeF := openRocksDBStore(b)
	defer closeF()

	reg := prometheus.NewRegistry()
	reg.MustRegister(PrometheusCollectors()...)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go http.ListenAndServe(":2112", nil)

	b.N = 10000000
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.Mutate([]*storage.Mutation{
			{
				Table: storage.IndexTable,
				Key:   rand.Bytes(32),
				Value: rand.Bytes(8),
			},
		})
	}

}

func BenchmarkMutateOnlyHyper(b *testing.B) {
	store, closeF := openRocksDBStore(b)
	defer closeF()

	reg := prometheus.NewRegistry()
	reg.MustRegister(PrometheusCollectors()...)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go http.ListenAndServe(":2112", nil)

	b.N = 10000000
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.Mutate([]*storage.Mutation{
			{
				Table: storage.HyperCacheTable,
				Key:   rand.Bytes(34),
				Value: rand.Bytes(1024),
			},
		})
	}

}

func BenchmarkMutateOnlyHistory(b *testing.B) {
	store, closeF := openRocksDBStore(b)
	defer closeF()

	reg := prometheus.NewRegistry()
	reg.MustRegister(PrometheusCollectors()...)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go http.ListenAndServe(":2112", nil)

	b.N = 10000000
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.Mutate([]*storage.Mutation{
			{
				Table: storage.HistoryCacheTable,
				Key:   rand.Bytes(34),
				Value: rand.Bytes(32),
			},
		})
	}

}

func BenchmarkMutateOnlyFSMState(b *testing.B) {
	store, closeF := openRocksDBStore(b)
	defer closeF()

	reg := prometheus.NewRegistry()
	reg.MustRegister(PrometheusCollectors()...)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go http.ListenAndServe(":2112", nil)

	b.N = 1000000
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.Mutate([]*storage.Mutation{
			{
				Table: storage.FSMStateTable,
				Key:   storage.FSMStateTableKey,
				Value: rand.Bytes(128),
			},
		})
	}

}

func BenchmarkMutateAllTables(b *testing.B) {
	store, closeF := openRocksDBStore(b)
	defer closeF()

	reg := prometheus.NewRegistry()
	reg.MustRegister(PrometheusCollectors()...)
	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	go http.ListenAndServe(":2112", nil)

	b.N = 10000000
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.Mutate([]*storage.Mutation{
			{
				Table: storage.IndexTable,
				Key:   rand.Bytes(128),
				Value: []byte("Value"),
			},
			{
				Table: storage.HyperCacheTable,
				Key:   rand.Bytes(128),
				Value: []byte("Value"),
			},
			{
				Table: storage.HistoryCacheTable,
				Key:   rand.Bytes(128),
				Value: []byte("Value"),
			},
			{
				Table: storage.FSMStateTable,
				Key:   storage.FSMStateTableKey,
				Value: rand.Bytes(128),
			},
		})
	}

}
