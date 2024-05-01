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

func (t *telegramForwarder) Forward(ctx context.Context, config config.ContactPointConfiguration, messages []*mailer.Message) error {
	for _, message := range messages {
		msgContent, err := renderTemplate(message, config.Template)
		if err != nil {
			return fmt.Errorf("failed to render message template: %w", err)
		}

		if err = t.sendMessage(ctx, config, bytes.NewBufferString(msgContent)); err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
	}
	return nil
}

func (t *telegramForwarder) sendMessage(ctx context.Context, config config.ContactPointConfiguration, msgBody *bytes.Buffer) error {
	msgParseMode := tgParseModeMarkdownV2
	if config.ParseMode != nil {
		msgParseMode = *config.ParseMode
	}

	if msgBody.Len() <= tgMsgTextSizeLimit {
		reqData := tgSendMsgRequest{
			ChatID:              config.TGChatID,
			ParseMode:           msgParseMode,
			Text:                msgBody.String(),
			DisableNotification: config.SilentMode,
			ProtectContent:      config.DisableForwarding,
		}
		return t.makeRequest(ctx, tgAPISendMessageMethod, reqData)
	}

	// Due to Telegram's limit on message text size,
	// we proceed to split and send message in 4096-byte sized chunks.
	for msgBody.Len() > 0 {
		msgChunk := msgBody.Next(tgMsgTextSizeLimit)
		reqData := tgSendMsgRequest{
			ChatID:              config.TGChatID,
			ParseMode:           msgParseMode,
			Text:                string(msgChunk),
			DisableNotification: config.SilentMode,
			ProtectContent:      config.DisableForwarding,
		}
		if err := t.makeRequest(ctx, tgAPISendMessageMethod, reqData); err != nil {
			return err
		}
	}

	return nil
}

func (t *telegramForwarder) makeRequest(ctx context.Context, method string, payload any) error {
	reqPayloadBytes, err := json.Marshal(&payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf(tgAPIURLTemplate, t.cfg.BotToken, method),
		bytes.NewReader(reqPayloadBytes),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}

	var respData tgResponse
	if err = json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return err
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

const tgMsgTextSizeLimit = 4096

const tgAPIURLTemplate = "https://api.telegram.org/bot%s/%s"

const tgAPISendMessageMethod = "sendMessage"

const tgParseModeHTML = "HTML"

const tgParseModeMarkdownV2 = "MarkdownV2"

const tgParseModeMarkdown = "Markdown"
