package log

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestNewLogger(t *testing.T) {
	l := New(false)
	if l == nil {
		t.Fatal("logger should not be nil")
	}
}

func TestDebugDisabledByDefault(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	l.Debug("hidden")
	if strings.Contains(buf.String(), "hidden") {
		t.Error("debug messages should be hidden when debug is off")
	}
}

func TestDebugEnabled(t *testing.T) {
	var buf bytes.Buffer
	l := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	l.Debug("visible")
	if !strings.Contains(buf.String(), "visible") {
		t.Error("debug messages should be visible when debug is on")
	}
}
