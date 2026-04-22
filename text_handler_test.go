package logger

import (
	"bytes"
	"context"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"testing/slogtest"
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
		{
			name: "inline group inside named group",
			handler: func(buf *bytes.Buffer) slog.Handler {
				return newTextHandler(buf, slog.LevelDebug, false).WithGroup("G")
			},
			record: func() slog.Record {
				r := newTestRecord(slog.LevelInfo, "test")
				r.AddAttrs(slog.Group("", slog.String("a", "1")))
				return r
			},
			wantParts: []string{"G.a=1"},
		},
		{
			name:    "value with spaces is quoted",
			handler: func(buf *bytes.Buffer) slog.Handler { return newTextHandler(buf, slog.LevelDebug, false) },
			record: func() slog.Record {
				r := newTestRecord(slog.LevelInfo, "test")
				r.AddAttrs(slog.String("msg", "hello world"))
				return r
			},
			wantParts: []string{`msg="hello world"`},
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

func TestTextHandlerSlogtest(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	slogtest.Run(t, func(_ *testing.T) slog.Handler {
		buf.Reset()
		return newTextHandler(&buf, slog.LevelDebug, false)
	}, func(_ *testing.T) map[string]any {
		return parseTextOutput(buf.String())
	})
}

// parseTextOutput parses text handler output into a map for slogtest.
// Format: `2006-01-02 15:04:05 LEVEL "msg" key=value group.key=value`
func parseTextOutput(line string) map[string]any {
	m := make(map[string]any)

	line = strings.TrimSpace(line)
	if line == "" {
		return m
	}

	// strip ANSI escape codes
	line = regexp.MustCompile(`\033\[[0-9;]*m`).ReplaceAllString(line, "")

	// time: first 19 chars "2006-01-02 15:04:05"
	if len(line) >= 19 {
		m[slog.TimeKey] = line[:19]
		line = line[19:]
	}

	line = strings.TrimLeft(line, " ")

	// level: next word
	if idx := strings.IndexByte(line, ' '); idx > 0 {
		m[slog.LevelKey] = strings.TrimSpace(line[:idx])
		line = strings.TrimLeft(line[idx:], " ")
	}

	// message: quoted string
	if len(line) > 0 && line[0] == '"' {
		end := strings.Index(line[1:], "\"")
		if end >= 0 {
			m[slog.MessageKey] = line[1 : end+1]
			line = strings.TrimLeft(line[end+2:], " ")
		}
	}

	// attributes: key=value pairs, dots denote nested groups
	for _, pair := range splitAttrs(line) {
		key, val, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}

		if unq, err := strconv.Unquote(val); err == nil {
			val = unq
		}

		setNested(m, strings.Split(key, "."), val)
	}

	return m
}

// splitAttrs splits "k1=v1 k2=\"v 2\" k3=v3" respecting quoted values.
func splitAttrs(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	var parts []string
	for len(s) > 0 {
		eqIdx := strings.IndexByte(s, '=')
		if eqIdx < 0 {
			break
		}

		rest := s[eqIdx+1:]
		var val string
		if len(rest) > 0 && rest[0] == '"' {
			end, err := strconv.Unquote(rest)
			if err == nil {
				quoted := rest[:len(rest)-len(strings.TrimPrefix(rest[1:], ""))+1]

				for i := 1; i < len(rest); i++ {
					if rest[i] == '\\' {
						i++
						continue
					}

					if rest[i] == '"' {
						val = rest[:i+1]
						_ = end
						break
					}
				}

				_ = quoted
			}
		}

		if val == "" {
			spIdx := strings.IndexByte(rest, ' ')
			if spIdx < 0 {
				parts = append(parts, s)
				break
			}

			parts = append(parts, s[:eqIdx+1+spIdx])
			s = strings.TrimLeft(rest[spIdx:], " ")

			continue
		}

		parts = append(parts, s[:eqIdx+1]+val)
		after := rest[len(val):]
		s = strings.TrimLeft(after, " ")
	}

	return parts
}

// setNested sets a value in a nested map: ["a","b","k"] → m["a"]["b"]["k"] = val.
func setNested(m map[string]any, keys []string, val any) {
	for i, k := range keys {
		if i == len(keys)-1 {
			m[k] = val
			return
		}

		sub, ok := m[k].(map[string]any)
		if !ok {
			sub = make(map[string]any)
			m[k] = sub
		}

		m = sub
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
