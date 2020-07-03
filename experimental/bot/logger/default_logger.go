// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package logger

import (
	"fmt"
	"time"

	"github.com/mattermost/mattermost-plugin-api/experimental/common"
	"github.com/mattermost/mattermost-plugin-api/experimental/telemetry"
)

type defaultLogger struct {
	Config
	admin      Admin
	tracker    telemetry.Tracker
	logContext LogContext
	logAPI     common.LogAPI
}

/*
New creates a new logger.

- api: LogAPI implementation

- options: Optional parameters for the Logger. Available options are MessageToAdmins and LogToTelemetry.
*/
func New(api common.LogAPI, options ...Option) Logger {
	l := &defaultLogger{
		logAPI: api,
	}
	for _, option := range options {
		option(l)
	}

	return l
}

func (l *defaultLogger) With(logContext LogContext) Logger {
	newLogger := *l
	if len(newLogger.logContext) == 0 {
		newLogger.logContext = map[string]interface{}{}
	}
	for k, v := range logContext {
		newLogger.logContext[k] = v
	}
	return &newLogger
}

func (l *defaultLogger) Timed() Logger {
	return l.With(LogContext{
		timed: time.Now(),
	})
}

func (l *defaultLogger) Debugf(format string, args ...interface{}) {
	measure(l.logContext)
	message := fmt.Sprintf(format, args...)
	l.logAPI.Debug(message, toKeyValuePairs(l.logContext)...)
	if level(l.AdminLogLevel) >= 4 {
		l.logToAdmins("DEBUG", message)
	}
	if level(l.TelemetryLogLevel) >= 4 {
		l.logToTelemetry("DEBUG", message)
	}
}

func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	measure(l.logContext)
	message := fmt.Sprintf(format, args...)
	l.logAPI.Error(message, toKeyValuePairs(l.logContext)...)
	if level(l.AdminLogLevel) >= 1 {
		l.logToAdmins("ERROR", message)
	}
	if level(l.TelemetryLogLevel) >= 1 {
		l.logToTelemetry("ERROR", message)
	}
}

func (l *defaultLogger) Infof(format string, args ...interface{}) {
	measure(l.logContext)
	message := fmt.Sprintf(format, args...)
	l.logAPI.Info(message, toKeyValuePairs(l.logContext)...)
	if level(l.AdminLogLevel) >= 3 {
		l.logToAdmins("INFO", message)
	}
	if level(l.TelemetryLogLevel) >= 3 {
		l.logToTelemetry("INFO", message)
	}
}

func (l *defaultLogger) Warnf(format string, args ...interface{}) {
	measure(l.logContext)
	message := fmt.Sprintf(format, args...)
	l.logAPI.Warn(message, toKeyValuePairs(l.logContext)...)
	if level(l.AdminLogLevel) >= 2 {
		l.logToAdmins("WARN", message)
	}
	if level(l.TelemetryLogLevel) >= 2 {
		l.logToTelemetry("WARN", message)
	}
}

func (l *defaultLogger) logToAdmins(level, message string) {
	if l.admin == nil {
		return
	}

	if l.AdminLogVerbose && len(l.logContext) > 0 {
		message += "\n" + common.JSONBlock(l.logContext)
	}
	_ = l.admin.DMAdmins("(log " + level + ") " + message)
}

func (l *defaultLogger) logToTelemetry(level, message string) {
	if l.tracker == nil {
		return
	}

	properties := map[string]interface{}{}
	properties["message"] = message
	for k, v := range l.logContext {
		properties["context_"+k] = fmt.Sprintf("%v", v)
	}

	l.tracker.TrackEvent("logger_"+level, properties)
}
