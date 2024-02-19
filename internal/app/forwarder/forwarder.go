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
	client   *http.Client
	botToken string
	logger   *slog.Logger
}

func NewTelegramForwarder(client *http.Client, botToken string, logger *slog.Logger) *telegramForwarder {
	return &telegramForwarder{
		client:   client,
		botToken: botToken,
		logger:   logger,
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
	msgParseMode := telegramParseModeMarkdownV2
	if config.ParseMode != nil {
		msgParseMode = *config.ParseMode
	}

	if msgBody.Len() <= telegramMsgTextSizeLimit {
		reqData := telegramSendMsgRequest{
			ChatID:              config.TGChatID,
			ParseMode:           msgParseMode,
			Text:                msgBody.String(),
			DisableNotification: config.SilentMode,
			ProtectContent:      config.DisableForwarding,
		}
		return t.makeRequest(ctx, telegramAPISendMessageMethod, reqData)
	}

	// Due to Telegram's limit on message text size,
	// we proceed to split and send message in 4096-byte sized chunks.
	for msgBody.Len() > 0 {
		msgChunk := msgBody.Next(telegramMsgTextSizeLimit)
		reqData := telegramSendMsgRequest{
			ChatID:              config.TGChatID,
			ParseMode:           msgParseMode,
			Text:                string(msgChunk),
			DisableNotification: config.SilentMode,
			ProtectContent:      config.DisableForwarding,
		}
		if err := t.makeRequest(ctx, telegramAPISendMessageMethod, reqData); err != nil {
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
		fmt.Sprintf(telegramAPIURLTemplate, t.botToken, method),
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

	var respData telegramResponse
	if err = json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return err
	}

	if !respData.Ok {
		return fmt.Errorf("request failed with error_code '%d' and following description '%s'", respData.Code, respData.Description)
	}

	return nil
}

type telegramSendMsgRequest struct {
	ChatID              int64  `json:"chat_id"`
	ParseMode           string `json:"parse_mode"`
	Text                string `json:"text"`
	DisableNotification bool   `json:"disable_notification,omitempty"`
	ProtectContent      bool   `json:"protect_content,omitempty"`
}

type telegramResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
	Code        int    `json:"error_code"`
}

const telegramMsgTextSizeLimit = 4096

const telegramAPIURLTemplate = "https://api.telegram.org/bot%s/%s"

const telegramAPISendMessageMethod = "sendMessage"

const telegramParseModeHTML = "HTML"

const telegramParseModeMarkdownV2 = "MarkdownV2"

const telegramParseModeMarkdown = "Markdown"
