package redis_util

import (
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func setupClient(url string, password string, databaseID int) *redis.Client {
	rdb = redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       databaseID,
	})

	return rdb
}

func SetupClientForTest(databaseID int) *redis.Client { // TODO move to test_util
	return setupClient("localhost:6379", "", databaseID)
}

func GetClient() *redis.Client {
	return rdb
}
