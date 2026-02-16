package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/queue/redis"
	"github.com/zhenklchhh/TaskManager/internal/service"
)

type Scheduler struct {
	taskService *service.TaskService
	redisClient *redis.RedisClient
	timeout      time.Duration
	done        chan bool
}

func NewScheduler(taskService *service.TaskService, timeout time.Duration, client *redis.RedisClient) *Scheduler {
	return &Scheduler{
		taskService: taskService,
		timeout:      timeout,
		done:        make(chan bool),
		redisClient: client,
	}
}

func (s *Scheduler) Start() error {
	t := time.NewTicker(s.timeout)
	go s.scheduleCmd(t)
}

func (s *Scheduler) Stop() {
	s.done <- true
}

func (s *Scheduler) scheduleCmd(t *time.Ticker) {
	for {
		select {
		case <-s.done:
			t.Stop()
			return
		case <-t.C:
			tasks := s.checkForUpcomingTasks(context.Background())

		}
	}
}

func (s *Scheduler) checkForUpcomingTasks(ctx context.Context) ([]int) {
	tasks, err := s.taskService.GetScheduledTasks(ctx)
	if err != nil {
		log.Printf("scheduler: error while checking upcoming tasks: %s", err)
		return nil
	}
	return tasks
}
