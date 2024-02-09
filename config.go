package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	TGBotToken       string         `yaml:"tg_bot_token"`
	MailPollInterval time.Duration  `yaml:"mail_poll_interval"`
	RetryCount       int            `yaml:"retry_count"`
	RetryDelayMin    int            `yaml:"retry_delay_min"`
	RetryDelayMax    int            `yaml:"retry_delay_max"`
	LogLevel         string         `yaml:"log_level"`
	Clients          []ClientConfig `yaml:"clients"`
}

type ClientConfig struct {
	Proto              string   `yaml:"proto"`
	Address            string   `yaml:"address"`
	Login              string   `yaml:"login"`
	Password           string   `yaml:"password"`
	MarkAsSeen         bool     `yaml:"mark_as_seen"`
	Filters            []string `yaml:"filters"`
	LastUIDNext        uint32
	LastUIDValidity    uint32
	IncludeAttachments bool                        `yaml:"include_attachments"`
	ContactPoint       []ContactPointConfiguration `yaml:"contact_points"`
}

type ContactPointConfiguration struct {
	TGBotToken        string `yaml:"tg_bot_token"`
	TGChatID          int64  `yaml:"tg_chat_id"`
	SilentMode        bool   `yaml:"silent_mode"`
	DisableForwarding bool   `yaml:"disable_forwarding"`
	Template          string `yaml:"template"`
}

func loadConfig(cfgFilepath, envFilepath string) (Config, error) {
	var cfg Config

	if _, err := os.Stat(envFilepath); err == nil {
		if err = godotenv.Load(envFilepath); err != nil {
			return cfg, fmt.Errorf("unable to load environment variables from file: %w", err)
		}
	}

	//nolint:gosec
	fileBytes, err := os.ReadFile(cfgFilepath)
	if err != nil {
		switch {
		case errors.Is(err, os.ErrNotExist):
			return cfg, fmt.Errorf("configuration file at this cfgFilepath doesn't exist: %w", err)
		case errors.Is(err, os.ErrPermission):
			return cfg, fmt.Errorf("permission denied for accessing configuration file: %w", err)
		default:
			return cfg, fmt.Errorf("unexpected error during reading configuration file: %w", err)
		}
	}

	envExpanded := os.ExpandEnv(string(fileBytes))
	if err = yaml.Unmarshal([]byte(envExpanded), &cfg); err != nil {
		return cfg, fmt.Errorf("unable to unmarshal configuration file: %w", err)
	}

	return cfg, nil
}
