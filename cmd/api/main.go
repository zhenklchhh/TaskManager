package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/zhenklchhh/TaskManager/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	// TODO: delete print
	fmt.Println(cfg)

	log := setupLogger(cfg.Env)
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// dsn := os.Getenv("DATABASE_URL")
	// if dsn == "" {
	// 	log.Fatal("DATABASE_URL is required")
	// }
	// addr := os.Getenv("PORT")
	// if addr == "" {
	// 	addr = "8080"
	// }
	// addr = ":" + addr

	// ctx := context.Background()
	// pool, err := pgxpool.New(ctx,dsn)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// repo := postgres.NewTaskRepository(pool)
	// s := service.NewTaskService(repo)
	// h := httpTransport.NewHandler(s)
	// r := httpTransport.Routes(h)
	// server := &http.Server{
	// 	Addr: addr,
	// 	Handler: r,
	// 	ReadHeaderTimeout: 5 * time.Second,
	// 	ReadTimeout: 15 * time.Second,
	// 	WriteTimeout: 15 * time.Second,
	// 	IdleTimeout: 60 * time.Second,
	// }

	// go func() {
	// 	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 		log.Fatal(err)
	// 	}
	// }()

	// stop := make(chan os.Signal, 1)
	// signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	// <-stop
	// shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	// if err := server.Shutdown(shutdownCtx); err != nil {
	// 	log.Println("shutdown error:", err)
	// }
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case envLocal:
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		logger = slog.New(
					slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return logger
}
