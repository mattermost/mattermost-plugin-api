// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package logger

// Config contains all configuration needed by the logger.
type Config struct {
	// AdminLogLevel is "debug", "info", "warn", or "error".
	AdminLogLevel LogLevel

	// AdminLogVerbose: set to include full context with admin log messages.
	AdminLogVerbose bool

	// TelemetryLogLevel is "debug", "info", "warn", or "error".
	TelemetryLogLevel LogLevel
}

// ToStorableConfig merge the configuration with the provided configMap.
func (c Config) ToStorableConfig(configMap map[string]interface{}) map[string]interface{} {
	if configMap == nil {
		configMap = map[string]interface{}{}
	}
	configMap["AdminLogLevel"] = c.AdminLogLevel
	configMap["AdminLogVerbose"] = c.AdminLogVerbose
	configMap["TelemetryLogLevel"] = c.TelemetryLogLevel
	return configMap
}
