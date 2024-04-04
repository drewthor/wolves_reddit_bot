package slogmiddleware

import (
	"context"
	"log/slog"
	"slices"
	"time"

	"go.opentelemetry.io/otel/trace"
)

func NewOtelSlogHandler(logKey string, levels []slog.Level, handler slog.Handler) OtelSlogHandler {
	return OtelSlogHandler{logKey: logKey, levels: levels, handler: handler}
}

type OtelSlogHandler struct {
	logKey  string
	levels  []slog.Level
	handler slog.Handler
}

func (h OtelSlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return slices.Contains(h.levels, level) || h.handler.Enabled(ctx, level)
}

func (h OtelSlogHandler) Handle(ctx context.Context, record slog.Record) error {
	if !slices.Contains(h.levels, record.Level) {
		h.handler.Handle(ctx, record)
		return nil
	}

	span := trace.SpanFromContext(ctx)
	if span != nil {
		span.AddEvent(record.Message, trace.WithTimestamp(time.Now()))
	}
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		if spanContext.HasTraceID() {
			record.AddAttrs(slog.String("trace_id", spanContext.TraceID().String()))
		}
		if spanContext.HasSpanID() {
			record.AddAttrs(slog.String("span_id", spanContext.SpanID().String()))
		}
	}

	h.handler.Handle(ctx, record)

	return nil
}

func (h OtelSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewOtelSlogHandler(h.logKey, h.levels, h.handler.WithAttrs(attrs))
}

func (h OtelSlogHandler) WithGroup(name string) slog.Handler {
	return NewOtelSlogHandler(h.logKey, h.levels, h.handler.WithGroup(name))
}
