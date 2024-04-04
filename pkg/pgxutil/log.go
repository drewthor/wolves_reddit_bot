package pgxutil

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/tracelog"
)

type Logger struct {
	l        *slog.Logger
	errorKey string
}

type Option func(*Logger)

func WithErrorKey(errorKey string) Option {
	return func(logger *Logger) {
		logger.errorKey = errorKey
	}
}

func NewLogger(l *slog.Logger, options ...Option) *Logger {
	logger := &Logger{l: l, errorKey: "INVALID_PGX_LOG_LEVEL"}

	for _, option := range options {
		option(logger)
	}

	return logger
}

func (l *Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	logger := l.l
	for k, v := range data {
		logger = logger.With(k, v)
	}

	switch level {
	case tracelog.LogLevelTrace:
		logger.Log(ctx, slog.LevelDebug-1, msg, "PGX_LOG_LEVEL", level)
	case tracelog.LogLevelDebug:
		logger.DebugContext(ctx, msg)
	case tracelog.LogLevelInfo:
		logger.InfoContext(ctx, msg)
	case tracelog.LogLevelWarn:
		logger.WarnContext(ctx, msg)
	case tracelog.LogLevelError:
		logger.ErrorContext(ctx, msg)
	default:
		logger.ErrorContext(ctx, msg, slog.Any(l.errorKey, fmt.Errorf("invalid pgx log level: %d", level)))
	}
}
