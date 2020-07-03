package logger

import (
	"github.com/mattermost/mattermost-plugin-api/experimental/bot/poster"
	"github.com/mattermost/mattermost-plugin-api/experimental/telemetry"
)

// Option defines each option that can be passed in the creation of a Logger.
// Options functions available are MessageToAdmins and LogToTelemetry.
type Option func(*defaultLogger)

// MessageToAdmins makes the Logger DM all the userIDs provided through the dmer all messages
// over the logLevel defined. If logVerbose is set, the log context will be also sent.
// Available logLevels are "debug", "info", "warn", or "error".
func MessageToAdmins(logLevel string, logVerbose bool, dmer poster.DMer, userIDs ...string) Option {
	return func(l *defaultLogger) {
		l.Config.AdminLogLevel = logLevel
		l.Config.AdminLogVerbose = logVerbose
		l.admin = NewAdmin(dmer, userIDs...)
	}
}

// LogToTelemetry makes the Logger send through the tracker all message over the log level defined.
// Available logLevels are "debug", "info", "warn", or "error".
func LogToTelemetry(logLevel string, tracker telemetry.Tracker) Option {
	return func(l *defaultLogger) {
		l.Config.TelemetryLogLevel = logLevel
		l.tracker = tracker
	}
}
