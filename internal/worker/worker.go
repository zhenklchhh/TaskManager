package worker

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

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
	wg          sync.WaitGroup
}

func NewWorker(taskService *service.TaskService, timeout time.Duration, client *rdc.RedisClient) *Worker {
	return &Worker{
		taskService: taskService,
		timeout:     timeout,
		done:        make(chan struct{}),
		taskQueue:   client,
	}
}

func (w *Worker) Start() {
	t := time.NewTicker(w.timeout)
	w.wg.Add(1)
	go w.workerCmd(t)
}

func (w *Worker) Stop() {
	close(w.done)
	w.wg.Wait()
}

func (w *Worker) workerCmd(t *time.Ticker) {
	defer w.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("Worker panicked and recovered", "error", r)
		}
	}()
	for {
		select {
		case <-w.done:
			t.Stop()
			return
		case <-t.C:
			id, err := w.taskQueue.PopTask(context.Background())
			if err != nil{
				if !errors.Is(err, redis.Nil) {
					slog.Error("worker: failed to pop task", "error", err)
				}
				continue
			}
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
			} else{
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
	slog.Info("worker: executing task", t.ID)
	time.Sleep(500 * time.Millisecond)
	return nil
}
