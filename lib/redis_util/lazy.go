package redis_util

import (
	"sync"
	"time"
)

const reconnectEvery = 30 * time.Second

type Lazy struct {
	addr    string
	db      int
	mu      sync.Mutex
	client  *RedisClient
	lastTry time.Time
}

func NewLazy(addr string, db int) *Lazy {
	if addr == "" {
		addr = DefaultRedisAddr
	}
	return &Lazy{addr: addr, db: db}
}

func NewLazyForTest() (*Lazy, error) {
	c, err := SetupRedisClientForTest()
	return &Lazy{addr: DefaultRedisAddr, db: TestDatabaseID, client: c}, err
}

// nil while redis is down
func (l *Lazy) Client() *RedisClient {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.client == nil && time.Since(l.lastTry) >= reconnectEvery {
		l.lastTry = time.Now()
		l.client, _ = SetupClient(l.addr, "", l.db)
	}
	return l.client
}
