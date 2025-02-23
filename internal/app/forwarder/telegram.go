package forwarder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/hickar/chatmailer/internal/app/config"
	"github.com/hickar/chatmailer/internal/app/mailer"
)

const (
	tgMsgTextSizeLimit     = 4096
	tgAPIURLTemplate       = "https://api.telegram.org/bot%s/%s"
	tgAPISendMessageMethod = "sendMessage"

	tgParseModeHTML       = "HTML"
	tgParseModeMarkdownV2 = "MarkdownV2"
	tgParseModeMarkdown   = "Markdown"
)

type telegramForwarder struct {
	client *http.Client
	cfg    config.TelegramConfiguration
	logger *slog.Logger
}

func NewTelegramForwarder(client *http.Client, cfg config.TelegramConfiguration, logger *slog.Logger) *telegramForwarder {
	return &telegramForwarder{
		client: client,
		cfg:    cfg,
		logger: logger,
	}
}

func (tf *telegramForwarder) Forward(ctx context.Context, cfg config.ContactPointConfiguration, messages []*mailer.Message) error {
	messages[0].BodyParts = messages[0].BodyParts[1:]

	for _, message := range messages {
		content, err := renderTemplate(message, cfg.Template)
		if err != nil {
			return fmt.Errorf("render message template: %w", err)
		}

		if err = tf.sendMessage(ctx, cfg, bytes.NewBufferString(content)); err != nil {
			return fmt.Errorf("send message: %w", err)
		}
	}

	return nil
}

func (tf *telegramForwarder) sendMessage(ctx context.Context, cfg config.ContactPointConfiguration, body *bytes.Buffer) error {
	parseMode := tgParseModeMarkdownV2
	if cfg.ParseMode != nil {
		parseMode = *cfg.ParseMode
	}

	// Due to Telegram's limit on message text size,
	// we proceed to split and send message in 4096-byte sized chunks.
	for body.Len() > 0 {
		payload := body.Next(tgMsgTextSizeLimit)
		req := tgSendMsgRequest{
			ChatID:              cfg.TGChatID,
			ParseMode:           parseMode,
			Text:                string(payload),
			DisableNotification: cfg.SilentMode,
			ProtectContent:      cfg.DisableForwarding,
		}
		if err := tf.makeRequest(ctx, tgAPISendMessageMethod, req); err != nil {
			return fmt.Errorf("make request: %w", err)
		}
	}

	return nil
}

func (tf *telegramForwarder) makeRequest(ctx context.Context, method string, payload any) error {
	b, err := json.Marshal(&payload)
	if err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf(tgAPIURLTemplate, tf.cfg.BotToken, method),
		bytes.NewReader(b),
	)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := tf.client.Do(req)
	if err != nil {
		return fmt.Errorf("request: %w", err)
	}

	var respData tgResponse
	if err = json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return fmt.Errorf("decode json: %w", err)
	}

	if !respData.Ok {
		return fmt.Errorf("request failed with error_code '%d' and following description '%s'", respData.Code, respData.Description)
	}

	return nil
}

type tgSendMsgRequest struct {
	ChatID              int64          `json:"chat_id"`
	ParseMode           string         `json:"parse_mode"`
	Text                string         `json:"text"`
	DisableNotification bool           `json:"disable_notification,omitempty"`
	ProtectContent      bool           `json:"protect_content,omitempty"`
	ReplyMarkup         tgInlineMarkup `json:"reply_markup,omitempty"`
}

type tgInlineMarkup struct {
	Keyboard [][]tgInlineButton `json:"inline_button"`
}

type tgInlineButton struct {
	Text         string       `json:"text"`
	URL          string       `json:"url,omitempty"`
	CallbackData string       `json:"callback_data,omitempty"`
	WebApp       tgWebAppInfo `json:"web_app,omitempty"`
}

type tgWebAppInfo struct {
	URL string `json:"url"`
}

type tgResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
	Code        int    `json:"error_code"`
}
