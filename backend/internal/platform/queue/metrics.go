package queue

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	publishedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bitik_queue_published_total",
			Help: "Total published queue messages by job type.",
		},
		[]string{"job_type"},
	)
	consumedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bitik_queue_consumed_total",
			Help: "Total consumed queue messages by job type.",
		},
		[]string{"job_type"},
	)
	ackedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bitik_queue_acked_total",
			Help: "Total acked queue messages by job type.",
		},
		[]string{"job_type"},
	)
	retriedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bitik_queue_retried_total",
			Help: "Total retried queue messages by job type.",
		},
		[]string{"job_type"},
	)
	dlqTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bitik_queue_dlq_total",
			Help: "Total queue messages sent to dead-letter by job type.",
		},
		[]string{"job_type"},
	)
	nackedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bitik_queue_nacked_total",
			Help: "Total nacked queue messages by job type.",
		},
		[]string{"job_type"},
	)
	handlerFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bitik_queue_handler_failures_total",
			Help: "Total handler failures by job type.",
		},
		[]string{"job_type"},
	)
	handlerDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "bitik_queue_handler_duration_seconds",
			Help:    "Handler runtime by job type.",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"job_type"},
	)
	inFlight = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bitik_queue_in_flight_workers",
			Help: "In-flight worker handlers by job type.",
		},
		[]string{"job_type"},
	)
)

func init() {
	prometheus.MustRegister(
		publishedTotal, consumedTotal, ackedTotal, retriedTotal, dlqTotal, nackedTotal,
		handlerFailuresTotal, handlerDuration, inFlight,
	)
}

func observeHandler(jobType JobType, started time.Time, err error) {
	handlerDuration.WithLabelValues(string(jobType)).Observe(time.Since(started).Seconds())
	if err != nil {
		handlerFailuresTotal.WithLabelValues(string(jobType)).Inc()
	}
}
