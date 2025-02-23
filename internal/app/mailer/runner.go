package mailer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hickar/chatmailer/internal/app/config"
	"github.com/hickar/chatmailer/internal/pkg/logger"
)

type ClientStore interface {
	Get(id string) (config.ClientConfig, bool)
	Set(id string, client config.ClientConfig)
}

type Forwarder interface {
	Forward(context.Context, config.ContactPointConfiguration, []*Message) error
}

type MailRetriever interface {
	GetMail(context.Context, config.ClientConfig) (Mail, error)
}

type TaskRunner struct {
	cfg           config.Config
	clientStore   ClientStore
	mailRetriever MailRetriever
	forwarder     Forwarder
	logger        *slog.Logger
}

func NewRunner(
	cfg config.Config,
	clientStore ClientStore,
	mailRetriever MailRetriever,
	forwarder Forwarder,
	logger *slog.Logger,
) TaskRunner {
	return TaskRunner{
		cfg:           cfg,
		clientStore:   clientStore,
		mailRetriever: mailRetriever,
		forwarder:     forwarder,
		logger:        logger,
	}
}

// Run retrieves emails for the all clients and forwards them to configured contact points.
//
// Updates client state (LastUIDNext, LastUIDValidity) in the operational memory storage
// to not re-execute parsing and forwarding for already handled emails next time.
func (r *TaskRunner) Run(ctx context.Context) error {
	for _, client := range r.cfg.Clients {
		ctx := logger.WithAttrs(ctx, slog.String("client", client.Login))

		// Retrieve current client state.
		if stored, ok := r.clientStore.Get(client.Login); ok {
			client = stored
		}

		if len(client.ContactPoints) == 0 {
			return errors.New("client has no contact points specified")
		}

		// Retrieve messages from client's mailbox.
		mail, err := r.mailRetriever.GetMail(ctx, client)
		if err != nil {
			r.logger.ErrorContext(ctx, "mail retrieval failed", slog.Any("error", err))
			return fmt.Errorf("retrieve mail: %w", err)
		}

		// Update client's last read mail UIDs.
		client.LastUIDNext = mail.LastUID
		client.LastUIDValidity = mail.LastUIDValidity
		r.clientStore.Set(client.Login, client)

		r.logger.InfoContext(ctx, fmt.Sprintf("received %d new messages received", len(mail.Messages)))
		if len(mail.Messages) == 0 {
			return nil
		}

		// Forward mail to each contact point specified for
		// current client.
		for _, contact := range client.ContactPoints {
			err = r.forwarder.Forward(ctx, contact, mail.Messages)
			if err != nil {
				return fmt.Errorf("forward message: %w", err)
			}
		}
	}

	return nil
}
