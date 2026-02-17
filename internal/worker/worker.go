package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/queue/redis"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

type Worker struct {
	taskService *service.TaskService
	taskQueue redis.TaskQueue
	timeout     time.Duration
	done        chan bool
}

func NewWorker(taskService *service.TaskService, timeout time.Duration, client *redis.RedisClient) *Worker {
	return &Worker{
		taskService: taskService,
		timeout:     timeout,
		done:        make(chan bool),
		taskQueue: client,
	}
}

func (w *Worker) Start() {
	t := time.NewTicker(w.timeout)
	go w.workerCmd(t)
}

func (w *Worker) Stop() {
	w.done <- true
}

func (w *Worker) workerCmd(t *time.Ticker) {
	for {
		select {
		case <-w.done:
			t.Stop()
			return
		case <-t.C:
			id, err := w.taskQueue.PopTask(context.Background())
			if err != nil {
				fmt.Printf("worker: error popping task: %v", err)
			}
			task, err := w.taskService.GetTaskById(context.Background(), id)
			w.executeTask(task)
		}
	}
}

func (w *Worker) executeTask(t *domain.Task) {
	fmt.Println("worker: executing task...")
	time.Sleep(500 * time.Millisecond)
}
