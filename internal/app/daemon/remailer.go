package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hickar/tg-remailer/internal/app/config"
)

type Remailer struct {
	cfg       config.Config
	logger    *slog.Logger
	scheduler scheduler
	runner    TaskRunner
}

type scheduler interface {
	Schedule(func(), time.Duration)
	Stop()
}

func NewRemailer(
	cfg config.Config,
	scheduler scheduler,
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

// Launches scheduler, which utilizes built-in Ticker (https://pkg.go.dev/time#Ticker),
// and performs emails retrieval from mail server with graceful shutdown and high-level error handling.
func (r *Remailer) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	// for _, client := range r.cfg.Clients {
	// 	err := r.runner.Run(ctx, client)
	// 	if err != nil {
	// 		return fmt.Errorf("initial task execution failed: %w", err)
	// 	}
	// }

	// Every 10 minutes execute TaskRunner job with core logic of retrieval emails
	// using IMAP protocol, parse them and forward to specified channel (as of now Telegram).
	go func() {
		r.scheduler.Schedule(func() {
			for _, client := range r.cfg.Clients {
				err := r.runner.Run(ctx, client)
				if err != nil {
					errCh <- fmt.Errorf("task execution failed: %w", err)
				}
			}
		}, time.Minute*10)
	}()
	defer r.scheduler.Stop()

	// Graceful termination and error handling
	select {
	// If the context is canceled (e.g., through external signal)
	// returning the context's error to indicate graceful termination.
	case <-ctx.Done():
		return ctx.Err()

	// If an error occurs during task execution returning the received
	// error to signal the failure.
	case err := <-errCh:
		return err
	}
}
