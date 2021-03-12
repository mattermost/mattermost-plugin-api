package pluginapi

import (
	"github.com/mattermost/mattermost-server/v5/plugin"
)

// AppsCacheService exposes methods to manipulate apps cached entities.
type AppsCacheService struct {
	api plugin.API
}

// Set stores a set of apps entities for a given app ID
// Returns err if an error occurred
// Returns nil if the value was set
//
// Minimum server version: 5.4
func (acs *AppsCacheService) Set(appID string, inputs map[string][][]byte) error {
	if appErr := acs.api.AppsCacheSet(appID, inputs); appErr != nil {
		return normalizeAppErr(appErr)
	}
	return nil
}

// Get gets the value for the given key for a given app ID
//
// Minimum server version: 5.4
func (acs *AppsCacheService) Get(appID string, key string) ([][]byte, error) {
	o := [][]byte{}

	o, appErr := acs.api.AppsCacheGet(appID, key)
	if appErr != nil {
		return nil, normalizeAppErr(appErr)
	}

	if len(o) == 0 {
		return nil, nil
	}

	return o, nil
}

// Delete deletes the entities associated to the key for an app ID.
//
// An error is returned only if the value failed to be deleted. A non-existent key will return
// no error.
//
// Minimum server version: 5.4
func (acs *AppsCacheService) Delete(appID string, key string) error {
	return normalizeAppErr(acs.api.AppsCacheDelete(appID, key))
}

// DeleteAll removes all entries for an app ID
//
// Minimum server version: 5.4
func (acs *AppsCacheService) DeleteAll(appID string) error {
	return normalizeAppErr(acs.api.AppsCacheDeleteAll(appID))
}