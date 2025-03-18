package ticker

import "time"

type Ticker struct {
	interval time.Duration
}

func NewTicker(interval int) *Ticker {
	return &Ticker{
		interval: time.Duration(interval) * time.Second,
	}
}

func (t *Ticker) Start(task func()) {
	ticker := time.NewTicker(t.interval)
	for range ticker.C {
		task()
	}
}
