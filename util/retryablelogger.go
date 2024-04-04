package util

import "log/slog"

type RetryableLogger struct {
	l slog.Logger
}

func (r RetryableLogger) Error(msg string, keysAndValues ...interface{}) {
	r.l.Error(msg, keysAndValues)
}

func (r RetryableLogger) Info(msg string, keysAndValues ...interface{}) {
	r.l.Info(msg, keysAndValues)
}

func (r RetryableLogger) Debug(msg string, keysAndValues ...interface{}) {
	r.l.Debug(msg, keysAndValues)
}

func (r RetryableLogger) Warn(msg string, keysAndValues ...interface{}) {
	r.l.Warn(msg, keysAndValues)
}
