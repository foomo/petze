package exporter

import (
	"time"

	"github.com/foomo/petze/watch"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	serviceErrors = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "petze_service_error_count",
		Help: "Number of services that are throwing errors for their scenarios",
	}, []string{"service_id"})

	hostErrors = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "petze_host_error_count",
		Help: "Number of hosts that are throwing errors for their scenarios",
	}, []string{"host_id"})

	serviceResponseTimes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "petze_service_session_execution_time",
		Help: "Service response times per session execution",
	}, []string{"service_id"})

	hostResponseTimes = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "petze_host_ping_execution_time",
		Help: "Host response times per ICMP echo reply",
	}, []string{"host_id"})
)

func init() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(serviceErrors)
	prometheus.MustRegister(serviceResponseTimes)
	prometheus.MustRegister(hostErrors)
	prometheus.MustRegister(hostResponseTimes)
}

func PrometheusServiceMetricsListener(serviceResult watch.ServiceResult) {
	serviceErrors.WithLabelValues(serviceResult.ID).Set(float64(len(serviceResult.Errors)))
	serviceResponseTimes.WithLabelValues(serviceResult.ID).Set(float64(serviceResult.RunTime / time.Millisecond))
}

func PrometheusHostMetricsListener(hostResult watch.HostResult) {
	hostErrors.WithLabelValues(hostResult.ID).Set(float64(len(hostResult.Errors)))
	hostResponseTimes.WithLabelValues(hostResult.ID).Set(float64(hostResult.RunTime / time.Millisecond))
}
