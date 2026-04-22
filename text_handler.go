package logger

import (
	"context"
	"io"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"
)

type textHandler struct {
	w         io.Writer
	mu        *sync.Mutex
	level     slog.Leveler
	addSource bool
	attrs     []slog.Attr
	group     string
}

func newTextHandler(w io.Writer, level slog.Leveler, addSource bool) *textHandler {
	return &textHandler{
		w:         w,
		mu:        &sync.Mutex{},
		level:     level,
		addSource: addSource,
	}
}

func (h *textHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *textHandler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder
	b.Grow(128)

	b.WriteString(r.Time.Format(time.DateTime))
	b.WriteByte(' ')

	level := r.Level.String()
	b.WriteString(level)

	for i := len(level); i < 5; i++ {
		b.WriteByte(' ')
	}

	if h.addSource && r.PC != 0 {
		file, line := sourceFromPC(r.PC)
		b.WriteByte(' ')
		b.WriteString(file)
		b.WriteByte(':')
		b.Write(strconv.AppendInt(nil, int64(line), 10))
	}

	b.WriteString(" \"")
	b.WriteString(r.Message)
	b.WriteByte('"')

	for _, attr := range h.attrs {
		appendAttr(&b, h.group, attr)
	}

	r.Attrs(func(a slog.Attr) bool {
		appendAttr(&b, h.group, a)

		return true
	})

	b.WriteByte('\n')

	h.mu.Lock()
	defer h.mu.Unlock()

	_, err := io.WriteString(h.w, b.String())

	return err
}

func (h *textHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &textHandler{
		w:         h.w,
		mu:        h.mu,
		level:     h.level,
		addSource: h.addSource,
		attrs:     append(h.attrs[:len(h.attrs):len(h.attrs)], attrs...),
		group:     h.group,
	}
}

func (h *textHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	prefix := name
	if h.group != "" {
		prefix = h.group + "." + name
	}

	return &textHandler{
		w:         h.w,
		mu:        h.mu,
		level:     h.level,
		addSource: h.addSource,
		attrs:     h.attrs,
		group:     prefix,
	}
}

func appendAttr(b *strings.Builder, group string, a slog.Attr) {
	a.Value = a.Value.Resolve()

	if a.Equal(slog.Attr{}) {
		return
	}

	if a.Value.Kind() == slog.KindGroup {
		prefix := a.Key
		if group != "" {
			prefix = group + "." + a.Key
		}

		for _, ga := range a.Value.Group() {
			appendAttr(b, prefix, ga)
		}

		return
	}

	b.WriteByte(' ')

	if group != "" {
		b.WriteString(group)
		b.WriteByte('.')
	}

	b.WriteString(a.Key)
	b.WriteByte('=')
	b.WriteString(a.Value.String())
}
