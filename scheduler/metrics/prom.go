package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	JobsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "jobs_total",
		Help: "Total enqueued jobs",
	})
	JobsSucceeded = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "jobs_succeeded_total",
		Help: "Total number of jobs completed successfully",
	})
	JobsFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "jobs_failed_total",
		Help: "Total number of jobs failed",
	})
	JobsDLQ = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "jobs_dlq_total",
		Help: "Total number of jobs sent to DLQ",
	})
	JobsInProgress = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "jobs_in_progress",
		Help: "Number of jobs currently being processed",
	})
	JobsProcessingDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "jobs_histogram",
		Help:    "Histogram for jobs duration in seconds",
		Buckets: prometheus.LinearBuckets(0.5, 0.5, 10),
	})
)

func InitProm() {
	prometheus.MustRegister(JobsTotal, JobsSucceeded, JobsFailed, JobsDLQ, JobsInProgress, JobsProcessingDuration)
}
