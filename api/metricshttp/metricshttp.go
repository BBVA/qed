package metricshttp

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewMetricsHttp() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	return mux
}
