package main

import (
	"bytes"
	"log"
	"os"

	"github.com/nextlag/logger/config"
	"github.com/nextlag/logger/l"
)

const configFile = "config.json"

func main() {
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatal("failed to read file config")
	}

	cfg, err := config.Load(bytes.NewReader(data))
	if err != nil {
		log.Fatal("failed to load config")
	}

	logger, err := l.NewLogger(cfg)
	if err != nil {
		log.Fatal("failed to initialize logger")
	}

	logger.Info("Config", "logging", cfg.Logging)
}
