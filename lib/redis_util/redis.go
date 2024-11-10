package redis_util

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

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
	return SetupClient("localhost:6379", "", TestDatabaseID)
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

// Key formatting
func GetUserSelectionStr(userID uint) string {
	return "user-selection:" + strconv.FormatUint(uint64(userID), 10)
}
