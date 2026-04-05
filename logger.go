package logger

import (
	"io"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
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
}

const defaultEnvName = "LOG_LEVEL"

var global = logger{
	writerList: []io.Writer{os.Stdout},
	envName:    defaultEnvName,
}

func WithAttr(attrs ...slog.Attr) {
	global.mx.Lock()
	defer global.mx.Unlock()

	for _, attr := range attrs {
		global.attrList = append(global.attrList, attr)
	}

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

	handler := slog.NewJSONHandler(newFanoutWriter(global.writerList...), &slog.HandlerOptions{
		Level:     global.level,
		AddSource: global.withSource,
	})

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

func parseLevel(level string) (parsed slog.Level) {
	err := parsed.UnmarshalText([]byte(level))
	if err != nil {
		return slog.LevelInfo
	}

	return parsed
}
