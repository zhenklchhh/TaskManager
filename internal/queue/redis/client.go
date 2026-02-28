package redis

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TaskQueue interface {
	PublishTask(context.Context, uuid.UUID) error
	PopTask(context.Context) (uuid.UUID, error)
}

type RedisClient struct {
	Client *redis.Client
}

const (
	taskQueueName = "task:scheduled"
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
// Popping task there is a chance to get error. implement rollback to redis task
func (rdc *RedisClient) PopTask(ctx context.Context) (uuid.UUID, error) {
	fields, err := rdc.Client.BLPop(ctx, 1*time.Second, taskQueueName).Result()
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(fields[1])
}
