package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthChecker struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

func NewHealthChecker(db *pgxpool.Pool, redis *redis.Client) *HealthChecker {
	return &HealthChecker{
		db:    db,
		redis: redis,
	}
}

func (h *HealthChecker) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	status := HealthStatus{
		Timestamp: time.Now(),
		Services:  make(map[string]string),
	}

	allHealthy := true

	// Check database
	if err := h.checkDatabase(ctx); err != nil {
		status.Services["database"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		status.Services["database"] = "healthy"
	}

	// Check Redis
	if err := h.checkRedis(ctx); err != nil {
		status.Services["redis"] = "unhealthy: " + err.Error()
		allHealthy = false
	} else {
		status.Services["redis"] = "healthy"
	}

	if allHealthy {
		status.Status = "healthy"
		w.WriteHeader(http.StatusOK)
	} else {
		status.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *HealthChecker) checkDatabase(ctx context.Context) error {
	if h.db == nil {
		return nil // Skip if no database connection
	}
	
	return h.db.Ping(ctx)
}

func (h *HealthChecker) checkRedis(ctx context.Context) error {
	if h.redis == nil {
		return nil // Skip if no Redis connection
	}
	
	return h.redis.Ping(ctx).Err()
}

func (h *HealthChecker) Ready(w http.ResponseWriter, r *http.Request) {
	// Readiness check - verify all critical dependencies are ready
	h.Check(w, r)
}
