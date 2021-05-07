package lease

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLeaseAlreadyExists = errors.New("lease already exists")
	ErrLessorTooFull      = errors.New("lessor has maximum leases")
)

type Lessor struct {
	m            sync.RWMutex
	MaxLeases    int
	TermDuration time.Duration
	leases       []*Lease
}

type Lease struct {
	ID     uuid.UUID
	Expiry time.Time
}

func (l *Lease) IsExpired() bool {
	return time.Now().After(l.Expiry)
}

func NewLessor(maxLeases int, termDuration time.Duration) *Lessor {
	return &Lessor{
		m:            sync.RWMutex{},
		MaxLeases:    maxLeases,
		TermDuration: termDuration,
	}
}

func (l *Lessor) HasLease(id uuid.UUID) bool {
	l.m.RLock()
	defer l.m.RUnlock()

	for _, lease := range l.leases {
		if lease.ID == id && !lease.IsExpired() {
			return true
		}
	}

	return false
}

func (l *Lessor) EndLease(id uuid.UUID) {
	l.removeLease(id)
}

func (l *Lessor) CountUsedLeases() int {
	l.m.RLock()
	defer l.m.RUnlock()

	return len(l.leases)
}

func (l *Lessor) RequestLease() (*Lease, error) {
	l.m.Lock()
	defer l.m.Unlock()

	if len(l.leases)+1 > l.MaxLeases {
		return nil, ErrLessorTooFull
	}

	newLease := Lease{
		ID:     uuid.New(),
		Expiry: time.Now().Add(l.TermDuration),
	}
	l.leases = append(l.leases, &newLease)

	expiryTimer := time.NewTimer(l.TermDuration)

	go func(l *Lessor, expiryTimer *time.Timer, newLease *Lease) {
		<-expiryTimer.C
		l.removeLease(newLease.ID)
	}(l, expiryTimer, &newLease)

	return &newLease, nil
}

func (l *Lessor) removeLease(id uuid.UUID) bool {
	l.m.Lock()
	defer l.m.Unlock()

	idx := -1
	for i, lease := range l.leases {
		if lease.ID == id {
			idx = i
		}
	}
	if idx != -1 {
		l.leases = append(l.leases[:idx], l.leases[idx+1:]...)
		return true
	}

	return false
}
