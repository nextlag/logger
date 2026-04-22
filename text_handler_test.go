package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func newTestRecord(level slog.Level, msg string) slog.Record {
	return slog.NewRecord(time.Date(2025, 1, 15, 12, 30, 0, 0, time.UTC), level, msg, 0)
}

func TestTextHandlerEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		handler   slog.Level
		check     slog.Level
		wantAllow bool
	}{
		{"info enabled at debug", slog.LevelDebug, slog.LevelInfo, true},
		{"debug disabled at warn", slog.LevelWarn, slog.LevelDebug, false},
		{"warn enabled at warn", slog.LevelWarn, slog.LevelWarn, true},
		{"error enabled at warn", slog.LevelWarn, slog.LevelError, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()

			h := newTextHandler(&bytes.Buffer{}, test.handler, false)
			got := h.Enabled(context.Background(), test.check)

			if got != test.wantAllow {
				tt.Errorf("Enabled(%v) = %v, want %v", test.check, got, test.wantAllow)
			}
		})
	}
}

func TestTextHandlerHandle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		handler   func(*bytes.Buffer) slog.Handler
		record    func() slog.Record
		wantParts []string
	}{
		{
			name:    "info with green color",
			handler: func(buf *bytes.Buffer) slog.Handler { return newTextHandler(buf, slog.LevelDebug, false) },
			record:  func() slog.Record { return newTestRecord(slog.LevelInfo, "hello") },
			wantParts: []string{
				"2025-01-15 12:30:00",
				"\033[32mINFO\033[0m",
				`"hello"`,
			},
		},
		{
			name:    "error with red color",
			handler: func(buf *bytes.Buffer) slog.Handler { return newTextHandler(buf, slog.LevelDebug, false) },
			record:  func() slog.Record { return newTestRecord(slog.LevelError, "fail") },
			wantParts: []string{
				"\033[31mERROR\033[0m",
				`"fail"`,
			},
		},
		{
			name:    "warn with magenta color",
			handler: func(buf *bytes.Buffer) slog.Handler { return newTextHandler(buf, slog.LevelDebug, false) },
			record:  func() slog.Record { return newTestRecord(slog.LevelWarn, "caution") },
			wantParts: []string{
				"\033[35mWARN\033[0m",
				`"caution"`,
			},
		},
		{
			name:    "debug with gray color",
			handler: func(buf *bytes.Buffer) slog.Handler { return newTextHandler(buf, slog.LevelDebug, false) },
			record:  func() slog.Record { return newTestRecord(slog.LevelDebug, "trace") },
			wantParts: []string{
				"\033[90mDEBUG\033[0m",
				`"trace"`,
			},
		},
		{
			name: "with attrs",
			handler: func(buf *bytes.Buffer) slog.Handler {
				return newTextHandler(buf, slog.LevelDebug, false).
					WithAttrs([]slog.Attr{slog.String("key", "value")})
			},
			record:    func() slog.Record { return newTestRecord(slog.LevelInfo, "test") },
			wantParts: []string{"key=value"},
		},
		{
			name: "with group",
			handler: func(buf *bytes.Buffer) slog.Handler {
				return newTextHandler(buf, slog.LevelDebug, false).
					WithGroup("svc").
					WithAttrs([]slog.Attr{slog.String("key", "value")})
			},
			record:    func() slog.Record { return newTestRecord(slog.LevelInfo, "test") },
			wantParts: []string{"svc.key=value"},
		},
		{
			name: "nested groups",
			handler: func(buf *bytes.Buffer) slog.Handler {
				return newTextHandler(buf, slog.LevelDebug, false).
					WithGroup("a").WithGroup("b").
					WithAttrs([]slog.Attr{slog.String("k", "v")})
			},
			record:    func() slog.Record { return newTestRecord(slog.LevelInfo, "test") },
			wantParts: []string{"a.b.k=v"},
		},
		{
			name:    "record attrs",
			handler: func(buf *bytes.Buffer) slog.Handler { return newTextHandler(buf, slog.LevelDebug, false) },
			record: func() slog.Record {
				r := newTestRecord(slog.LevelInfo, "test")
				r.AddAttrs(slog.String("rkey", "rval"))
				return r
			},
			wantParts: []string{"rkey=rval"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()

			var buf bytes.Buffer
			h := test.handler(&buf)

			r := test.record()
			if err := h.Handle(context.Background(), r); err != nil {
				tt.Fatalf("Handle() error: %v", err)
			}

			output := buf.String()
			for _, part := range test.wantParts {
				if !strings.Contains(output, part) {
					tt.Errorf("output %q missing %q", output, part)
				}
			}
		})
	}
}

func TestTextHandlerEmptyGroup(t *testing.T) {
	t.Parallel()

	h := newTextHandler(&bytes.Buffer{}, slog.LevelDebug, false)
	h2 := h.WithGroup("")

	if h != h2 {
		t.Error("WithGroup(\"\") should return the same handler")
	}
}
