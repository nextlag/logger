package logger

import (
	"runtime"
	"strings"
)

func sourceFromPC(pc uintptr) (string, int) {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "???", 0
	}

	file, line := fn.FileLine(pc)

	return strings.TrimPrefix(file, sourcePrefix), line
}
