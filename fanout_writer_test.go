package logger

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

type errWriter struct {
	err error
}

func (ew *errWriter) Write([]byte) (int, error) {
	return 0, ew.err
}

func TestFanoutWriterWrite(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		writers func() []io.Writer
		data    string
		wantN   int
		wantErr bool
		check   func(t *testing.T, writers []io.Writer)
	}{
		{
			name:    "single writer",
			writers: func() []io.Writer { return []io.Writer{&bytes.Buffer{}} },
			data:    "hello",
			wantN:   5,
			check: func(t *testing.T, writers []io.Writer) {
				if got := writers[0].(*bytes.Buffer).String(); got != "hello" {
					t.Errorf("writer got %q, want %q", got, "hello")
				}
			},
		},
		{
			name: "multiple writers receive same data",
			writers: func() []io.Writer {
				return []io.Writer{&bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}}
			},
			data:  "hello",
			wantN: 5,
			check: func(t *testing.T, writers []io.Writer) {
				for i, writer := range writers {
					if got := writer.(*bytes.Buffer).String(); got != "hello" {
						t.Errorf("writer[%d] got %q, want %q", i, got, "hello")
					}
				}
			},
		},
		{
			name:    "empty data",
			writers: func() []io.Writer { return []io.Writer{&bytes.Buffer{}} },
			data:    "",
			wantN:   0,
		},
		{
			name:    "no writers",
			writers: func() []io.Writer { return nil },
			data:    "hello",
			wantN:   5,
		},
		{
			name: "error writer",
			writers: func() []io.Writer {
				return []io.Writer{&errWriter{err: errors.New("fail")}}
			},
			data:    "hello",
			wantN:   0,
			wantErr: true,
		},
		{
			name: "mixed ok and error returns error",
			writers: func() []io.Writer {
				return []io.Writer{&bytes.Buffer{}, &errWriter{err: errors.New("fail")}}
			},
			data:    "hello",
			wantN:   0,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()

			writers := test.writers()
			fw := newFanoutWriter(writers...)

			n, err := fw.Write([]byte(test.data))

			if (err != nil) != test.wantErr {
				tt.Errorf("Write() error = %v, wantErr %v", err, test.wantErr)
			}

			if n != test.wantN {
				tt.Errorf("Write() n = %d, want %d", n, test.wantN)
			}

			if test.check != nil {
				test.check(tt, writers)
			}
		})
	}
}
