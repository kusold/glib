package log

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
)

func TestNewDefaultsToTextInfoLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{
		Writer:      &buf,
		ReplaceAttr: dropTime,
	})

	logger.Debug("hidden")
	logger.Info("visible", Component("api"))

	got := buf.String()
	if strings.Contains(got, "hidden") {
		t.Fatalf("expected debug log to be filtered, got %q", got)
	}
	if !strings.Contains(got, `level=INFO msg=visible component=api`) {
		t.Fatalf("expected text info log, got %q", got)
	}
}

func TestNewSupportsJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{
		Writer:      &buf,
		Format:      FormatJSON,
		ReplaceAttr: dropTime,
	})

	logger.Info("created", RequestID("req-123"))

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal json log: %v", err)
	}
	if got["level"] != "INFO" {
		t.Fatalf("expected level INFO, got %v", got["level"])
	}
	if got["msg"] != "created" {
		t.Fatalf("expected msg created, got %v", got["msg"])
	}
	if got["request_id"] != "req-123" {
		t.Fatalf("expected request id, got %v", got["request_id"])
	}
}

func TestNewUsesLevelOption(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{
		Writer:      &buf,
		Level:       slog.LevelWarn,
		ReplaceAttr: dropTime,
	})

	logger.Info("hidden")
	logger.Warn("visible")

	got := buf.String()
	if strings.Contains(got, "hidden") {
		t.Fatalf("expected info log to be filtered, got %q", got)
	}
	if !strings.Contains(got, `level=WARN msg=visible`) {
		t.Fatalf("expected warn log, got %q", got)
	}
}

func TestNewColorAlwaysColorsTextLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{
		Writer:      &buf,
		Color:       ColorAlways,
		ReplaceAttr: dropTime,
	})

	logger.Info("visible")

	got := buf.String()
	if !strings.Contains(got, "\x1b[32mlevel=INFO\x1b[0m") {
		t.Fatalf("expected green info level, got %q", got)
	}
}

func TestNewColorOnlyMatchesLevelField(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{
		Writer:      &buf,
		Color:       ColorAlways,
		ReplaceAttr: dropTime,
	})

	logger.Info("visible", slog.String("note", "level=INFO"))

	got := buf.String()
	if strings.Count(got, "\x1b[32mlevel=INFO\x1b[0m") != 1 {
		t.Fatalf("expected only the level field to be colored, got %q", got)
	}
	if !strings.Contains(got, `note="level=INFO"`) {
		t.Fatalf("expected note value to remain intact, got %q", got)
	}
}

func TestNewColorDoesNotAffectJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{
		Writer:      &buf,
		Format:      FormatJSON,
		Color:       ColorAlways,
		ReplaceAttr: dropTime,
	})

	logger.Info("visible")

	got := buf.String()
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("expected JSON log without color, got %q", got)
	}
}

func TestNewColorAutoSkipsNonTerminalWriters(t *testing.T) {
	var buf bytes.Buffer
	logger := New(Options{
		Writer:      &buf,
		Color:       ColorAuto,
		ReplaceAttr: dropTime,
	})

	logger.Info("visible")

	got := buf.String()
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("expected non-terminal writer without color, got %q", got)
	}
}

func TestFieldHelpers(t *testing.T) {
	err := errors.New("boom")

	tests := []struct {
		name string
		attr slog.Attr
		key  string
		want any
	}{
		{name: "request id", attr: RequestID("req-123"), key: "request_id", want: "req-123"},
		{name: "component", attr: Component("api"), key: "component", want: "api"},
		{name: "operation", attr: Operation("create_user"), key: "operation", want: "create_user"},
		{name: "error", attr: Err(err), key: "error", want: err},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.attr.Key != tt.key {
				t.Fatalf("expected key %q, got %q", tt.key, tt.attr.Key)
			}
			if got := tt.attr.Value.Any(); got != tt.want {
				t.Fatalf("expected value %v, got %v", tt.want, got)
			}
		})
	}
}

func dropTime(_ []string, attr slog.Attr) slog.Attr {
	if attr.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return attr
}
