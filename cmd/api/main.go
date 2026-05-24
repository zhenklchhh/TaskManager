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
	"github.com/redis/go-redis/v9"
	api "github.com/zhenklchhh/TaskManager/internal/api"
	"github.com/zhenklchhh/TaskManager/internal/config"
	appRedis "github.com/zhenklchhh/TaskManager/internal/queue/redis"
	"github.com/zhenklchhh/TaskManager/internal/repository/postgres"
	"github.com/zhenklchhh/TaskManager/internal/scheduler"
	"github.com/zhenklchhh/TaskManager/internal/service"
	"github.com/zhenklchhh/TaskManager/logger"
	"gopkg.in/mail.v2"
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
		log.Error("failed to create pg pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisConfig.Address,
	})
	
	// Task repositories and services
	repo := postgres.NewTaskRepository(pool)
	execRepo := postgres.NewExecutionRepository(pool)
	s := service.NewTaskService(repo, cfg.DefaultTaskMaxRetries)
	s.SetExecutionRepository(execRepo)
	h := api.NewHandler(s)
	dashboardHandler := api.NewDashboardHandler(s)
	batchHandler := api.NewBatchHandler(s)

	// Notification
	notifRepo := postgres.NewNotificationRepository(pool)
	dialer := mail.NewDialer(cfg.MailHogConfig.Host, cfg.MailHogConfig.Port, cfg.MailHogConfig.Username, cfg.MailHogConfig.Password)
	notifService := service.NewNotificationService(notifRepo, dialer)
	notifHandler := api.NewNotificationHandler(notifService)

	// Dependencies
	depRepo := postgres.NewDependencyRepository(pool)
	depService := service.NewDependencyService(depRepo, repo)
	depHandler := api.NewDependencyHandler(depService)

	// Auth, Company, Group
	userRepo := postgres.NewUserRepository(pool)
	companyRepo := postgres.NewCompanyRepository(pool)
	groupRepo := postgres.NewGroupRepository(pool)

	authService := service.NewAuthService(userRepo, companyRepo, cfg.Auth)
	companyService := service.NewCompanyService(companyRepo, userRepo, groupRepo)
	groupService := service.NewGroupService(groupRepo)

	authHandler := api.NewAuthHandler(authService)
	companyHandler := api.NewCompanyHandler(companyService)
	groupHandler := api.NewGroupHandler(groupService)
	
	// Create health checker
	healthChecker := api.NewHealthChecker(pool, redisClient)
	
	r := api.Routes(api.RouteParams{
		Handler:          h,
		HealthChecker:    healthChecker,
		DashboardHandler: dashboardHandler,
		BatchHandler:     batchHandler,
		DepHandler:       depHandler,
		NotifHandler:     notifHandler,
		AuthHandler:      authHandler,
		CompanyHandler:   companyHandler,
		GroupHandler:     groupHandler,
		AuthService:      authService,
	})

	scheduler := scheduler.NewScheduler(s, time.Minute, &appRedis.RedisClient{
		Client: redisClient,
	}, cfg.SchedulerConfig.StaleTaskThreshold)
	scheduler.Start()
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
			log.Error("http server failed", "error", err)
			os.Exit(1)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	scheduler.Stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error:", "error", err)
	}
}
