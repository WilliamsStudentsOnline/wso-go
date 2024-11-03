package redis_util

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func SetupClient(url string, password string, databaseID int) (*redis.Client, error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       databaseID,
	})

	if err := rdb.Ping(context.TODO()).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}

func GetClient() *redis.Client {
	return rdb
}

// Wrappers for Get/Set
func Get(ctx context.Context, key string) (interface{}, error) {
	value := rdb.Get(ctx, key)
	return value, value.Err()
}

func Set(ctx context.Context, key string, value interface{}) error {
	return rdb.Set(ctx, key, value, 0).Err()
}

func SetWithExpiration(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	return rdb.Set(ctx, key, value, duration).Err()
}

// Key formatting
func GetUserSelectionStr(userID uint) string {
	return "user-selection:" + strconv.FormatUint(uint64(userID), 10)
}
