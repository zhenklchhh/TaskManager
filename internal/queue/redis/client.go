package redis

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type TaskQueue interface {
	PublishTask(context.Context, uuid.UUID)
}

type RedisClient struct {
	client *redis.Client
}

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

func (rdc *RedisClient) PublishTask(ctx context.Context, taskID uuid.UUID) {
	rdc.client.LPush(ctx, taskID.String())
}
