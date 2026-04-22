package logger_test

import (
	"log/slog"
	"os"

	"github.com/nextlag/logger"
)

func Example() {
	logger.SetLevel("DEBUG")
	logger.WithJSON(false)

	log := logger.GetInstance()
	log.Info("server started", "port", 8080)
}

func Example_json() {
	logger.SetLevel("INFO")
	logger.WithJSON(true)

	log := logger.GetInstance()
	log.Info("request handled", "method", "GET", "status", 200)
}

func Example_withAttr() {
	logger.SetLevel("INFO")
	logger.WithAttr(slog.String("service", "api"))

	log := logger.GetInstance()
	log.Info("ready")
}

func Example_addWriter() {
	f, err := os.CreateTemp("", "log-*.jsonl")
	if err != nil {
		panic(err)
	}

	defer func() { _ = os.Remove(f.Name()) }()
	defer func() { _ = f.Close() }()

	logger.AddWriter(f)

	log := logger.GetInstance()
	log.Info("writes to stdout and file")
}
