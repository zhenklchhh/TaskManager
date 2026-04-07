package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/zhenklchhh/TaskManager/internal/config"
	"github.com/zhenklchhh/TaskManager/internal/queue/redis"
	"github.com/zhenklchhh/TaskManager/internal/repository/postgres"
	"github.com/zhenklchhh/TaskManager/internal/service"
	"github.com/zhenklchhh/TaskManager/internal/worker"
	"github.com/zhenklchhh/TaskManager/logger"
	"gopkg.in/mail.v2"
)

var workerAmount = 10

func main() {
	cfg := config.MustLoad()
	log := logger.SetupLogger(cfg.Env)

	if cfg.DBConfig.Url == "" {
		log.Error("database url is required")
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DBConfig.Url)
	if err != nil {
		log.Error("failed to create pg pool", "error", err)
		os.Exit(1)
	}
	redisClient := redis.NewRedisClient(cfg.RedisConfig.Address)
	repo := postgres.NewTaskRepository(pool)
	s := service.NewTaskService(repo, cfg.DefaultTaskMaxRetries)

	notifRepo := postgres.NewNotificationRepository(pool)
	dialer := mail.NewDialer(cfg.MailHogConfig.Host, cfg.MailHogConfig.Port, cfg.MailHogConfig.Username, cfg.MailHogConfig.Password)
	notifService := service.NewNotificationService(notifRepo, dialer)

	depRepo := postgres.NewDependencyRepository(pool)
	depService := service.NewDependencyService(depRepo, repo)

	worker := worker.NewWorker(s, notifService, depService, time.Minute, redisClient, workerAmount, cfg.MailHogConfig)
	worker.Start()

	// Expose worker metrics on port 8082
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsServer := &http.Server{Addr: ":8082", Handler: metricsMux}
	go func() {
		slog.Info("worker metrics server starting", "addr", ":8082")
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("worker metrics server failed", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	worker.Stop()
	metricsServer.Close()
}
