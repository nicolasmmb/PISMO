package logger

import (
	"log/slog"

	"github.com/nicolasmmb/pismo-challenge/internal/port"
)

// SlogLogger adapts log/slog to the Logger port.
type SlogLogger struct {
	log *slog.Logger
}

// New creates a new SlogLogger with default handler.
func New() SlogLogger {
	return SlogLogger{log: slog.Default()}
}

// Info logs an informational message.
func (l SlogLogger) Info(msg string, fields map[string]any) {
	l.log.Info(msg, flatten(fields)...)
}

// Error logs an error message.
func (l SlogLogger) Error(msg string, fields map[string]any) {
	l.log.Error(msg, flatten(fields)...)
}

func flatten(fields map[string]any) []any {
	if len(fields) == 0 {
		return nil
	}
	out := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		out = append(out, k, v)
	}
	return out
}

var _ port.Logger = SlogLogger{}
