package ticker

import (
	"context"
	"time"
)

type Ticker struct {
	interval time.Duration
}

func NewTicker(interval int) *Ticker {
	return &Ticker{
		interval: time.Duration(interval) * time.Second,
	}
}

// Start runs the task at each interval until the context is cancelled.
func (t *Ticker) Start(ctx context.Context, task func()) {
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			task()
		}
	}
}
