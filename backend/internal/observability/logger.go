// Package observability provides structured logging, metrics, and tracing hooks.
package observability

import (
	"context"
	"log/slog"
	"os"
)

type ctxKey string

const (
	requestIDKey ctxKey = "request_id"
	loggerKey    ctxKey = "logger"
)

// NewLogger returns a JSON slog logger at the given level (debug|info|warn|error).
func NewLogger(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	return slog.New(h)
}

// WithRequestID stores a request id in the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestID extracts the request id from the context, or "" if absent.
func RequestID(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// WithLogger attaches a logger (already enriched with correlation fields) to the context.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// LoggerFrom returns the context logger, falling back to the default logger.
func LoggerFrom(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
