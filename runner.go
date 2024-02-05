package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type Message struct {
	BodyParts   []BodySegment
	Subject     string
	From        []string
	To          []string
	CC          []string
	ReplyTo     []string
	Date        time.Time
	Mailbox     string
	UID         uint32
	Attachments []Attachment
}

type BodySegment struct {
	MIMEType string
	Content  []byte
	Charset  string
}

type Attachment struct {
	Filename string
	MIMEType string
	Body     []byte
}

type ClientStorage interface {
	Get(user string) (ClientConfig, bool)
	Set(user string, config ClientConfig)
	Remove(user string) bool
}

type Forwarder interface {
	Forward(context.Context, ClientConfig, []*Message) error
}

type TaskRunner struct {
	clientStorage ClientStorage
	mailSource    MailRetriever
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
		mailSource:    mailRetriever,
		forwarder:     forwarder,
		logger:        logger,
	}
}

func (r *TaskRunner) Run(ctx context.Context, client ClientConfig) error {
	r.logger.Debug("starting email retrieval", slog.String("client", client.Login))

	storedClient, ok := r.clientStorage.Get(client.Login)
	if ok {
		client.LastUID = storedClient.LastUID
		client.LastUIDValidity = storedClient.LastUIDValidity
	}

	client, messages, err := r.mailSource.GetMail(client)
	if err != nil {
		r.logger.Error("mail retrieval failed", slog.String("client", client.Login))
		return fmt.Errorf("failed to retrieve mail: %w", err)
	}
	r.logger.Info(fmt.Sprintf("%d new messages received", len(messages)), slog.String("client", client.Login))

	r.clientStorage.Set(client.Login, client)

	err = r.forwarder.Forward(ctx, client, messages)
	if err != nil {
		return fmt.Errorf("failed to forward message: %w", err)
	}

	return nil
}
