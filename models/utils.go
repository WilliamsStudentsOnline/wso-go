package models

import "time"

type Clock interface {
	Now() time.Time
}

type localClock struct{}

func (localClock) Now() time.Time { return time.Now().Local() }
