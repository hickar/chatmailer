package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/hickar/chatmailer/internal/pkg/units"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Forwarders ForwarderConfiguration `yaml:"forwarders"`
	// Interval between email polling tasks.
	MailPollInterval time.Duration `yaml:"mail_poll_interval"`
	// Timeout for individual email processing tasks.
	MailPollTaskTimeout time.Duration `yaml:"mail_poll_task_timeout"`
	// Number of retries for failed email processing tasks.
	RetryCount int `yaml:"retry_count"`
	// Minimum delay between retries in seconds.
	RetryDelayMin int `yaml:"retry_delay_min"`
	// Maximum delay between retries in seconds.
	RetryDelayMax int `yaml:"retry_delay_max"`
	// Logging level
	LogLevel slog.Level `yaml:"log_level"`
	// List of email client configurations.
	Clients []ClientConfig `yaml:"clients"`
}

type ForwarderConfiguration struct {
	Telegram TelegramConfiguration `yaml:"telegram"`
}

type TelegramConfiguration struct {
	BotToken  string `yaml:"bot_token"`
	WebAppURL string `yaml:"web_app_url"`
}

type ClientConfig struct {
	// Email protocol (e.g., imap, pop3).
	Proto string `yaml:"proto"`
	// Email server address.
	Address string `yaml:"address"`
	// Email account username.
	Login string `yaml:"login"`
	// Email account password (stored securely).
	Password string `yaml:"password"`
	// Whether to mark retrieved emails as seen on the server.
	MarkAsSeen bool `yaml:"mark_as_seen"`
	// Optional filters for selecting specific emails.
	Filters []string `yaml:"filters"`
	// Internal state for tracking processed emails. (TODO: Explain usage)
	LastUIDNext uint32 `yaml:"last_uid_next"`
	// Internal state for tracking processed emails. (TODO: Explain usage)
	LastUIDValidity uint32 `yaml:"last_uid_validity"`
	// Whether to include email attachments in notifications.
	IncludeAttachments bool `yaml:"include_attachments"`
	// Maximum size of attachments allowed to be processed and uploaded.
	MaximumAttachmentsSize units.ByteSize `yaml:"maximum_attachments_size"`
	// List of notification destinations.
	ContactPoints []ContactPointConfiguration `yaml:"contact_points"`
}

type ContactPointConfiguration struct {
	// Telegram bot token for sending notifications.
	TGBotToken string `yaml:"tg_bot_token"`
	// Telegram chat ID for receiving notifications.
	TGChatID int64 `yaml:"tg_chat_id"`
	// Whether to send notifications silently (without notification sound).
	SilentMode bool `yaml:"silent_mode"`
	// Whether to disable message forwarding in Telegram.
	DisableForwarding bool `yaml:"disable_forwarding"`
	// Optional template for customizing notification content.
	Template string `yaml:"template"`
	// Forwarding client type, for example telegram.
	Type string `yaml:"type"`
	// Mode for parsing entities in the message text.
	// Possible values: 'HTML', 'MarkdownV2', 'Markdown'.
	ParseMode *string `yaml:"parse_mode,omitempty"`
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
