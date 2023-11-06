package redistore

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient interface {
	Get(ctx context.Context, key *SessionKey) ([]byte, bool, error)
	Set(ctx context.Context, key *SessionKey, data []byte, expiration time.Duration) error
	SetNX(ctx context.Context, key *SessionKey, data []byte, expiration time.Duration) (bool, error)
	Del(ctx context.Context, key *SessionKey) error
	Close(ctx context.Context) error
}

type GoRedisV9Client struct {
	Client *redis.Client
}

func (c *GoRedisV9Client) Get(ctx context.Context, key *SessionKey) ([]byte, bool, error) {
	bytes, err := c.Client.Get(ctx, key.ToString()).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("failed to get a value from redis through go-redis/v9: %w", err)
	}
	return bytes, true, nil
}

func (c *GoRedisV9Client) Set(ctx context.Context, key *SessionKey, data []byte, expiration time.Duration) error {
	err := c.Client.Set(ctx, key.ToString(), data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set a value to redis through go-redis/v9: %w", err)
	}
	return nil
}

func (c *GoRedisV9Client) SetNX(ctx context.Context, key *SessionKey, data []byte, expiration time.Duration) (bool, error) {
	succeeded, err := c.Client.SetNX(ctx, key.ToString(), data, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to setNX a value to redis through go-redis/v9: %w", err)
	}
	return succeeded, nil
}

func (c *GoRedisV9Client) Del(ctx context.Context, key *SessionKey) error {
	err := c.Client.Del(ctx, key.ToString()).Err()
	if err != nil {
		return fmt.Errorf("failed to delete a value from redis through go-redis/v9: %w", err)
	}
	return nil
}

func (c *GoRedisV9Client) Close(ctx context.Context) error {
	err := c.Client.Close()
	if err != nil {
		return fmt.Errorf("failed to close the connection for redis through go-redis/v9: %w", err)
	}
	return nil
}
