package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/foomo/petze/watch"
	"time"
)

var (
	serviceErrors = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "petze_service_error_count",
		Help: "Number of services that are throwing errors for their scenarios",
	}, []string{"service_id"})

	serviceResponseTimes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "petze_service_session_execution_time",
		Help: "Service response times per session execution",
	}, []string{"service_id"})
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(serviceErrors)
	prometheus.MustRegister(serviceResponseTimes)
}

func PrometheusMetricsListener(result watch.Result) {
	serviceErrors.WithLabelValues(result.ID).Set(float64(len(result.Errors)))
	serviceResponseTimes.WithLabelValues(result.ID).Set(float64(result.RunTime / time.Millisecond))
}
