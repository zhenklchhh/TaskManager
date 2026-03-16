package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TaskQueue interface {
	PublishTask(context.Context, uuid.UUID) error
	PublishTaskWithPriority(context.Context, uuid.UUID, int) error
	PopTask(context.Context) (uuid.UUID, error)
}

type RedisClient struct {
	Client *redis.Client
}

const (
	taskQueueName         = "task:scheduled"
	taskPriorityQueueName = "task:priority:queue"
)

func NewRedisClient(addr string) *RedisClient {
	return &RedisClient{
		Client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (rdc *RedisClient) PublishTask(ctx context.Context, taskID uuid.UUID) error {
	if err := rdc.Client.RPush(ctx, taskQueueName, taskID.String()).Err(); err != nil {
		return err
	}
	return nil
}

func (rdc *RedisClient) PublishTaskWithPriority(ctx context.Context, taskID uuid.UUID, priority int) error {
	score := float64(priority)
	if err := rdc.Client.ZAdd(ctx, taskPriorityQueueName, redis.Z{
		Score:  score,
		Member: taskID.String(),
	}).Err(); err != nil {
		return err
	}
	return nil
}

func (rdc *RedisClient) PopTask(ctx context.Context) (uuid.UUID, error) {
	result, err := rdc.Client.ZPopMin(ctx, taskPriorityQueueName, 1).Result()
	if err == nil && len(result) > 0 {
		taskIDStr, ok := result[0].Member.(string)
		if ok {
			return uuid.Parse(taskIDStr)
		}
	}
	
	if err != nil && err != redis.Nil {
		return uuid.Nil, err
	}
	
	fields, err := rdc.Client.BLPop(ctx, 1*time.Second, taskQueueName).Result()
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(fields[1])
}

func (rdc *RedisClient) GetQueueLength(ctx context.Context) (int64, error) {
	priorityCount, err := rdc.Client.ZCard(ctx, taskPriorityQueueName).Result()
	if err != nil {
		return 0, err
	}
	
	regularCount, err := rdc.Client.LLen(ctx, taskQueueName).Result()
	if err != nil {
		return 0, err
	}
	
	return priorityCount + regularCount, nil
}

func (rdc *RedisClient) RemoveTask(ctx context.Context, taskID uuid.UUID) error {
	taskIDStr := taskID.String()
	
	if err := rdc.Client.ZRem(ctx, taskPriorityQueueName, taskIDStr).Err(); err != nil {
		return fmt.Errorf("failed to remove from priority queue: %w", err)
	}
	
	if err := rdc.Client.LRem(ctx, taskQueueName, 0, taskIDStr).Err(); err != nil {
		return fmt.Errorf("failed to remove from regular queue: %w", err)
	}
	
	return nil
}
