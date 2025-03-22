package config

import (
	"errors"
	"log/slog"

	"github.com/nextlag/logger/er"
)

// LogLevelValue реализует интерфейс flag.Value.
type LogLevelValue struct {
	Value *slog.Level
}

func (l *LogLevelValue) String() string {
	if l.Value == nil {
		return ""
	}

	return l.Value.String()
}

func (l *LogLevelValue) Set(value string) error {
	logLevelMap := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	level, found := logLevelMap[value]
	if !found {
		return errors.Join(er.Format("invalid log level: %s", value))
	}

	*l.Value = level

	return nil
}
