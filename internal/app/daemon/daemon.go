package daemon

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/hickar/chatmailer/internal/app/config"
	"github.com/hickar/chatmailer/internal/app/mailer"
)

type Daemon struct {
	cfg       config.Config
	logger    *slog.Logger
	scheduler scheduler
	runner    mailer.TaskRunner
}

type scheduler interface {
	ScheduleWithCtx(context.Context, schedulerSettings) error
	Stop()
}

func NewDaemon(
	cfg config.Config,
	scheduler scheduler,
	runner mailer.TaskRunner,
	logger *slog.Logger,
) *Daemon {
	return &Daemon{
		cfg:       cfg,
		scheduler: scheduler,
		runner:    runner,
		logger:    logger,
	}
}

// Start launches scheduler, which utilizes built-in Ticker (https://pkg.go.dev/time#Ticker),
// and performs emails retrieval from mail server with graceful shutdown and high-level error handling.
func (r *Daemon) Start(ctx context.Context) error {
	errCh := make(chan error, 1)

	// Executes the TaskRunner job periodically with configurable mail polling interval.
	// The job retrieves emails using IMAP, parses them, and forwards them to a specified channel.
	go func() {
		err := r.scheduler.ScheduleWithCtx(ctx, schedulerSettings{
			LaunchInitially: true,                   // Execute the job immediately upon scheduling.
			Interval:        r.cfg.MailPollInterval, // Time interval between job executions.
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
			errCh <- fmt.Errorf("error occurred while launching the scheduler: %w", err)
		}
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
