package usecase

import (
	"context"
	"time"

	"ArticlesScanner/internal/ports"
)

// Scheduler wires the cron-like driver with the pipeline use case.
type Scheduler struct {
	driver   ports.Scheduler
	pipeline *Pipeline
}

// NewScheduler returns a helper to start/stop recurring jobs.
func NewScheduler(driver ports.Scheduler, pipeline *Pipeline) *Scheduler {
	return &Scheduler{driver: driver, pipeline: pipeline}
}

// Start registers the pipeline with the provided scheduler.
func (s *Scheduler) Start(ctx context.Context) error {
	if s.driver == nil || s.pipeline == nil {
		return nil
	}

	job := func(trigger time.Time) {
		_ = s.pipeline.ProcessDay(ctx, trigger)
	}

	return s.driver.Start(ctx, job)
}

// Stop gracefully tears down the underlying scheduler.
func (s *Scheduler) Stop(ctx context.Context) error {
	if s.driver == nil {
		return nil
	}

	return s.driver.Stop(ctx)
}
