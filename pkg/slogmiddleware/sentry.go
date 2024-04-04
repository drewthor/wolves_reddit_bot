package slogmiddleware

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/getsentry/sentry-go"
)

var (
	slogLevelsToSentryLevels = map[slog.Level]sentry.Level{
		slog.LevelError: sentry.LevelError,
		slog.LevelWarn:  sentry.LevelWarning,
		slog.LevelInfo:  sentry.LevelInfo,
		slog.LevelDebug: sentry.LevelDebug,
	}
)

func NewSentrySlogHandler(errorKey string, levels []slog.Level, handler slog.Handler) SentrySlogHandler {
	return SentrySlogHandler{errorKey: errorKey, levels: levels, handler: handler}
}

type SentrySlogHandler struct {
	errorKey string
	levels   []slog.Level
	handler  slog.Handler
}

func (h SentrySlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return slices.Contains(h.levels, level) || h.handler.Enabled(ctx, level)
}

func (h SentrySlogHandler) Handle(ctx context.Context, record slog.Record) error {
	defer h.handler.Handle(ctx, record)

	if !slices.Contains(h.levels, record.Level) {
		return nil
	}
	exception := fmt.Errorf(record.Message)
	attrs := make(map[string]string)
	record.Attrs(func(attr slog.Attr) bool {
		attrs[attr.Key] = fmt.Sprintf("%v", attr.Value)
		if attr.Key == h.errorKey {
			if err, ok := attr.Value.Any().(error); ok {
				exception = err
			}
		}
		return true
	})

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.AddEventProcessor(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			event.Message = record.Message
			return event
		})

		scope.SetLevel(slogLevelsToSentryLevels[record.Level])
		scope.SetTags(attrs)
		sentry.CaptureException(exception)
	})

	return nil
}

func (h SentrySlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewSentrySlogHandler(h.errorKey, h.levels, h.handler.WithAttrs(attrs))
}

func (h SentrySlogHandler) WithGroup(name string) slog.Handler {
	return NewSentrySlogHandler(h.errorKey, h.levels, h.handler.WithGroup(name))
}
