package scheduler

import (
	"log"
	"os"
	"strconv"
	"github.com/joho/godotenv"
	"github.com/zhenklchhh/TaskManager/internal/domain"
	"github.com/zhenklchhh/TaskManager/internal/service"
	"time"
)

type Scheduler struct {
	scheduledTasks []domain.Task
	taskService *service.TaskService
}

func (s Scheduler) StartScheduler() {
	timeout := getTimeoutFromEnv()
	ticker := time.NewTicker(time.Duration(timeout) * time.Millisecond)
	go func() {
		
	}
}

func (s Scheduler) getScheduledTasks() {
}

func getTimeoutFromEnv() (int) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
		return -1
	}
	timeout, err := strconv.Atoi(os.Getenv("SCHEDULER_TIMEOUT"))
	if err != nil {
		log.Fatal(err)
		return -1
	}
	return timeout
}