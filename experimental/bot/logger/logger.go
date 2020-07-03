package logger

import "time"

const timed = "__since"
const elapsed = "Elapsed"

// LogContext defines the context for the logs.
type LogContext map[string]interface{}

// Logger defines an object able to log messages.
type Logger interface {
	// With adds a logContext to the logger.
	With(LogContext) Logger
	// Timed add a timed log context.
	Timed() Logger
	// Debugf logs a formatted string as a debug message.
	Debugf(format string, args ...interface{})
	// Errorf logs a formatted string as an error message.
	Errorf(format string, args ...interface{})
	// Infof logs a formatted string as an info message.
	Infof(format string, args ...interface{})
	// Warnf logs a formatted string as an warning message.
	Warnf(format string, args ...interface{})
}

func measure(lc LogContext) {
	if lc[timed] == nil {
		return
	}
	started := lc[timed].(time.Time)
	lc[elapsed] = time.Since(started).String()
	delete(lc, timed)
}

func level(l string) int {
	switch l {
	case "debug":
		return 4
	case "info":
		return 3
	case "warn":
		return 2
	case "error":
		return 1
	}
	return 0
}

func toKeyValuePairs(in map[string]interface{}) (out []interface{}) {
	for k, v := range in {
		out = append(out, k, v)
	}
	return out
}
