package jobctl

import (
	"context"
	"fmt"
	"os"

	"github.com/WilliamsStudentsOnline/wso-go/lib/redis_util"
)

var (
	redisClient     *redis_util.RedisClient
	disabledJobsKey = redis_util.GetDisabledJobsKey()
)

func ConfigureRedis(addr string) error {
	if redisClient != nil {
		return nil
	}
	if addr == "" {
		addr = redis_util.DefaultRedisAddr
	}
	c, err := redis_util.SetupClient(addr, "", redis_util.AdminDatabaseID)
	if err != nil {
		return err
	}
	redisClient = c
	return nil
}

// per-process key for parallel test packages
func ConfigureRedisForTest() error {
	c, err := redis_util.SetupRedisClientForTest()
	if err != nil {
		return err
	}
	redisClient = c
	disabledJobsKey = fmt.Sprintf("%s:test:%d", redis_util.GetDisabledJobsKey(), os.Getpid())
	return c.Del(context.Background(), disabledJobsKey)
}

func IsJobDisabled(ctx context.Context, name string) (bool, error) {
	return redisClient.SIsMember(ctx, disabledJobsKey, name)
}

func DisabledJobs(ctx context.Context) (map[string]bool, error) {
	names, err := redisClient.SMembers(ctx, disabledJobsKey)
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(names))
	for _, n := range names {
		out[n] = true
	}
	return out, nil
}

func SetJobDisabled(ctx context.Context, name string, disabled bool) error {
	if disabled {
		return redisClient.SAdd(ctx, disabledJobsKey, name)
	}
	return redisClient.SRem(ctx, disabledJobsKey, name)
}
