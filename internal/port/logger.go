package port

// Logger is a minimal structured logger abstraction.
type Logger interface {
	Info(msg string, fields map[string]any)
	Error(msg string, fields map[string]any)
}
