package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/queue/redis"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

type Scheduler struct {
	taskService *service.TaskService
	taskQueue   redis.TaskQueue
	timeout     time.Duration
	done        chan bool
	wg          sync.WaitGroup
}

func NewScheduler(taskService *service.TaskService, timeout time.Duration, client *redis.RedisClient) *Scheduler {
	return &Scheduler{
		taskService: taskService,
		timeout:     timeout,
		done:        make(chan bool),
		taskQueue:   client,
		wg:          sync.WaitGroup{},
	}
}

func (s *Scheduler) Start() {
	t := time.NewTicker(s.timeout)
	s.wg.Add(1)
	go s.scheduleCmd(t)
}

func (s *Scheduler) Stop() {
	s.done <- true
	s.wg.Wait()
}

func (s *Scheduler) scheduleCmd(t *time.Ticker) {
	defer s.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("scheduler panicked and recovered", "error", r)
		}
	}()
	for {
		select {
		case <-s.done:
			t.Stop()
			return
		case <-t.C:
			tasks, err := s.taskService.ProcessPendingTasks(context.Background(), 50)
			for _, taskID := range tasks {
				if err := s.taskQueue.PublishTask(context.Background(), taskID); err != nil {
					slog.Error("scheduler: error scheduling tasks", "error", err)
					s.taskService.UpdateTaskStatus(context.Background(), &service.TaskUpdateStatusCmd{
						ID:     taskID,
						Status: domain.TaskStatusPending,
					})
				}
			}
			if err != nil {
				slog.Error("scheduler transaction failed", "error", err)
			}
		}
	}
}
