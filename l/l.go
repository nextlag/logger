// nolint
package l

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-colorable"
	"golang.org/x/term"

	"github.com/nextlag/logger/config"
)

var (
	colorEnabled = term.IsTerminal(int(os.Stdout.Fd()))
	pool         = sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, 1024)
		},
	}
)

type LoggerOptions struct {
	Attrs     []slog.Attr
	SlogOpts  *slog.HandlerOptions
	LogToFile bool
	LogPath   string
	Out       io.Writer
	log       *log.Logger
}
type logHandler struct {
	opts *LoggerOptions
	slog.Handler
}

func NewLogger(cfg *config.Config) (*slog.Logger, error) {
	if cfg == nil || cfg.Logging == nil {
		return nil, errors.New("nil logger configuration")
	}

	opts := LoggerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level:     cfg.Logging.Level,
			AddSource: true,
		},

		LogToFile: cfg.Logging.LogToFile,
		LogPath:   cfg.Logging.LogPath,
	}

	output := colorable.NewColorable(os.Stdout)
	if opts.LogToFile {
		file, err := os.OpenFile(opts.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		output = file
	}

	handler := newLogHandler(output, &opts)

	return slog.New(handler), nil
}

func newLogHandler(out io.Writer, opts *LoggerOptions) *logHandler {
	var baseHandler slog.Handler

	if opts.LogToFile {
		baseHandler = slog.NewJSONHandler(out, opts.SlogOpts)
	}

	baseHandler = slog.NewTextHandler(out, opts.SlogOpts)

	return &logHandler{
		Handler: baseHandler,
		opts: &LoggerOptions{
			LogPath:   opts.LogPath,
			LogToFile: opts.LogToFile,
			log:       log.New(out, "", 0),
		},
	}
}

func (lh *logHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"
	msg := r.Message

	fields := make(map[string]interface{}, r.NumAttrs())
	r.Attrs(func(a slog.Attr) bool {
		processAttr(a, fields)

		return true
	})

	buf := pool.Get().([]byte)
	defer func() {
		buf = buf[:0]
		pool.Put(buf)
	}()

	var fieldsJSON []byte

	if len(fields) > 0 {
		var err error
		if fieldsJSON, err = json.Marshal(fields); err != nil {
			return fmt.Errorf("failed to marshal log fields: %w", err)
		}
	}

	fs := runtime.CallersFrames([]uintptr{r.PC})
	frame, _ := fs.Next()
	relPath := lh.shortenFilePath(frame.File)

	timeStr := r.Time.Format("[02.01.2006 15:04:05.000]")

	callerInfo := fmt.Sprintf("%s:%d", relPath, frame.Line)

	if !lh.opts.LogToFile && colorEnabled {
		level = applyLevelColor(r.Level, level)
		msg = color.CyanString(msg)
		callerInfo = color.WhiteString(callerInfo)
	}

	logLine := fmt.Sprintf("%s %s %s %s %s",
		timeStr,
		level,
		msg,
		callerInfo,
		string(fieldsJSON),
	)

	lh.opts.log.Println(logLine)

	return nil
}

func processAttr(a slog.Attr, fields map[string]interface{}) {
	switch v := a.Value.Any().(type) {
	case []byte:
		fields[a.Key] = string(v)
	case *[]byte:
		fields[a.Key] = string(*v)
	case error:
		fields[a.Key] = v.Error()
	default:
		fields[a.Key] = a.Value.Any()
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

func (lh *logHandler) shortenFilePath(path string) string {
	if relPath, err := filepath.Rel(lh.getProjectRoot(), path); err == nil {
		return relPath
	}

	return filepath.Base(path)
}

func (lh *logHandler) getProjectRoot() string {
	if lh.opts.LogToFile {
		_, file, _, _ := runtime.Caller(0)
		return filepath.Join(filepath.Dir(file), "..", "..")
	}

	_, file, _, _ := runtime.Caller(8)

	return filepath.Dir(file)
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

func Float32Attr(key string, val float32) slog.Attr {
	return slog.Float64(key, float64(val))
}

func UInt32Attr(key string, val uint32) slog.Attr {
	return slog.Int(key, int(val))
}

func Int32Attr(key string, val int32) slog.Attr {
	return slog.Int(key, int(val))
}

func TimeAttr(key string, time time.Time) slog.Attr {
	return slog.String(key, time.String())
}

func ErrAttr(err error) slog.Attr {
	return slog.String("error", err.Error())
}
