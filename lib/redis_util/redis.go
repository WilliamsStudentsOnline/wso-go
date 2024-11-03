package redis_util

import (
	"strconv"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func SetupClient(url string, password string, databaseID int) *redis.Client {
	rdb = redis.NewClient(&redis.Options{
		Addr:     url,
		Password: password,
		DB:       databaseID,
	})

	return rdb
}

func GetClient() *redis.Client {
	return rdb
}

func GetUserSelectionStr(userID uint) string {
	return "user-selection:" + strconv.FormatUint(uint64(userID), 10)
}
