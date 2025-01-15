package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Forwarders          ForwarderConfiguration `yaml:"forwarders"`
	MailPollInterval    time.Duration          `yaml:"mail_poll_interval"`     // Interval between email polling tasks.
	MailPollTaskTimeout time.Duration          `yaml:"mail_poll_task_timeout"` // Timeout for individual email processing tasks.
	RetryCount          int                    `yaml:"retry_count"`            // Number of retries for failed email processing tasks.
	RetryDelayMin       int                    `yaml:"retry_delay_min"`        // Minimum delay between retries in seconds.
	RetryDelayMax       int                    `yaml:"retry_delay_max"`        // Maximum delay between retries in seconds.
	LogLevel            int                    `yaml:"log_level"`              // Logging level (e.g., 0: debug, 1: info, etc.).
	Clients             []ClientConfig         `yaml:"clients"`                // List of email client configurations.
}

type ForwarderConfiguration struct {
	Telegram TelegramConfiguration `yaml:"telegram"`
}

type TelegramConfiguration struct {
	BotToken  string `yaml:"bot_token"`
	WebAppURL string `yaml:"web_app_url"`
}

type ClientConfig struct {
	Proto              string                      `yaml:"proto"`               // Email protocol (e.g., imap, pop3).
	Address            string                      `yaml:"address"`             // Email server address.
	Login              string                      `yaml:"login"`               // Email account username.
	Password           string                      `yaml:"password"`            // Email account password (stored securely).
	MarkAsSeen         bool                        `yaml:"mark_as_seen"`        // Whether to mark retrieved emails as seen on the server.
	Filters            []string                    `yaml:"filters"`             // Optional filters for selecting specific emails.
	LastUIDNext        uint32                      `yaml:"last_uid_next"`       // Internal state for tracking processed emails. (TODO: Explain usage)
	LastUIDValidity    uint32                      `yaml:"last_uid_validity"`   // Internal state for tracking processed emails. (TODO: Explain usage)
	IncludeAttachments bool                        `yaml:"include_attachments"` // Whether to include email attachments in notifications.
	ContactPoints      []ContactPointConfiguration `yaml:"contact_points"`      // List of notification destinations.
}

type ContactPointConfiguration struct {
	TGBotToken        string  `yaml:"tg_bot_token"`         // Telegram bot token for sending notifications.
	TGChatID          int64   `yaml:"tg_chat_id"`           // Telegram chat ID for receiving notifications.
	SilentMode        bool    `yaml:"silent_mode"`          // Whether to send notifications silently (without notification sound).
	DisableForwarding bool    `yaml:"disable_forwarding"`   // Whether to disable message forwarding in Telegram.
	Template          string  `yaml:"template"`             // Optional template for customizing notification content.
	Type              string  `yaml:"type"`                 // Forwarding client type, for example telegram.
	ParseMode         *string `yaml:"parse_mode,omitempty"` // Mode for parsing entities in the message text. Possible values: 'HTML', 'MarkdownV2', 'Markdown'.
}

func NewFromFile(configPath string) (Config, error) {
	var cfg Config

	//TODO: Need to consider secure alternatives in production.
	//nolint:gosec
	file, err := os.Open(configPath)
	if err != nil {
		return cfg, fmt.Errorf("open file: %w", err)
	}

	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return cfg, fmt.Errorf("decode yaml: %w", err)
	}

	return cfg, nil
}
