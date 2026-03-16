package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrSemaphoreAcquireFailed = errors.New("failed to acquire semaphore")
	ErrSemaphoreReleaseFailed = errors.New("failed to release semaphore")
)

type Semaphore struct {
	client      *redis.Client
	key         string
	maxWorkers  int
	lockTimeout time.Duration
}

func NewSemaphore(client *redis.Client, key string, maxWorkers int, lockTimeout time.Duration) *Semaphore {
	return &Semaphore{
		client:      client,
		key:         key,
		maxWorkers:  maxWorkers,
		lockTimeout: lockTimeout,
	}
}

func (s *Semaphore) Acquire(ctx context.Context, workerID string) (bool, error) {
	now := time.Now().Unix()
	
	s.cleanupExpiredLocks(ctx, now)
	
	count, err := s.client.ZCard(ctx, s.key).Result()
	if err != nil {
		return false, err
	}
	
	if count >= int64(s.maxWorkers) {
		return false, nil
	}
	
	score := float64(now + int64(s.lockTimeout.Seconds()))
	added, err := s.client.ZAdd(ctx, s.key, redis.Z{
		Score:  score,
		Member: workerID,
	}).Result()
	
	if err != nil {
		return false, err
	}
	
	return added > 0, nil
}

func (s *Semaphore) Release(ctx context.Context, workerID string) error {
	removed, err := s.client.ZRem(ctx, s.key, workerID).Result()
	if err != nil {
		return err
	}
	
	if removed == 0 {
		return fmt.Errorf("%w: worker %s not found", ErrSemaphoreReleaseFailed, workerID)
	}
	
	return nil
}

func (s *Semaphore) cleanupExpiredLocks(ctx context.Context, now int64) {
	s.client.ZRemRangeByScore(ctx, s.key, "-inf", fmt.Sprintf("%d", now))
}

func (s *Semaphore) GetActiveWorkers(ctx context.Context) (int, error) {
	count, err := s.client.ZCard(ctx, s.key).Result()
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

type DistributedLock struct {
	client      *redis.Client
	key         string
	value       string
	lockTimeout time.Duration
}

func NewDistributedLock(client *redis.Client, resourceKey string, lockTimeout time.Duration) *DistributedLock {
	return &DistributedLock{
		client:      client,
		key:         fmt.Sprintf("lock:%s", resourceKey),
		value:       uuid.New().String(),
		lockTimeout: lockTimeout,
	}
}

func (l *DistributedLock) Acquire(ctx context.Context) (bool, error) {
	acquired, err := l.client.SetNX(ctx, l.key, l.value, l.lockTimeout).Result()
	if err != nil {
		return false, err
	}
	return acquired, nil
}

func (l *DistributedLock) Release(ctx context.Context) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	
	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Result()
	if err != nil {
		return err
	}
	
	if result == int64(0) {
		return errors.New("lock not owned by this instance")
	}
	
	return nil
}

func (l *DistributedLock) Extend(ctx context.Context) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`
	
	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value, int(l.lockTimeout.Seconds())).Result()
	if err != nil {
		return err
	}
	
	if result == int64(0) {
		return errors.New("lock not owned by this instance")
	}
	
	return nil
}
