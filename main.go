package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	configFilepath = flag.String("config", "./config.yaml", "Filepath to configuration file. Default is '.config.yaml'")
	envFilepath    = flag.String("env-file", "", "Filepath to environment variables file. Default is '.env'")
)

func main() {
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	cfg, err := loadConfig(*configFilepath, *envFilepath)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to load configuration: %s", err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)
	defer cancel()

	runner := NewRunner(
		newInMemoryStorage(),
		MailSourceFunc(IMAPGetMailFunc),
		NewTelegramForwarder(&http.Client{}, logger.With(slog.String("module", "telegram_forwarder"))),
		logger.With(slog.String("module", "runner")),
	)

	remailer := NewRemailer(
		cfg,
		&scheduler{},
		runner,
		logger.With(slog.String("module", "remailer")),
	)

	if err = remailer.Start(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			logger.Error(fmt.Sprintf("Application exited with error: %s", err), slog.String("module", "main"))
			cancel()
			//nolint:gocritic
			os.Exit(1)
		}
	}

	cancel()
}
