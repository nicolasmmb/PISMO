package logger

import (
	"log/slog"

	"github.com/nicolasmmb/pismo-challenge/internal/port"
)

type SlogLogger struct {
	log *slog.Logger
}

func New() SlogLogger {
	return SlogLogger{log: slog.Default()}
}

func (l SlogLogger) Info(msg string, fields map[string]any) {
	l.log.Info(msg, flatten(fields)...)
}

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
