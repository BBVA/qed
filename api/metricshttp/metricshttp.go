package metricshttp

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewMetricsHTTP(r *prometheus.Registry) *http.ServeMux {
	mux := http.NewServeMux()
	g := prometheus.Gatherers{
		prometheus.DefaultGatherer,
		r,
	}
	mux.Handle(
		"/metrics",
		promhttp.InstrumentMetricHandler(r, promhttp.HandlerFor(g, promhttp.HandlerOpts{})),
	)
	return mux
}
