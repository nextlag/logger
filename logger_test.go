package logger

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

func resetGlobal(t *testing.T) {
	t.Helper()

	global.mx.Lock()
	defer global.mx.Unlock()

	global.writerList = nil
	global.attrList = nil
	global.serviceName = ""
	global.envName = defaultEnvName
	global.level = nil
	global.withSource = false
	global.instance.Store(nil)
}

func TestParseLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  slog.Level
	}{
		{"debug lowercase", "debug", slog.LevelDebug},
		{"DEBUG uppercase", "DEBUG", slog.LevelDebug},
		{"info", "INFO", slog.LevelInfo},
		{"warn", "WARN", slog.LevelWarn},
		{"error", "ERROR", slog.LevelError},
		{"empty string defaults to info", "", slog.LevelInfo},
		{"invalid string defaults to info", "invalid", slog.LevelInfo},
		{"numeric offset DEBUG+2", "DEBUG+2", slog.LevelDebug + 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()

			got := parseLevel(test.input)
			if got != test.want {
				tt.Errorf("parseLevel(%q) = %v, want %v", test.input, got, test.want)
			}
		})
	}
}

func TestGetInstance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func(buf *bytes.Buffer)
		logFunc     func(log *slog.Logger)
		wantFields  map[string]any
		wantContain string
	}{
		{
			name: "default level info writes info message",
			setup: func(buf *bytes.Buffer) {
				SetLevel("INFO")
				AddWriter(buf)
			},
			logFunc:     func(log *slog.Logger) { log.Info("test") },
			wantContain: "test",
		},
		{
			name: "debug level filters debug message",
			setup: func(buf *bytes.Buffer) {
				SetLevel("ERROR")
				AddWriter(buf)
			},
			logFunc: func(log *slog.Logger) { log.Debug("hidden") },
		},
		{
			name: "service name creates group",
			setup: func(buf *bytes.Buffer) {
				SetLevel("INFO")
				SetServiceName("myapp")
				AddWriter(buf)
			},
			logFunc: func(log *slog.Logger) { log.Info("test", "key", "value") },
			wantFields: map[string]any{"myapp": map[string]any{
				"key": "value",
			}},
		},
		{
			name: "with attrs adds fields",
			setup: func(buf *bytes.Buffer) {
				SetLevel("INFO")
				WithAttr(slog.String("version", "1.0"))
				AddWriter(buf)
			},
			logFunc:    func(log *slog.Logger) { log.Info("test") },
			wantFields: map[string]any{"version": "1.0"},
		},
		{
			name: "with source adds source field",
			setup: func(buf *bytes.Buffer) {
				SetLevel("INFO")
				WithSource(true)
				AddWriter(buf)
			},
			logFunc:     func(log *slog.Logger) { log.Info("test") },
			wantContain: "source",
		},
		{
			name: "returns cached instance",
			setup: func(buf *bytes.Buffer) {
				SetLevel("INFO")
				AddWriter(buf)
			},
			logFunc: func(_ *slog.Logger) {
				first := GetInstance()
				second := GetInstance()
				if first != second {
					panic("instances differ")
				}
			},
		},
		{
			name: "config change resets instance",
			setup: func(buf *bytes.Buffer) {
				SetLevel("INFO")
				AddWriter(buf)
				_ = GetInstance()
				SetLevel("ERROR")
			},
			logFunc: func(log *slog.Logger) { log.Info("should be hidden") },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			resetGlobal(tt)

			var buf bytes.Buffer

			test.setup(&buf)

			log := GetInstance()

			test.logFunc(log)

			output := buf.String()

			if test.wantContain != "" && !strings.Contains(output, test.wantContain) {
				tt.Errorf("output %q does not contain %q", output, test.wantContain)
			}

			if test.wantFields != nil && output != "" {
				var parsed map[string]any

				if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsed); err != nil {
					tt.Fatalf("failed to parse JSON output: %v", err)
				}

				for key, want := range test.wantFields {
					got, ok := parsed[key]
					if !ok {
						tt.Errorf("missing field %q in output", key)

						continue
					}

					switch w := want.(type) {
					case string:
						if got != w {
							tt.Errorf("field %q = %v, want %v", key, got, w)
						}
					case map[string]any:
						_, ok = got.(map[string]any)
						if !ok {
							tt.Errorf("field %q is not an object", key)
						}
					}
				}
			}

			if test.wantContain == "" && test.wantFields == nil && output != "" {
				tt.Errorf("expected no output, got %q", output)
			}
		})
	}
}
