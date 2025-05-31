package util

import (
	"errors"
	"log/slog"
)

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
	BindAddr string    `env:"pluralkit__status__addr" envDefault:"0.0.0.0:8080"`
	RunDev   bool      `env:"pluralkit__status__run_dev" envDefault:"false"`
	DBLoc    string    `env:"pluralkit__status__db_location" envDefault:"file:status.db"`
	LogLevel SlogLevel `env:"pluralkit__consoleloglevel" envDefault:"info"`
}
