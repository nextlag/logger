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
	var err error

	for _, writer := range fw.list {
		_, wrErr := writer.Write(bs)
		if wrErr != nil {
			err = errors.Join(err, wrErr)
		}
	}

	if err != nil {
		return 0, err
	}

	return len(bs), nil
}
