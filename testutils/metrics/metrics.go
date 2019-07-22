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

package metrics

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/bbva/qed/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StartMetricsServer(registerers ...metrics.Registerer) func() {
	reg := prometheus.NewRegistry()
	for _, r := range registerers {
		r.RegisterMetrics(reg)
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	srv := &http.Server{Addr: ":2112", Handler: mux}
	go srv.ListenAndServe()

	closeF := func() {
		if srv == nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}
	return closeF
}

type customMetrics struct {
	registerMetrics func(metrics.Registry)
}

func (cm customMetrics) RegisterMetrics(r metrics.Registry) {
	cm.registerMetrics(r)
}

func CustomRegister(c ...prometheus.Collector) metrics.Registerer {
	m := struct{ customMetrics }{}
	m.registerMetrics = func(r metrics.Registry) {
		r.MustRegister(c...)
	}
	return m
}
