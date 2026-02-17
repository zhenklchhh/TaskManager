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
	client *redis.Client
}

const (
	taskQueueName = "task:scheduled"
)

func NewRedisClient(addr, password string, protocol, db int) *RedisClient {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			Protocol: protocol,
			DB:       db,
		}),
	}
}

func (rdc *RedisClient) PublishTask(ctx context.Context, taskID string) error {
	if err := rdc.client.RPush(ctx, taskQueueName, taskID).Err(); err != nil {
		return err
	}
	return nil
}

func (rdc *RedisClient) PopTask(ctx context.Context) (string,error) {
	result, err := rdc.client.BLPop(ctx, 1 * time.Second, taskQueueName).Result()
	if err != nil {
		return "", err
	}
	return result[1], nil
}
