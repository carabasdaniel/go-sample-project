package metrics

import (
	"github.com/slok/go-http-metrics/metrics"
	"github.com/slok/go-http-metrics/metrics/prometheus"
)

func NewPrometheusRecorder() metrics.Recorder {
	return prometheus.NewRecorder(prometheus.Config{})
}
