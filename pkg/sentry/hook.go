package sentry

import (
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

var (
	logrusLevelsToSentryLevels = map[logrus.Level]sentry.Level{
		logrus.PanicLevel: sentry.LevelFatal,
		logrus.FatalLevel: sentry.LevelFatal,
		logrus.ErrorLevel: sentry.LevelError,
		logrus.WarnLevel:  sentry.LevelWarning,
		logrus.InfoLevel:  sentry.LevelInfo,
		logrus.DebugLevel: sentry.LevelDebug,
		logrus.TraceLevel: sentry.LevelDebug,
	}
)

func NewHook(levels []logrus.Level) Hook {
	return Hook{levels: levels}
}

type Hook struct {
	levels []logrus.Level
}

func (h Hook) Levels() []logrus.Level {
	return h.levels
}

func (h Hook) Fire(entry *logrus.Entry) error {
	var exception error

	if err, ok := entry.Data[logrus.ErrorKey].(error); ok && err != nil {
		exception = err
	} else {
		// Make a new error with the log message if there is no error provided
		// because stacktraces are neat
		exception = fmt.Errorf(entry.Message)
	}

	tags, hasTags := entry.Data["tags"].(map[string]string)

	sentry.WithScope(func(scope *sentry.Scope) {
		scope.AddEventProcessor(func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			event.Message = entry.Message
			return event
		})

		scope.SetLevel(logrusLevelsToSentryLevels[entry.Level])

		if hasTags {
			scope.SetTags(tags)
			delete(entry.Data, "tags")  // Remove ugly map rendering
			scope.SetExtras(entry.Data) // Set the extras in Sentry without the redundant tag data
			for k, v := range tags {    // Add the tags in a sane way back to Logrus
				entry.Data[k] = v
			}
		} else {
			scope.SetExtras(entry.Data)
		}

		sentry.CaptureException(exception)
	})

	return nil
}
