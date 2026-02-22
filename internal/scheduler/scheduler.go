package scheduler

import (
	"context"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
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
	s.wg.Wait()
	s.wg.Add(1)
	go s.scheduleCmd(t)
}

func (s *Scheduler) Stop() {
	s.done <- true

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
			tasks := s.checkForUpcomingTasks(context.Background())
			for _, taskID := range tasks {
				if err := s.taskQueue.PublishTask(context.Background(), taskID.String()); err != nil {
					log.Printf("scheduler error: %v", err)
				}
				cmd := &service.TaskUpdateStatusCmd{
					ID:     taskID.String(),
					Status: domain.TaskStatusScheduled,
				}
				if err := s.taskService.UpdateTaskStatus(context.Background(), cmd); err != nil {
					log.Printf("scheduler error: %v", err)
				}
			}
		}
	}
}

func (s *Scheduler) checkForUpcomingTasks(ctx context.Context) []uuid.UUID {
	tasks, err := s.taskService.GetPendingTasks(ctx)
	if err != nil {
		log.Printf("scheduler: error while checking upcoming tasks: %s", err)
		return nil
	}
	return tasks
}
