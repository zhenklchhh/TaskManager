package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/zhenklchhh/TaskManager/internal/config"
	"github.com/zhenklchhh/TaskManager/internal/queue/redis"
	"github.com/zhenklchhh/TaskManager/internal/repository/postgres"
	"github.com/zhenklchhh/TaskManager/internal/service"
	"github.com/zhenklchhh/TaskManager/internal/worker"
	"github.com/zhenklchhh/TaskManager/logger"
)

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
		log.Error("error: ", err)
		os.Exit(1)
	}
	redisClient := redis.NewRedisClient(cfg.RedisConfig.Address)
	repo := postgres.NewTaskRepository(pool)
	s := service.NewTaskService(repo)
	worker := worker.NewWorker(s, time.Minute, redisClient)
	worker.Start()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	worker.Stop()
}
