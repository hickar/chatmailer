package daemon

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/hickar/tg-remailer/internal/app/config"
)

type telegramForwarder struct {
	client *http.Client
	logger *slog.Logger
}

func NewTelegramForwarder(client *http.Client, logger *slog.Logger) *telegramForwarder {
	return &telegramForwarder{
		client: client,
		logger: logger,
	}
}

func (n *telegramForwarder) Forward(_ context.Context, _ config.ContactPointConfiguration, _ []*Message) error {
	// TODO: implement
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
