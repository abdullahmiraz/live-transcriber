package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics groups the Prometheus collectors used across the app. Construct once and
// inject where needed so collectors are registered exactly once.
type Metrics struct {
	Registry *prometheus.Registry

	MeetingsActive       prometheus.Gauge
	WSConnectionsActive  prometheus.Gauge
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	ErrorsTotal          *prometheus.CounterVec
	TranscriptionLatency *prometheus.HistogramVec
	ChatMessagesTotal    prometheus.Counter
}

// NewMetrics creates and registers all collectors on a fresh registry.
func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	factory := promauto.With(reg)

	// Standard process/go collectors for baseline visibility.
	reg.MustRegister(prometheus.NewGoCollector())
	reg.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	return &Metrics{
		Registry: reg,
		MeetingsActive: factory.NewGauge(prometheus.GaugeOpts{
			Name: "meetings_active",
			Help: "Number of active meeting rooms with connected participants.",
		}),
		WSConnectionsActive: factory.NewGauge(prometheus.GaugeOpts{
			Name: "ws_connections_active",
			Help: "Number of currently open WebSocket connections.",
		}),
		HTTPRequestsTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests.",
		}, []string{"route", "method", "status"}),
		HTTPRequestDuration: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		}, []string{"route", "method", "status"}),
		ErrorsTotal: factory.NewCounterVec(prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total errors by component.",
		}, []string{"component"}),
		TranscriptionLatency: factory.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "transcription_latency_seconds",
			Help:    "End-to-end transcription latency in seconds.",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10},
		}, []string{"provider"}),
		ChatMessagesTotal: factory.NewCounter(prometheus.CounterOpts{
			Name: "chat_messages_total",
			Help: "Total chat messages sent.",
		}),
	}
}
