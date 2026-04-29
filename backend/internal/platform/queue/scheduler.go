package queue

import (
	"context"
	"time"
)

type Scheduler struct {
	tickers []*time.Ticker
}

func NewScheduler() *Scheduler {
	return &Scheduler{tickers: []*time.Ticker{}}
}

func (s *Scheduler) Every(ctx context.Context, interval time.Duration, fn func(context.Context)) {
	if interval <= 0 {
		return
	}
	t := time.NewTicker(interval)
	s.tickers = append(s.tickers, t)
	go func() {
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				fn(ctx)
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	for _, t := range s.tickers {
		t.Stop()
	}
}
