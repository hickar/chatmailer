package mailer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hickar/chatmailer/internal/app/config"
)

type KVStore[K comparable, V any] interface {
	Get(key string) (V, bool)
	Set(key string, value V)
	Remove(key K) bool
}

type Forwarder interface {
	Forward(context.Context, config.ContactPointConfiguration, []*Message) error
}

type MailRetriever interface {
	GetMail(config.ClientConfig) (MailResponse, error)
}

type TaskRunner struct {
	cfg           config.Config
	clientStore   KVStore[string, config.ClientConfig]
	mailRetriever MailRetriever
	forwarder     Forwarder
	logger        *slog.Logger
}

func NewRunner(
	cfg config.Config,
	clientStorage KVStore[string, config.ClientConfig],
	mailRetriever MailRetriever,
	forwarder Forwarder,
	logger *slog.Logger,
) TaskRunner {
	return TaskRunner{
		cfg:           cfg,
		clientStore:   clientStorage,
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
		// Retrieve current client state.
		if stored, ok := r.clientStore.Get(client.Login); ok {
			client = stored
		}

		logger := r.logger.With(slog.String("client", client.Login))
		logger.Debug("starting email retrieval")

		if len(client.ContactPoints) == 0 {
			logger.Error("client has no contact points specified")
			return errors.New("client has no contact points specified")
		}

		// Retrieve messages from client's mailbox.
		mail, err := r.mailRetriever.GetMail(client)
		if err != nil {
			logger.Error(fmt.Sprintf("mail retrieval failed: %v", err))
			return fmt.Errorf("retrieve mail: %w", err)
		}

		logger.Info(fmt.Sprintf("received %d new messages received", len(mail.Messages)))
		if len(mail.Messages) == 0 {
			return nil
		}

		// Update client's last read mail UIDs.
		client.LastUIDNext = mail.LastUID
		client.LastUIDValidity = mail.LastUIDValidity
		r.clientStore.Set(client.Login, client)

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
