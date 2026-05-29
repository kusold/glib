package log

import (
	"bytes"
	"io"
	"log/slog"
	"os"
)

// Format selects the handler encoding used by NewHandler.
type Format string

const (
	// FormatText writes slog's text format.
	FormatText Format = "text"

	// FormatJSON writes slog's JSON format.
	FormatJSON Format = "json"
)

// ColorMode controls ANSI colors for text log levels.
type ColorMode int

const (
	// ColorNever disables color output.
	ColorNever ColorMode = iota

	// ColorAlways enables ANSI colors for text log levels.
	ColorAlways

	// ColorAuto enables ANSI colors when the output appears to be a terminal.
	ColorAuto
)

// Options customizes logger construction.
type Options struct {
	// Writer receives log output. It defaults to os.Stderr.
	Writer io.Writer

	// Format selects text or JSON output. It defaults to FormatText.
	Format Format

	// Level filters records below this level. It defaults to slog.LevelInfo.
	Level slog.Leveler

	// AddSource includes source file and line information.
	AddSource bool

	// ReplaceAttr rewrites attributes before they are logged.
	ReplaceAttr func(groups []string, attr slog.Attr) slog.Attr

	// Color controls ANSI colorized level tags for text output. It defaults to
	// ColorNever and has no effect for JSON logs.
	Color ColorMode
}

// New returns a standard slog.Logger configured with production-friendly
// defaults.
func New(opts Options) *slog.Logger {
	return slog.New(NewHandler(opts))
}

// NewHandler returns a standard slog.Handler configured from opts.
func NewHandler(opts Options) slog.Handler {
	w := opts.Writer
	if w == nil {
		w = os.Stderr
	}

	handlerOpts := &slog.HandlerOptions{
		AddSource:   opts.AddSource,
		Level:       opts.Level,
		ReplaceAttr: opts.ReplaceAttr,
	}

	switch opts.Format {
	case FormatJSON:
		return slog.NewJSONHandler(w, handlerOpts)
	default:
		if shouldColor(opts.Color, w) {
			w = colorLevelWriter{writer: w}
		}
		return slog.NewTextHandler(w, handlerOpts)
	}
}

func shouldColor(mode ColorMode, w io.Writer) bool {
	switch mode {
	case ColorAlways:
		return true
	case ColorAuto:
		file, ok := w.(*os.File)
		if !ok {
			return false
		}
		info, err := file.Stat()
		if err != nil {
			return false
		}

		// ModeCharDevice marks character devices such as terminals and TTYs.
		// Auto color only enables ANSI output for those devices, avoiding escape
		// codes when logs are written to regular files, buffers, or pipes.
		return info.Mode()&os.ModeCharDevice != 0
	default:
		return false
	}
}

type colorLevelWriter struct {
	writer io.Writer
}

func (w colorLevelWriter) Write(p []byte) (int, error) {
	colored := colorLevels(p)
	_, err := w.writer.Write(colored)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func colorLevels(p []byte) []byte {
	out := p
	for _, rule := range levelColorRules {
		out = colorLevel(out, []byte("level="+rule.level), []byte(rule.color+"level="+rule.level+colorReset))
	}
	return out
}

func colorLevel(p, target, colored []byte) []byte {
	var out []byte
	last := 0
	start := 0

	for {
		idx := bytes.Index(p[start:], target)
		if idx < 0 {
			if out == nil {
				return p
			}
			out = append(out, p[last:]...)
			return out
		}
		idx += start

		start = idx + len(target)
		if !isLevelToken(p, idx, len(target)) {
			continue
		}

		if out == nil {
			out = make([]byte, 0, len(p)+len(colored)-len(target))
		}
		out = append(out, p[last:idx]...)
		out = append(out, colored...)
		last = start
	}
}

func isLevelToken(p []byte, idx, length int) bool {
	if idx > 0 && p[idx-1] != ' ' {
		return false
	}
	end := idx + length
	return end == len(p) || p[end] == ' ' || p[end] == '\n'
}

const colorReset = "\x1b[0m"

var levelColorRules = []struct {
	level string
	color string
}{
	{level: slog.LevelError.String(), color: "\x1b[31m"},
	{level: slog.LevelWarn.String(), color: "\x1b[33m"},
	{level: slog.LevelInfo.String(), color: "\x1b[32m"},
	{level: slog.LevelDebug.String(), color: "\x1b[34m"},
}
