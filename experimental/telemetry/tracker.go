package telemetry

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/plugin"

	"github.com/mattermost/mattermost-plugin-api/experimental/bot/logger"
)

type TrackerConfig struct {
	EnabledTracking bool
	EnabledLogging  bool
}

// NewTrackerConfig returns a new trackerConfig from the current values from the plugin API.
func NewTrackerConfig(api plugin.API) TrackerConfig {
	var enabledTracking, enabledLogging bool
	if config := api.GetConfig(); config != nil {
		if enableDiagnostics := config.LogSettings.EnableDiagnostics; enableDiagnostics != nil {
			enabledTracking = *enableDiagnostics
		}
        
		if enableDeveloper := config.ServiceSettings.EnableDeveloper; enableDeveloper != nil {
			enabledLogging = *enableDeveloper
		}
    }

    return TrackerConfig{
		EnabledTracking: enabledTracking,
		EnabledLogging:  enabledLogging,
	}
}

// Tracker defines a telemetry tracker
type Tracker interface {
	// TrackEvent registers an event through the configured telemetry client
	TrackEvent(event string, properties map[string]interface{}) error
	// TrackUserEvent registers an event through the configured telemetry client associated to a user
	TrackUserEvent(event string, userID string, properties map[string]interface{}) error
	// Reload Config re-evaluates tracker config to determine if tracking behavior should change
	ReloadConfig(config TrackerConfig)
}

// Client defines a telemetry client
type Client interface {
	// Enqueue adds a tracker event (Track) to be registered
	Enqueue(t Track) error
	// Close closes the client connection, flushing any event left on the queue
	Close() error
}

// Track defines an event ready for the client to process
type Track struct {
	UserID     string
	Event      string
	Properties map[string]interface{}
}

type tracker struct {
	client             Client
	diagnosticID       string
	serverVersion      string
	pluginID           string
	pluginVersion      string
	telemetryShortName string
	enabledTracking    bool
	enabledLogging     bool
	logger             logger.Logger
}

// NewTracker creates a default Tracker
// - c Client: A telemetry client. If nil, the tracker will not track any event.
// - diagnosticID: Server unique ID used for telemetry.
// - severVersion: Mattermost server version.
// - pluginID: The plugin ID.
// - pluginVersion: The plugin version.
// - telemetryShortName: Short name for the plugin to use in telemetry. Used to avoid dot separated names like `com.company.pluginName`.
// If a empty string is provided, it will use the pluginID.
// - config: Whether the system has enabled sending telemetry data. If false, the tracker will not track any event.
// - l Logger: A logger to debug event tracking and some important changes (it wont log if nil is passed as logger).
func NewTracker(
	c Client,
	diagnosticID,
	serverVersion,
	pluginID,
	pluginVersion,
	telemetryShortName string,
	config TrackerConfig,
	l logger.Logger,
) Tracker {
	if telemetryShortName == "" {
		telemetryShortName = pluginID
	}
	return &tracker{
		telemetryShortName: telemetryShortName,
		client:             c,
		diagnosticID:       diagnosticID,
		serverVersion:      serverVersion,
		pluginID:           pluginID,
		pluginVersion:      pluginVersion,
		enabledLogging:     config.EnabledLogging,
		enabledTracking:    config.EnabledTracking,
		logger:             l,
	}
}

func (t *tracker) ReloadConfig(config TrackerConfig) {
	if config.EnabledTracking != t.enabledTracking {
		if config.EnabledTracking {
			t.debugf("Enabling plugin telemetry")
		} else {
			t.debugf("Disabling plugin telemetry")
		}
		t.enabledTracking = config.EnabledTracking
	}
	t.enabledLogging = config.EnabledLogging
}

func (t *tracker) debugf(message string, args ...interface{}) {
	if t.logger == nil || !t.enabledLogging {
		return
	}
	t.logger.Debugf(message, args...)
}

func (t *tracker) TrackEvent(event string, properties map[string]interface{}) error {
	event = t.telemetryShortName + "_" + event
	if !t.enabledTracking || t.client == nil {
		t.debugf("Plugin telemetry event `%s` tracked, but not sent due to configuration", event)
		return nil
	}

	if properties == nil {
		properties = map[string]interface{}{}
	}
	properties["PluginID"] = t.pluginID
	properties["PluginVersion"] = t.pluginVersion
	properties["ServerVersion"] = t.serverVersion

	err := t.client.Enqueue(Track{
		UserID:     t.diagnosticID, // We consider the server the "user" on the telemetry system. Any reference to the actual user is passed by properties.
		Event:      event,
		Properties: properties,
	})

	if err != nil {
		return errors.Wrap(err, "cannot enqueue the track")
	}
	t.debugf("Tracked plugin telemetry event `%s`", event)

	return nil
}

func (t *tracker) TrackUserEvent(event, userID string, properties map[string]interface{}) error {
	if properties == nil {
		properties = map[string]interface{}{}
	}

	properties["UserActualID"] = userID
	return t.TrackEvent(event, properties)
}
