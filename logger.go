package logger

import (
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type logger struct {
	mx          sync.Mutex
	writerList  []io.Writer
	attrList    []any
	serviceName string
	envName     string
	level       *slog.Level
	instance    atomic.Pointer[slog.Logger]
	withSource  bool
	useJSON     bool
	handler     slog.Handler
}

const defaultEnvName = "LOG_LEVEL"

var (
	global = logger{
		writerList: []io.Writer{os.Stdout},
		envName:    defaultEnvName,
		useJSON:    true,
	}

	sourcePrefix = func() string {
		wd, err := os.Getwd()
		if err != nil {
			return ""
		}

		return wd + "/"
	}()
)

func WithAttr(attrs ...slog.Attr) {
	global.mx.Lock()
	defer global.mx.Unlock()

	for _, attr := range attrs {
		global.attrList = append(global.attrList, attr)
	}

	global.instance.Store(nil)
}

func WithJSON(state bool) {
	global.mx.Lock()
	defer global.mx.Unlock()

	global.useJSON = state
	global.handler = nil
	global.instance.Store(nil)
}

func WithHandler(h slog.Handler) {
	global.mx.Lock()
	defer global.mx.Unlock()

	global.handler = h
	global.instance.Store(nil)
}

func WithSource(state bool) {
	global.mx.Lock()
	defer global.mx.Unlock()

	global.withSource = state
	global.instance.Store(nil)
}

func SetServiceName(serviceName string) {
	global.mx.Lock()
	defer global.mx.Unlock()

	global.serviceName = serviceName
	global.instance.Store(nil)
}

func SetLevel(logLevel string) {
	global.mx.Lock()
	defer global.mx.Unlock()

	global.level = new(parseLevel(logLevel))
	global.instance.Store(nil)
}

func AddWriter(writer io.Writer) {
	global.mx.Lock()
	defer global.mx.Unlock()

	global.writerList = append(global.writerList, writer)
	global.instance.Store(nil)
}

func GetInstance() *slog.Logger {
	inst := global.instance.Load()
	if inst != nil {
		return inst
	}

	global.mx.Lock()
	defer global.mx.Unlock()

	inst = global.instance.Load()
	if inst != nil {
		return inst
	}

	if global.level == nil {
		global.level = new(parseLevel(os.Getenv(global.envName)))
	}

	var handler slog.Handler
	switch {
	case global.handler != nil:
		handler = global.handler
	case global.useJSON:
		handler = slog.NewJSONHandler(newFanoutWriter(global.writerList...), &slog.HandlerOptions{
			Level:       global.level,
			AddSource:   global.withSource,
			ReplaceAttr: sourceReplacer(),
		})
	default:
		handler = newTextHandler(newFanoutWriter(global.writerList...), global.level, global.withSource)
	}

	inst = slog.New(handler)

	if global.serviceName != "" {
		inst = inst.WithGroup(global.serviceName)
	}

	if len(global.attrList) > 0 {
		inst = inst.With(global.attrList...)
	}

	global.instance.Store(inst)

	return inst
}

func sourceReplacer() func([]string, slog.Attr) slog.Attr {
	if sourcePrefix == "" {
		return nil
	}

	return func(_ []string, a slog.Attr) slog.Attr {
		switch a.Key {
		case slog.TimeKey:
			t, ok := a.Value.Any().(time.Time)
			if ok {
				return slog.String(slog.TimeKey, t.Format("2006-01-02T15:04:05"))
			}

		case slog.SourceKey:
			s, ok := a.Value.Any().(*slog.Source)
			if !ok {
				return a
			}

			file := strings.TrimPrefix(s.File, sourcePrefix)

			buf := make([]byte, 0, len(file)+5)
			buf = append(buf, file...)
			buf = append(buf, ':')
			buf = strconv.AppendInt(buf, int64(s.Line), 10)

			return slog.String("file", string(buf))
		}

		return a
	}
}

func parseLevel(level string) (parsed slog.Level) {
	err := parsed.UnmarshalText([]byte(level))
	if err != nil {
		return slog.LevelInfo
	}

	return parsed
}
