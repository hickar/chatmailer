package mailer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hickar/chatmailer/internal/app/config"
)

type ClientStorage interface {
	Get(user string) (config.ClientConfig, bool)
	Set(user string, config config.ClientConfig)
	Remove(user string) bool
}

type Forwarder interface {
	Forward(context.Context, config.ContactPointConfiguration, []*Message) error
}

type MailRetriever interface {
	GetMail(config.ClientConfig) (MailResponse, error)
}

type TaskRunner struct {
	clientStorage ClientStorage
	mailRetriever MailRetriever
	forwarder     Forwarder
	logger        *slog.Logger
}

func NewRunner(
	clientStorage ClientStorage,
	mailRetriever MailRetriever,
	forwarder Forwarder,
	logger *slog.Logger,
) TaskRunner {
	return TaskRunner{
		clientStorage: clientStorage,
		mailRetriever: mailRetriever,
		forwarder:     forwarder,
		logger:        logger,
	}
}

// Run retrieves emails for the specified client and forwards them to configured contact points.
//
// Updates client state (LastUIDNext, LastUIDValidity) in the operational memory storage
// to not re-execute parsing and forwarding for already handled emails next time.
//
// Handles errors during retrieval and forwarding.
// Returns an error if any of the following occur:
// - The client has no configured contact points.
// - Mail retrieval fails.
// - Forwarding to any contact point fails.
func (r *TaskRunner) Run(ctx context.Context, client config.ClientConfig) error {
	logger := r.logger.With(slog.String("client", client.Login))
	logger.Debug("starting email retrieval")

	if len(client.ContactPoints) == 0 {
		logger.Error("client has no contact points specified")
		return errors.New("client has no contact points specified")
	}

	storedClient, ok := r.clientStorage.Get(client.Login)
	if ok {
		client.LastUIDNext = storedClient.LastUIDNext
		client.LastUIDValidity = storedClient.LastUIDValidity
	}

	mailResp, err := r.mailRetriever.GetMail(client)
	if err != nil {
		logger.Error(fmt.Sprintf("mail retrieval failed: %v", err))
		return fmt.Errorf("failed to retrieve mail: %w", err)
	}

	if len(mailResp.Messages) == 0 {
		logger.Info("no new messages received")
		return nil
	}
	logger.Info(fmt.Sprintf("%d new messages received", len(mailResp.Messages)))

	client.LastUIDNext = mailResp.LastUID
	client.LastUIDValidity = mailResp.LastUIDValidity
	r.clientStorage.Set(client.Login, client)

	for _, contactPoint := range client.ContactPoints {
		err = r.forwarder.Forward(ctx, contactPoint, mailResp.Messages)
		if err != nil {
			return fmt.Errorf("failed to forward message: %w", err)
		}
	}

	return nil
}
