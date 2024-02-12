package forwarder

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/hickar/tg-remailer/internal/app/config"
	"github.com/hickar/tg-remailer/internal/app/mailer"
)

type telegramForwarder struct {
	client *http.Client
	logger *slog.Logger
}

const TypeTelegram = "telegram"

// NewForwarder creates new instance of object implementing Forwarder interface.
// Choose type of client which will receive emails.
// Default is Telegram, but it's possible to add new clients in future
// if we implement loop.
func NewForwarder(
	client *http.Client,
	logger *slog.Logger,
	forwarderType string,
) (mailer.Forwarder, error) {
	switch forwarderType {
	case TypeTelegram:
		return newTelegramForwarder(client, logger.With(slog.String("module", "telegram_forwarder"))), nil
	default:
		return nil, fmt.Errorf("unsupported forwarder type: %s", forwarderType)
	}
}

func newTelegramForwarder(client *http.Client, logger *slog.Logger) *telegramForwarder {
	return &telegramForwarder{
		client: client,
		logger: logger,
	}
}

func (t *telegramForwarder) Forward(ctx context.Context, config config.ContactPointConfiguration, messages []*mailer.Message) error {
	return nil
}

func sendMessage(ctx context.Context) error {
	return nil
}

type telegramSendMsgRequest struct {
	ChatID              int64  `json:"chat_id"`
	ParseMode           string `json:"parse_mode"`
	Text                string `json:"text"`
	DisableNotification bool   `json:"disable_notification"`
	ProtectContent      bool   `json:"protect_content"`
}
