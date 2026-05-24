package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/zhenklchhh/TaskManager/internal/config"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/metrics"
	rdc "github.com/zhenklchhh/TaskManager/internal/queue/redis"
	"github.com/zhenklchhh/TaskManager/internal/service"
	task "github.com/zhenklchhh/TaskManager/internal/task"
)

type Worker struct {
	id                  string
	taskService         *service.TaskService
	notificationService *service.NotificationService
	dependencyService   *service.DependencyService
	taskQueue           rdc.TaskQueue
	timeout             time.Duration
	done                chan struct{}
	queuedTasks         chan uuid.UUID
	sleep               chan struct{}
	workers             int
	wg                  sync.WaitGroup
	taskHandlers        map[string]task.TaskHandler
}

func NewWorker(taskService *service.TaskService, notificationService *service.NotificationService,
	dependencyService *service.DependencyService, timeout time.Duration, client *rdc.RedisClient,
	workerAmount int, cfg config.MailHogConfig) *Worker {
	return &Worker{
		id:                  uuid.New().String(),
		taskService:         taskService,
		notificationService: notificationService,
		dependencyService:   dependencyService,
		timeout:             timeout,
		done:                make(chan struct{}),
		queuedTasks:         make(chan uuid.UUID),
		sleep:               make(chan struct{}),
		taskQueue:           client,
		workers:             workerAmount,
		taskHandlers:        initTaskHandlers(cfg),
	}
}

func initTaskHandlers(cfg config.MailHogConfig) map[string]task.TaskHandler {
	emailTaskHandler := task.NewEmailTaskHandler(
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
	)
	webhookTaskHandler := task.NewWebhookTaskHandler()
	grpcTaskHandler := task.NewGrpcTaskHandler()
	httpTaskHandler := task.NewHttpHandler()
	return map[string]task.TaskHandler{
		task.SendEmailTask:   emailTaskHandler,
		task.SendWebhookTask: webhookTaskHandler,
		task.GRPCTask:        grpcTaskHandler,
		task.HttpTask:        httpTaskHandler,
	}
}

func (w *Worker) Start() {
	metrics.SetWorkersActive(float64(w.workers))
	w.wg.Add(1)
	go w.pullTasksFromRedis()
	for i := 0; i < w.workers; i++ {
		w.wg.Add(1)
		go w.workerCmd()
	}
}

func (w *Worker) Stop() {
	close(w.done)
	metrics.SetWorkersActive(0)

	done := make(chan struct{})
	go func() {
		w.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		slog.Info("worker: graceful shutdown")
	case <-time.After(30 * time.Second):
		slog.Error("worker: forced shutdown timeout exceeded")
	}
}

func (w *Worker) pullTasksFromRedis() {
	defer w.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Worker panicked and recovered", "error", r)
		}
	}()
	for {
		select {
		case <-w.done:
			return
		case <-w.sleep:
			time.Sleep(w.timeout)
		default:
			id, err := w.taskQueue.PopTask(context.Background())
			if err != nil {
				if !errors.Is(err, redis.Nil) {
					slog.Error("worker: failed to pop task", "error", err)
				}
				continue
			}
			w.queuedTasks <- id
		}
	}
}

func (w *Worker) workerCmd() {
	defer w.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Worker panicked and recovered", "error", r)
		}
	}()
	for {
		select {
		case <-w.done:
			return
		case id := <-w.queuedTasks:

			taskUpdateCmd := &domain.TaskUpdateStatusCmd{
				ID: id,
			}
			task, err := w.taskService.GetTaskById(context.Background(), id)
			if err != nil {
				slog.Error("worker: failed to get task by id", "task_id", id, "error", err)
				continue
			}
			slog.Info("worker: picked up task", "id", id)

			if task.ExpiresAt != nil && task.ExpiresAt.Before(time.Now()) {
				slog.Warn("worker: task expired, marking as failed", "id", id, "expired_at", task.ExpiresAt)
				oldStatus := task.Status
				taskUpdateCmd.Status = domain.TaskStatusFailed
				if err := w.taskService.UpdateTaskStatus(context.Background(), taskUpdateCmd); err != nil {
					slog.Error("worker: failed to update expired task status", "error", err)
				}
				w.notify(task, oldStatus, domain.TaskStatusFailed)
				continue
			}

			oldStatus := task.Status

			if _, supported := w.taskHandlers[task.Type]; !supported {
				slog.Error("worker: unsupported task type, failing permanently", "type", task.Type, "id", id)
				taskUpdateCmd.Status = domain.TaskStatusFailed
				_ = w.taskService.UpdateTaskStatus(context.Background(), taskUpdateCmd)
				w.notify(task, oldStatus, domain.TaskStatusFailed)
				continue
			}

			taskUpdateCmd.Status = domain.TaskStatusRunning
			if err = w.taskService.UpdateTaskStatus(context.Background(), taskUpdateCmd); err != nil {
				slog.Error("worker: failed to update task status", "error", err)
			}

			execRun, _ := w.taskService.CreateExecution(context.Background(), id, w.id)

			startTime := time.Now()
			execErr := w.executeTask(context.Background(), task)
			duration := time.Since(startTime)
			durationSec := duration.Seconds()
			priorityStr := fmt.Sprintf("%d", task.Priority)
			if execErr != nil {
				slog.Error("worker: failed to complete task", "error", execErr)
				metrics.RecordTaskProcessingDuration(task.Type, "failed", durationSec)
				metrics.RecordWorkerTaskProcessed(fmt.Sprintf("%d", id), "failed")
				metrics.RecordTaskProcessedByPriority(priorityStr, "failed")
				if execRun != nil {
					w.taskService.FinishExecution(context.Background(), execRun.ID, domain.TaskStatusFailed, execErr.Error(), "", duration.Milliseconds())
				}
				taskUpdateCmd.Status = domain.TaskStatusScheduled
				w.taskService.RetryTask(context.Background(), id, execErr)
				w.notify(task, oldStatus, domain.TaskStatusFailed)
				continue
			}
			metrics.RecordTaskProcessingDuration(task.Type, "completed", durationSec)
			metrics.RecordWorkerTaskProcessed(fmt.Sprintf("%d", id), "completed")
			metrics.RecordTaskProcessedByPriority(priorityStr, "completed")
			if execRun != nil {
				w.taskService.FinishExecution(context.Background(), execRun.ID, domain.TaskStatusCompleted, "", "", duration.Milliseconds())
			}
			if err = w.taskService.MarkCompleted(context.Background(), id); err != nil {
				slog.Error("worker: failed to mark task completed", "error", err)
			}
			w.notify(task, oldStatus, domain.TaskStatusCompleted)
		}
	}
}

func (w *Worker) notify(task *domain.Task, oldStatus, newStatus domain.TaskStatus) {
	if w.notificationService != nil {
		w.notificationService.OnTaskStatusChanged(context.Background(), task, oldStatus, newStatus)
	}
	if w.dependencyService != nil && (newStatus == domain.TaskStatusCompleted || newStatus == domain.TaskStatusFailed) {
		w.dependencyService.OnTaskCompleted(context.Background(), task.ID, newStatus)
	}
}

func (w *Worker) executeTask(ctx context.Context, t *domain.Task) error {
	slog.Info("worker: executing task", "type", t.Type, "title", t.Title)
	h, ok := w.taskHandlers[t.Type]
	if !ok {
		slog.Error("worker: failed to execute task: unsupported task type")
		return fmt.Errorf("unsupported task type: %s", t.Type)
	}
	return h.Handle(ctx, t)
}
