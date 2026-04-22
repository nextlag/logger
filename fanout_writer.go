package logger

import (
	"errors"
	"io"
)

type fanoutWriter struct {
	list []io.Writer
}

func newFanoutWriter(list ...io.Writer) *fanoutWriter {
	return &fanoutWriter{list: list}
}

func (fw *fanoutWriter) Write(bs []byte) (int, error) {
	if len(fw.list) == 1 {
		return fw.list[0].Write(bs)
	}

	var errs []error

	for _, writer := range fw.list {
		_, err := writer.Write(bs)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return 0, errors.Join(errs...)
	}

	return len(bs), nil
}
