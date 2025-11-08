package scheduler

import (
	"context"
	"time"

	"ArticlesScanner/internal/ports"
)

// CronScheduler is a lightweight placeholder scheduler using time.Ticker.
type CronScheduler struct {
	spec string
	stop chan struct{}
}

var _ ports.Scheduler = (*CronScheduler)(nil)

// NewCronScheduler builds a scheduler configured via cron expression string.
func NewCronScheduler(spec string) *CronScheduler {
	return &CronScheduler{spec: spec}
}

// Start begins ticking; replace with real cron later.
func (c *CronScheduler) Start(ctx context.Context, job func(time.Time)) error {
	if job == nil {
		return nil
	}

	if c.stop != nil {
		return nil
	}

	c.stop = make(chan struct{})
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		job(time.Now())
		for {
			select {
			case t := <-ticker.C:
				job(t)
			case <-ctx.Done():
				return
			case <-c.stop:
				return
			}
		}
	}()

	return nil
}

// Stop halts the ticker goroutine.
func (c *CronScheduler) Stop(ctx context.Context) error {
	if c.stop == nil {
		return nil
	}
	close(c.stop)
	c.stop = nil
	return nil
}
