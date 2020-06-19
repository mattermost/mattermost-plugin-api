package telemetry

import rudder "github.com/rudderlabs/analytics-go"

// rudderDataPlaneURL is common for all Mattermost Projects
const rudderDataPlaneURL = "https://pdat.matterlytics.com"

// rudderWriteKey is set during build time adding the following line to build/custom.mk
// LDFLAGS += -X "github.com/mattermost/mattermost-plugin-api/experimental/telemetry.rudderWriteKey=$(MM_RUDDER_WRITE_KEY)"
// MM_RUDDER_WRITE_KEY environment variable must be set also during CI to "1dP7Oi78p0PK1brYLsfslgnbD1I"
// In order to use telemetry in development environment, use the following lines in build/custom.mk
// ifndef MM_RUDDER_WRITE_KEY
// MM_RUDDER_WRITE_KEY = 1d5bMvdrfWClLxgK1FvV3s4U1tg
// endif
var rudderWriteKey string

// NewRudderClient creates a new telemetry client with Rudder using the default configuration
func NewRudderClient() (Client, error) {
	return NewRudderClientWithCredentials(rudderWriteKey, rudderDataPlaneURL)
}

// NewRudderClientWithCredentials lets you create a Rudder client with your own credentials
func NewRudderClientWithCredentials(writeKey, dataPlaneURL string) (Client, error) {
	client, err := rudder.NewWithConfig(writeKey, dataPlaneURL, rudder.Config{})
	if err != nil {
		return nil, err
	}

	return &rudderWrapper{client: client}, nil
}

type rudderWrapper struct {
	client rudder.Client
}

func (r *rudderWrapper) Enqueue(t Track) error {
	err := r.client.Enqueue(rudder.Track{
		UserId:     t.UserID,
		Event:      t.Event,
		Properties: t.Properties,
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *rudderWrapper) Close() error {
	err := r.client.Close()
	if err != nil {
		return err
	}
	return nil
}
