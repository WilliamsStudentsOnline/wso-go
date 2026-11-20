package redis_util

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const DefaultRedisAddr = "127.0.0.1:6379"

var ErrRedisClientNotConfigured error = errors.New("no redis client configured")

type RedisClient struct {
	rdb                  *redis.Client
	configuredForTesting bool
}

func SetupClient(url string, password string, databaseID int) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       databaseID,
	})

	if err := rdb.Ping(context.TODO()).Err(); err != nil {
		return nil, err
	}

	client := RedisClient{
		rdb:                  rdb,
		configuredForTesting: false,
	}

	if databaseID == TestDatabaseID {
		client.configuredForTesting = true
	}

	return &client, nil
}

// Connect to a Redis database for testing
func SetupRedisClientForTest() (*RedisClient, error) {
	return SetupClient(DefaultRedisAddr, "", TestDatabaseID)
}

// Wrappers for Get/Set
func (client *RedisClient) Get(ctx context.Context, key string) (interface{}, error) {
	if client == nil {
		return nil, ErrRedisClientNotConfigured
	}
	return client.rdb.Get(ctx, key).Result()
}

func (client *RedisClient) Set(ctx context.Context, key string, value interface{}) error {
	if client == nil {
		return ErrRedisClientNotConfigured
	}
	return client.rdb.Set(ctx, key, value, 0).Err()
}

func (client *RedisClient) SetWithExpiration(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	if client == nil {
		return ErrRedisClientNotConfigured
	}
	return client.rdb.Set(ctx, key, value, duration).Err()
}

func (client *RedisClient) Del(ctx context.Context, keys ...string) error {
	if client == nil {
		return ErrRedisClientNotConfigured
	}
	return client.rdb.Del(ctx, keys...).Err()
}

// Wrappers for sets
func (client *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	if client == nil {
		return ErrRedisClientNotConfigured
	}
	return client.rdb.SAdd(ctx, key, members...).Err()
}

func (client *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	if client == nil {
		return ErrRedisClientNotConfigured
	}
	return client.rdb.SRem(ctx, key, members...).Err()
}

func (client *RedisClient) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	if client == nil {
		return false, ErrRedisClientNotConfigured
	}
	return client.rdb.SIsMember(ctx, key, member).Result()
}

func (client *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	if client == nil {
		return nil, ErrRedisClientNotConfigured
	}
	return client.rdb.SMembers(ctx, key).Result()
}

// Key formatting
func GetHomepageBannerKey() string {
	return "homepage:banner"
}

func GetDisabledJobsKey() string {
	return "jobs:disabled"
}
