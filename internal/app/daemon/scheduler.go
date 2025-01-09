package daemon

import (
	"context"
	"errors"
	"time"
)

// Scheduler provides a mechanism to execute a callback function repeatedly at a specified interval.
type Scheduler struct {
	settings schedulerSettings
	ticker   *time.Ticker  // Internal ticker for managing timed events.
	quit     chan struct{} // Channel for signaling termination.
}

// Encapsulates the configuration options for a Scheduler.
type schedulerSettings struct {
	Callback        func()        // Function to be executed at each interval.
	Interval        time.Duration // Duration between callback executions.
	LaunchInitially bool          // Flag indicating whether to execute the callback immediately upon scheduling.
}

func (s *Scheduler) Schedule(settings schedulerSettings) error {
	return s.ScheduleWithCtx(context.Background(), settings)
}

// ScheduleWithCtx launches a Scheduler with the provided settings.
//
// Launches a time.Ticker that signals the execution of the callback function at regular intervals.
// Returns an error only if invalid settings are provided (e.g., interval <= 0 or nil callback).
func (s *Scheduler) ScheduleWithCtx(ctx context.Context, settings schedulerSettings) error {
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

func (s *Scheduler) runSchedule(ctx context.Context) {
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

// Stop gracefully terminates Scheduler.
func (s *Scheduler) Stop() {
	close(s.quit)
}
