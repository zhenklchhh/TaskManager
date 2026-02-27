package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	rdc "github.com/zhenklchhh/TaskManager/internal/queue/redis"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

type Worker struct {
	taskService *service.TaskService
	taskQueue   rdc.TaskQueue
	timeout     time.Duration
	done        chan struct{}
	queuedTasks chan uuid.UUID
	sleep       chan struct{}
	workers     int
	wg          sync.WaitGroup
}

func NewWorker(taskService *service.TaskService, timeout time.Duration, client *rdc.RedisClient,
	workerAmount int) *Worker {
	return &Worker{
		taskService: taskService,
		timeout:     timeout,
		done:        make(chan struct{}),
		queuedTasks: make(chan uuid.UUID),
		sleep:       make(chan struct{}),
		taskQueue:   client,
		workers:     workerAmount,
	}
}

func (w *Worker) Start() {
	w.wg.Add(1)
	go w.pullTasksFromRedis()
	for i := 0; i < w.workers; i++ {
		w.wg.Add(1)
		go w.workerCmd()
	}
}

// start() -> 1) pull tasks from redis with timeout when no tasks in queue -> after that start fixed amount of workers
// with channel of tasks

func (w *Worker) Stop() {
	close(w.done)
	w.wg.Wait()
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
			taskUpdateCmd := &service.TaskUpdateStatusCmd{
				ID: id,
			}
			task, err := w.taskService.GetTaskById(context.Background(), id)
			if err != nil {
				slog.Error("worker: failed to get task by id", "task_id", id, "error", err)
				continue
			}
			slog.Info("worker: picked up task", "id", id)
			taskUpdateCmd.Status = domain.TaskStatusRunning
			if w.taskService.UpdateTaskStatus(context.Background(), taskUpdateCmd); err != nil {
				slog.Error("worker: failed to update task status", "error", err)
			}
			err = w.executeTask(task)
			if err != nil {
				slog.Error("worker: failed to complete task", "error", err)
				taskUpdateCmd.Status = domain.TaskStatusFailed
			} else {
				taskUpdateCmd.Status = domain.TaskStatusCompleted
			}
			if w.taskService.UpdateTaskStatus(context.Background(), taskUpdateCmd); err != nil {
				slog.Error("worker: failed to update task status", "error", err)
			}
		}
	}
}

// todo: retries
func (w *Worker) executeTask(t *domain.Task) error {
	slog.Info("worker: executing task", "id", t.ID.String())
	time.Sleep(500 * time.Millisecond)
	return nil
}
