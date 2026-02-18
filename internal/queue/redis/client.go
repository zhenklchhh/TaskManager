package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type TaskQueue interface {
	PublishTask(context.Context, string) error
	PopTask(context.Context) (string, error)
}

type RedisClient struct {
	Client *redis.Client
}

const (
	taskQueueName = "task:scheduled"
)

func NewRedisClient(addr string) *RedisClient {
	return &RedisClient {
		Client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

func (rdc *RedisClient) PublishTask(ctx context.Context, taskID string) error {
	if err := rdc.Client.RPush(ctx, taskQueueName, taskID).Err(); err != nil {
		return err
	}
	return nil
}

func (rdc *RedisClient) PopTask(ctx context.Context) (string,error) {
	result, err := rdc.Client.BLPop(ctx, 1 * time.Second, taskQueueName).Result()
	if err != nil {
		return "", err
	}
	return result[1], nil
}
