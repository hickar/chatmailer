package main

import (
	"context"
	"errors"
	"time"
)

type scheduler struct {
	settings schedulerSettings
	ticker   *time.Ticker
	quit     chan struct{}
}

type schedulerSettings struct {
	Callback        func()
	Interval        time.Duration
	LaunchInitially bool
}

// Schedule launches a time.Ticker, which launches provided callback on ticker.C channel signal.
// error is returned only on launch in case of invalid settings were provided.
func (s *scheduler) Schedule(settings schedulerSettings) error {
	return s.ScheduleWithCtx(context.Background(), settings)
}

func (s *scheduler) ScheduleWithCtx(ctx context.Context, settings schedulerSettings) error {
	if settings.Interval <= 0 {
		return errors.New("interval must be larger than 0")
	}
	if settings.Callback == nil {
		return errors.New("callback is nil")
	}

	s.quit = make(chan struct{})
	s.settings = settings
	go s.runSchedule(ctx)
	return nil
}

func (s *scheduler) runSchedule(ctx context.Context) {
	if s.settings.LaunchInitially {
		s.settings.Callback()
	}

	s.ticker = time.NewTicker(s.settings.Interval)
	defer s.ticker.Stop()

	for {
		select {
		case <-s.ticker.C:
			s.settings.Callback()
		case <-s.quit:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *scheduler) Stop() {
	close(s.quit)
}
