package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Task metrics
	tasksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tasks_total",
			Help: "Total number of tasks",
		},
		[]string{"type", "status"},
	)

	taskProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "task_processing_duration_seconds",
			Help:    "Task processing duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type", "status"},
	)

	// Database metrics
	dbConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Redis metrics
	redisConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "redis_connections_active",
			Help: "Number of active Redis connections",
		},
	)

	// Scheduler metrics
	schedulerRunsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "scheduler_runs_total",
			Help: "Total number of scheduler runs",
		},
	)

	schedulerErrorsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "scheduler_errors_total",
			Help: "Total number of scheduler errors",
		},
	)

	// Worker metrics
	workersActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "workers_active",
			Help: "Number of active workers",
		},
	)

	workerTasksProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "worker_tasks_processed_total",
			Help: "Total number of tasks processed by workers",
		},
		[]string{"worker_id", "status"},
	)

	// Priority queue metrics
	queueLengthByPriority = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_length_by_priority",
			Help: "Number of tasks in queue by priority",
		},
		[]string{"priority"},
	)

	tasksProcessedByPriority = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tasks_processed_by_priority_total",
			Help: "Total number of tasks processed by priority",
		},
		[]string{"priority", "status"},
	)

	// Semaphore metrics
	semaphoreAcquired = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "semaphore_acquired_total",
			Help: "Total number of semaphore acquisitions",
		},
	)

	semaphoreReleased = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "semaphore_released_total",
			Help: "Total number of semaphore releases",
		},
	)

	semaphoreSlots = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "semaphore_slots_available",
			Help: "Number of available semaphore slots",
		},
	)
)

// HTTP metrics functions
func RecordHTTPRequest(method, endpoint, status string) {
	httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
}

func RecordHTTPRequestDuration(method, endpoint string, duration float64) {
	httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// Task metrics functions
func RecordTaskCreated(taskType, status string) {
	tasksTotal.WithLabelValues(taskType, status).Inc()
}

func RecordTaskProcessingDuration(taskType, status string, duration float64) {
	taskProcessingDuration.WithLabelValues(taskType, status).Observe(duration)
}

// Database metrics functions
func SetDBConnectionsActive(count float64) {
	dbConnectionsActive.Set(count)
}

func RecordDBQueryDuration(operation string, duration float64) {
	dbQueryDuration.WithLabelValues(operation).Observe(duration)
}

// Redis metrics functions
func SetRedisConnectionsActive(count float64) {
	redisConnectionsActive.Set(count)
}

// Scheduler metrics functions
func RecordSchedulerRun() {
	schedulerRunsTotal.Inc()
}

func RecordSchedulerError() {
	schedulerErrorsTotal.Inc()
}

// Worker metrics functions
func SetWorkersActive(count float64) {
	workersActive.Set(count)
}

func RecordWorkerTaskProcessed(workerID, status string) {
	workerTasksProcessed.WithLabelValues(workerID, status).Inc()
}

// Priority queue metrics functions
func SetQueueLengthByPriority(priority string, length float64) {
	queueLengthByPriority.WithLabelValues(priority).Set(length)
}

func RecordTaskProcessedByPriority(priority, status string) {
	tasksProcessedByPriority.WithLabelValues(priority, status).Inc()
}

// Semaphore metrics functions
func RecordSemaphoreAcquired() {
	semaphoreAcquired.Inc()
}

func RecordSemaphoreReleased() {
	semaphoreReleased.Inc()
}

func SetSemaphoreSlotsAvailable(slots float64) {
	semaphoreSlots.Set(slots)
}
