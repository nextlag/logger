package logger

import (
	"runtime"
	"strings"
)

func sourceFromPC(pc uintptr) (string, int) {
	frames := runtime.CallersFrames([]uintptr{pc})
	f, _ := frames.Next()
	if f.File == "" {
		return "???", 0
	}

	return strings.TrimPrefix(f.File, sourcePrefix), f.Line
}
