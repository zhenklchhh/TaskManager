package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

type TaskQueue interface {
	PublishTask(context.Context, string) error
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
	cmd := rdc.client.LPush(ctx, taskQueueName, taskID)
	_, err := cmd.Result()
	if err != nil {
		return err
	}
	return nil
}
