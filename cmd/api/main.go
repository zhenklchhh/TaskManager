package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	api "github.com/zhenklchhh/TaskManager/internal/api"
	"github.com/zhenklchhh/TaskManager/internal/config"
	"github.com/zhenklchhh/TaskManager/internal/repository/postgres"
	"github.com/zhenklchhh/TaskManager/internal/service"
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
	repo := postgres.NewTaskRepository(pool)
	s := service.NewTaskService(repo)
	h := api.NewHandler(s)
	r := api.Routes(h)
	server := &http.Server{
		Addr:              cfg.Server.Address,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       cfg.Server.Timeout,
		WriteTimeout:      cfg.Server.Timeout,
		IdleTimeout:       cfg.Server.IddleTimeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("error: ", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown error:", err)
	}
}
