package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/emersion/go-imap/v2/imapclient"

	"github.com/hickar/tg-remailer/internal/app/config"
	"github.com/hickar/tg-remailer/internal/app/daemon"
	"github.com/hickar/tg-remailer/internal/app/storage"
)

var (
	configFilepath = flag.String("config", "./config.yaml", "Filepath to configuration file. Default is '.config.yaml'")
	envFilepath    = flag.String("env-file", "", "Filepath to environment variables file. Default is '.env'")
)

func main() {
	flag.Parse()

	cfg, err := config.LoadConfig(*configFilepath, *envFilepath)
	if err != nil {
		log.Fatalf(fmt.Sprintf("failed to load configuration: %s", err))
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(cfg.LogLevel),
	}))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)
	defer cancel()

	runner := daemon.NewRunner(
		storage.NewInMemoryStorage(),
		daemon.NewIMAPRetriever(daemon.ImapDialerFunc(imapclient.DialTLS)),
		daemon.NewForwarder(&http.Client{}, logger, config.ContactPointConfiguration}),
		logger.With(slog.String("module", "runner")),
	)

	remailer := daemon.NewRemailer(
		cfg,
		&daemon.Scheduler{},
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
