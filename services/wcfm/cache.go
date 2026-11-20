package wcfm

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/lib/redis_util"
	"go.uber.org/zap"
)

const staleTTL = 24 * time.Hour

var store *redis_util.Lazy

func ConfigureControllerForTest() error {
	var err error
	store, err = redis_util.NewLazyForTest()
	return err
}

type flight struct {
	done    chan struct{}
	payload []byte
	err     error
}

var (
	flightsMu sync.Mutex
	flights   = map[string]*flight{}

	errFetchPanicked = errors.New("wcfm fetch panicked")
)

// prevent spam polling / race on song end
func shared(key string, fetch func() ([]byte, error)) ([]byte, error) {
	flightsMu.Lock()
	if f, ok := flights[key]; ok {
		flightsMu.Unlock()
		<-f.done
		return f.payload, f.err
	}
	f := &flight{done: make(chan struct{}), err: errFetchPanicked}
	flights[key] = f
	flightsMu.Unlock()

	defer func() {
		flightsMu.Lock()
		delete(flights, key)
		flightsMu.Unlock()
		close(f.done)
	}()
	f.payload, f.err = fetch()
	return f.payload, f.err
}

func redisGet(ctx context.Context, key string) ([]byte, bool) {
	val, err := store.Client().Get(ctx, key)
	if err != nil {
		return nil, false
	}
	s, ok := val.(string)
	return []byte(s), ok
}

func cached[T any](ctx context.Context, log *zap.SugaredLogger, name string, ttl time.Duration, fetch func(context.Context) (T, error)) (T, error) {
	var out T
	key := redis_util.GetWCFMCacheKey(name)
	staleKey := redis_util.GetWCFMCacheKey("stale:" + name)
	if payload, ok := redisGet(ctx, key); ok && json.Unmarshal(payload, &out) == nil {
		return out, nil
	}

	payload, err := shared(key, func() ([]byte, error) {
		// allow someone to hang up
		ctx := context.WithoutCancel(ctx)
		val, err := fetch(ctx)
		if err != nil {
			return nil, err
		}
		b, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		rc := store.Client()
		_ = rc.SetWithExpiration(ctx, key, string(b), ttl)
		_ = rc.SetWithExpiration(ctx, staleKey, string(b), staleTTL)
		return b, nil
	})
	if err != nil {
		stale, ok := redisGet(ctx, staleKey)
		if !ok || json.Unmarshal(stale, &out) != nil {
			return out, err
		}
		log.Warnf("wcfm %s: serving stale copy: %v", name, err)
		return out, nil
	}

	err = json.Unmarshal(payload, &out)
	return out, err
}
