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

	"github.com/hickar/chatmailer/internal/app/config"
	"github.com/hickar/chatmailer/internal/app/daemon"
	"github.com/hickar/chatmailer/internal/app/forwarder"
	"github.com/hickar/chatmailer/internal/app/mailer"
	"github.com/hickar/chatmailer/internal/app/retriever"
	"github.com/hickar/chatmailer/internal/pkg/kvstore"
	xlogger "github.com/hickar/chatmailer/internal/pkg/logger"

	"github.com/emersion/go-imap/v2/imapclient"
)

func main() {
	configPath := flag.String("config", "./config.yaml", "Filepath to configuration file. Default is '.config.yaml'")
	flag.Parse()

	cfg, err := config.NewFromFile(*configPath)
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	// Create logger with custom handler able
	// to store log attributes within context.Context.
	logger := slog.New(xlogger.NewContextHandler(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:       cfg.LogLevel,
			ReplaceAttr: xlogger.ReplaceAttr,
		}),
	))

	runner := mailer.NewRunner(
		cfg,
		kvstore.New[string, config.ClientConfig](),
		retriever.NewIMAPRetriever(
			retriever.ImapDialerFunc(imapclient.DialTLS),
			logger,
		),
		forwarder.NewTelegramForwarder(
			http.DefaultClient,
			cfg.Forwarders.Telegram,
			logger.With(slog.String("module", "telegram_forwarder")),
		),
		logger.With(slog.String("module", "runner")),
	)

	chatmailer := daemon.NewDaemon(
		cfg,
		daemon.NewScheduler(),
		runner,
		logger.With(slog.String("module", "remailer")),
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err = chatmailer.Start(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			logger.Error(fmt.Sprintf("Application exited with error: %s", err), slog.String("module", "main"))
			cancel()
			//nolint:gocritic
			os.Exit(1)
		}
	}

	logger.Info("Application exited successfully")
}
