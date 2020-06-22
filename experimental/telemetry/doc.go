// Package telemetry allows you to add telemetry to your plugins.
// For Rudder, you can set the data plane URL and the write key on build time,
// to allow having different keys for production and development.
// If you are working on a Mattermost project, the data plane URL is already set.
// In order to add the write key for both development and production environment,
// add the following lines to build/custom.mk
//   ifndef MM_RUDDER_WRITE_KEY
//     MM_RUDDER_WRITE_KEY = 1d5bMvdrfWClLxgK1FvV3s4U1tg
//   endif
//   LDFLAGS += -X "github.com/mattermost/mattermost-plugin-api/experimental/telemetry.rudderWriteKey=$(MM_RUDDER_WRITE_KEY)"
// MM_RUDDER_WRITE_KEY environment variable must be set also during CI
// to the production write key ("1dP7Oi78p0PK1brYLsfslgnbD1I").
// If you want to use your own data plane URL, add also this line and
// make sure the MM_RUDDER_DATA_PLANE_URL variable is set.
//   LDFLAGS += -X "github.com/mattermost/mattermost-plugin-api/experimental/telemetry.rudderDataPlaneURL=$(MM_RUDDER_DATA_PLANE_URL)"
package telemetry
