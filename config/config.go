package config

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"

	"github.com/nextlag/logger/er"
)

type (
	Config struct {
		Logging *Logging `json:"logging"`
	}

	Logging struct {
		Level     slog.Level `json:"level"`
		LogToFile bool       `json:"log_to_file"`
		LogPath   string     `json:"log_path"`
	}
)

func Load(reader io.Reader) (*Config, error) {
	if reader == nil {
		return nil, errors.Join(er.ErrIncorrectReader, nil)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Join(er.ErrReadBuffer, nil)
	}

	var cfg Config
	if err = json.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Join(er.ErrConfigParse, err)
	}

	return &cfg, nil
}
