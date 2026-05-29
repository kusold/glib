package log

import "log/slog"

// RequestID returns the conventional request correlation field.
func RequestID(id string) slog.Attr {
	return slog.String("request_id", id)
}

// Component returns the conventional application component field.
func Component(name string) slog.Attr {
	return slog.String("component", name)
}

// Operation returns the conventional operation field.
func Operation(name string) slog.Attr {
	return slog.String("operation", name)
}

// Err returns the conventional error field.
func Err(err error) slog.Attr {
	return slog.Any("error", err)
}
