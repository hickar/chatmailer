package main

import (
	"context"
	"fmt"
	"log/slog"
)

type Remailer struct {
	cfg       Config
	logger    *slog.Logger
	scheduler Scheduler
	runner    TaskRunner
}

type Scheduler interface {
	ScheduleWithCtx(context.Context, schedulerSettings) error
	Stop()
}

func NewRemailer(
	cfg Config,
	scheduler Scheduler,
	runner TaskRunner,
	logger *slog.Logger,
) *Remailer {
	return &Remailer{
		cfg:       cfg,
		scheduler: scheduler,
		runner:    runner,
		logger:    logger,
	}
}

func (r *Remailer) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		err := r.scheduler.ScheduleWithCtx(ctx, schedulerSettings{
			LaunchInitially: true,
			Interval:        r.cfg.MailPollInterval,
			Callback: func() {
				tctx, cancel := context.WithTimeout(ctx, r.cfg.MailPollTaskTimeout)
				defer cancel()

				for _, client := range r.cfg.Clients {
					err := r.runner.Run(tctx, client)
					if err != nil {
						// TODO: handle client configuration ignoring
						// in case of invalid settings specified
						errCh <- fmt.Errorf("task execution failed: %w", err)
					}
				}
			},
		})
		if err != nil {
			errCh <- fmt.Errorf("error occurred on scheduler launch: %w", err)
		}
	}()
	defer r.scheduler.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
