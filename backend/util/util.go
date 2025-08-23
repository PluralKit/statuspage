package util

import (
	"errors"
	"log/slog"
)

var ErrNotFound = errors.New("resource not found")
var ErrInvalid = errors.New("invalid struct/type data")

type SlogLevel slog.Level

var LevelMappings = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func (l *SlogLevel) UnmarshalText(text []byte) error {
	lvl, ok := LevelMappings[string(text)]
	if !ok {
		return errors.New("invalid log level")
	}
	*l = SlogLevel(lvl)
	return nil
}

type Config struct {
	BindAddr            string    `env:"pluralkit__status__addr" envDefault:"0.0.0.0:8080"`
	ShardsEndpoint      string    `env:"pluralkit__status__shards_endpoint" envDefault:"https://api.pluralkit.me/private/discord/shard_state"`
	AuthToken           string    `env:"pluralkit__status__auth_token"`
	NotificationWebhook string    `env:"pluralkit__status__notification_webhook"`
	NotificationRole    string    `env:"pluralkit__status__notification_role"`
	RunDev              bool      `env:"pluralkit__status__run_dev" envDefault:"false"`
	DBLoc               string    `env:"pluralkit__status__db_location" envDefault:"file:status.db?_foreign_keys=on"`
	LogLevel            SlogLevel `env:"pluralkit__consoleloglevel" envDefault:"info"`
}
