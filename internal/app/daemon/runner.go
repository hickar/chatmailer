package daemon

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/hickar/tg-remailer/internal/app/config"
)

type ClientStorage interface {
	Get(user string) (config.ClientConfig, bool)
	Set(user string, config config.ClientConfig)
	Remove(user string) bool
}

type Forwarder interface {
	Forward(context.Context, config.ContactPointConfiguration, []*Message) error
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

func (r *TaskRunner) Run(ctx context.Context, client config.ClientConfig) error {
	logger := r.logger.With(slog.String("client", client.Login))
	logger.Debug("starting email retrieval", slog.String("client", client.Login))

	if len(client.Filters) == 0 {
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
		logger.Error("mail retrieval failed", slog.String("client", client.Login))
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

	for _, contactPoint := range client.ContactPoint {
		err = r.forwarder.Forward(ctx, contactPoint, mailResp.Messages)
		if err != nil {
			return fmt.Errorf("failed to forward message: %w", err)
		}
	}

	return nil
}
