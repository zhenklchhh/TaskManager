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
	taskService        *service.TaskService
	taskQueue          redis.TaskQueue
	timeout            time.Duration
	staleTaskThreshold time.Duration
	done               chan struct{}
	wg                 sync.WaitGroup
}

func NewScheduler(taskService *service.TaskService, timeout time.Duration, client *redis.RedisClient,
	staleTaskThreshold time.Duration) *Scheduler {
	return &Scheduler{
		taskService:        taskService,
		timeout:            timeout,
		done:               make(chan struct{}),
		taskQueue:          client,
		wg:                 sync.WaitGroup{},
		staleTaskThreshold: staleTaskThreshold,
	}
}

func (s *Scheduler) Start() {
	t := time.NewTicker(s.timeout)
	s.wg.Add(2)
	go s.schedulePendingTasksCycle(t)
	go s.rollbackScheduleTasksCycle(t)
}

func (s *Scheduler) Stop() {
	close(s.done)
	s.wg.Wait()
}

func (s *Scheduler) schedulePendingTasksCycle(t *time.Ticker) {
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
					s.taskService.UpdateTaskStatus(context.Background(), &domain.TaskUpdateStatusCmd{
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

func (s *Scheduler) rollbackScheduleTasksCycle(t *time.Ticker) {
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
			rowsAffected, err := s.taskService.
				UpdateStaleTasksToPending(context.Background(), s.staleTaskThreshold)
			if err != nil {
				slog.Error("scheduler: failed to update stale tasks", "error", err)
				continue
			}
			if rowsAffected != 0 {
				slog.Info("scheduler: stale tasks rescheduled ", "amount", rowsAffected)
			}
		}
	}
}
