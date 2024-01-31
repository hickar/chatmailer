package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Remailer struct {
	cfg       Config
	logger    *slog.Logger
	scheduler Scheduler
	runner    TaskRunner
}

type Scheduler interface {
	Schedule(func(), time.Duration)
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

	for _, client := range r.cfg.Clients {
		err := r.runner.Run(ctx, client)
		if err != nil {
			return fmt.Errorf("initial task execution failed: %w", err)
		}
	}

	go func() {
		r.scheduler.Schedule(func() {
			for _, client := range r.cfg.Clients {
				err := r.runner.Run(ctx, client)
				if err != nil {
					errCh <- fmt.Errorf("task execution failed: %w", err)
				}
			}
		}, time.Minute)
	}()
	defer r.scheduler.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
