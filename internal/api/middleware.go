package api

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zhenklchhh/TaskManager/internal/metrics"
)

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		// Process the request
		next.ServeHTTP(wrapped, r)
		
		// Record metrics
		duration := time.Since(start).Seconds()
		status := http.StatusText(wrapped.statusCode)
		
		metrics.RecordHTTPRequest(r.Method, r.URL.Path, status)
		metrics.RecordHTTPRequestDuration(r.Method, r.URL.Path, duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
