package redis_util

import (
	"sync"
	"time"
)

const reconnectEvery = 30 * time.Second

// Lazy reconnects to a redis that was down when the service started
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

// Client is nil while redis is unreachable; RedisClient methods handle a nil receiver
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
