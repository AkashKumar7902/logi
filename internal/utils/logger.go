package utils

import (
	"context"
	"io"
	"log/slog"
	"os"
	"sync/atomic"
)

const RequestIDHeader = "X-Request-ID"

type contextKey string

const requestIDContextKey contextKey = "request_id"

var defaultLogger atomic.Pointer[slog.Logger]

func init() {
	ConfigureLogger("")
}

func ConfigureLogger(environment string) {
	setLogger(newLogger(os.Stdout, environment != "production"))
}

func newLogger(output io.Writer, addSource bool) *slog.Logger {
	return slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{
		AddSource: addSource,
		Level:     slog.LevelInfo,
	}))
}

func setLogger(logger *slog.Logger) {
	defaultLogger.Store(logger)
	slog.SetDefault(logger)
}

func baseLogger() *slog.Logger {
	if logger := defaultLogger.Load(); logger != nil {
		return logger
	}

	ConfigureLogger("")
	return defaultLogger.Load()
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if requestID == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDContextKey, requestID)
}

func RequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	requestID, _ := ctx.Value(requestIDContextKey).(string)
	return requestID
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if requestID := RequestIDFromContext(ctx); requestID != "" {
		return baseLogger().With("request_id", requestID)
	}
	return baseLogger()
}

func Info(ctx context.Context, msg string, args ...any) {
	LoggerFromContext(ctx).Info(msg, args...)
}

func Warn(ctx context.Context, msg string, args ...any) {
	LoggerFromContext(ctx).Warn(msg, args...)
}

func Error(ctx context.Context, msg string, args ...any) {
	LoggerFromContext(ctx).Error(msg, args...)
}

func InfoBackground(msg string, args ...any) {
	baseLogger().Info(msg, args...)
}

func WarnBackground(msg string, args ...any) {
	baseLogger().Warn(msg, args...)
}

func ErrorBackground(msg string, args ...any) {
	baseLogger().Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	baseLogger().Error(msg, args...)
	os.Exit(1)
}
