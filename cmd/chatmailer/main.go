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

	"github.com/hickar/chatmailer/internal/app/config"
	"github.com/hickar/chatmailer/internal/app/daemon"
	"github.com/hickar/chatmailer/internal/app/forwarder"
	"github.com/hickar/chatmailer/internal/app/mailer"
	"github.com/hickar/chatmailer/internal/app/retriever"
	"github.com/hickar/chatmailer/internal/app/storage"
)

var (
	configFilepath = flag.String("config", "./config.yaml", "Filepath to configuration file. Default is '.config.yaml'")
	envFilepath    = flag.String("env-file", "./.env", "Filepath to environment variables file. Default is '.env'")
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

	runner := mailer.NewRunner(
		storage.NewInMemoryStorage(),
		retriever.NewIMAPRetriever(retriever.ImapDialerFunc(imapclient.DialTLS)),
		forwarder.NewTelegramForwarder(
			&http.Client{},
			cfg.Forwarders.Telegram,
			logger.With(slog.String("module", "telegram_forwarder")),
		),
		logger.With(slog.String("module", "runner")),
	)

	remailer := daemon.NewDaemon(
		cfg,
		&daemon.Scheduler{},
		runner,
		logger.With(slog.String("module", "remailer")),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)
	defer cancel()

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
