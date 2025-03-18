package l

import (
	"log/slog"
	"time"
)

type Logger = slog.Logger

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
