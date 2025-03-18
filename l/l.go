package l

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"golang.org/x/term"

	"github.com/nextlag/logger/config"
)

const fileRule = 0644

// ErrNilLoggerConfig представляет ошибку отсутствия конфигурации логгера.
var ErrNilLoggerConfig = errors.New("nil logger config")

// colorEnabled определяет, поддерживает ли терминал цветной вывод.
//
//nolint:gochecknoglobals
var colorEnabled = term.IsTerminal(int(os.Stdout.Fd()))

type LoggerOptions struct {
	Attrs     []slog.Attr
	SlogOpts  *slog.HandlerOptions
	Out       io.Writer
	log       *log.Logger
	LogToFile bool
	LogPath   string
}

type logHandler struct {
	opts *LoggerOptions
	slog.Handler
}

// NewLogger создает новый экземпляр логгера.
func NewLogger(cfg *config.Config) (*slog.Logger, error) {
	if cfg == nil || cfg.Logging == nil {
		return nil, ErrNilLoggerConfig
	}

	opts := LoggerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level:     cfg.Logging.Level,
			AddSource: true,
		},
		LogToFile: cfg.Logging.LogToFile,
		LogPath:   cfg.Logging.LogPath,
	}

	if opts.LogToFile {
		file, err := os.OpenFile(opts.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileRule)
		if err != nil {
			log.Fatalf("failed to open log file: %v", err)
		}

		opts.Out = file
	} else {
		opts.Out = os.Stdout
	}

	output := colorable.NewColorable(os.Stdout)
	handler := newLogHandler(output, &opts)

	return slog.New(handler), nil
}

func newLogHandler(out io.Writer, opts *LoggerOptions) *logHandler {
	baseHandler := slog.NewTextHandler(out, opts.SlogOpts)

	return &logHandler{
		Handler: baseHandler,
		opts: &LoggerOptions{
			log: log.New(out, "", 0),
		},
	}
}

func (lh *logHandler) Handle(_ context.Context, record slog.Record) error {
	level := record.Level.String() + ":"
	message := record.Message

	var fields []slog.Attr

	record.Attrs(func(attr slog.Attr) bool {
		fields = append(fields, attr)

		return true
	})

	formattedFields := formatFields(fields)

	fs := runtime.CallersFrames([]uintptr{record.PC})
	frame, _ := fs.Next()
	relPath := shortenFilePath(frame.File)
	callerInfo := fmt.Sprintf("%s:%d", relPath, frame.Line)

	timeStr := record.Time.Format("[02.01.2006 15:04:05.000]")

	if colorEnabled {
		level = applyLevelColor(record.Level, level)
		message = color.CyanString(message)
		callerInfo = color.WhiteString(callerInfo)
	}

	parts := []string{
		timeStr,
		level,
		message,
		callerInfo,
	}

	if formattedFields != "" {
		parts = append(parts, formattedFields)
	}

	lh.opts.log.Println(strings.Join(parts, " "))

	return nil
}

func formatFields(attrs []slog.Attr) string {
	var fields []string

	var specialFields []string

	for _, attr := range attrs {
		if attr.Key == "record_id" || attr.Key == "field" {
			specialFields = append(specialFields, formatAttr(attr))
		}
	}

	for _, attr := range attrs {
		if attr.Key != "record_id" && attr.Key != "field" {
			fields = append(fields, formatAttr(attr))
		}
	}

	return strings.Join(append(specialFields, fields...), " ")
}

func formatAttr(attr slog.Attr) string {
	switch value := attr.Value.Any().(type) {
	case string:
		return fmt.Sprintf("%s=%q", attr.Key, value)
	case []byte:
		return fmt.Sprintf("%s=%q", attr.Key, string(value))
	case *[]byte:
		return fmt.Sprintf("%s=%q", attr.Key, string(*value))
	case error:
		return fmt.Sprintf("%s=%q", attr.Key, value.Error())
	default:
		return fmt.Sprintf("%s=%v", attr.Key, attr.Value)
	}
}

func applyLevelColor(level slog.Level, str string) string {
	if !colorEnabled {
		return str
	}

	switch {
	case level >= slog.LevelError:
		return color.RedString(str)
	case level >= slog.LevelWarn:
		return color.MagentaString(str)
	case level >= slog.LevelInfo:
		return color.GreenString(str)
	default:
		return color.YellowString(str)
	}
}

func shortenFilePath(path string) string {
	if relPath, err := filepath.Rel(getProjectRoot(), path); err == nil {
		return relPath
	}

	return filepath.Base(path)
}

func getProjectRoot() string {
	_, filePath, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}

	dir := filepath.Dir(filePath)

	return filepath.Join(dir, "..", "..")
}

func (lh *logHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &logHandler{
		Handler: lh.Handler.WithAttrs(attrs),
		opts:    lh.opts,
	}
}

func (lh *logHandler) WithGroup(name string) slog.Handler {
	return &logHandler{
		Handler: lh.Handler.WithGroup(name),
		opts:    lh.opts,
	}
}
