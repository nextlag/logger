package logger

import "log/slog"

var (
	coloredDebug = "\033[90mDEBUG\033[0m"
	coloredInfo  = "\033[32mINFO\033[0m"
	coloredWarn  = "\033[35mWARN\033[0m"
	coloredError = "\033[31mERROR\033[0m"
)

func colorLevel(level slog.Level) string {
	switch level {
	case slog.LevelDebug:
		return coloredDebug
	case slog.LevelInfo:
		return coloredInfo
	case slog.LevelWarn:
		return coloredWarn
	case slog.LevelError:
		return coloredError
	default:
		return level.String()
	}
}
