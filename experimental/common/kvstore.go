package common

import (
	"errors"
	"time"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
)

var ErrNotFound = errors.New("not found")

type KVStore interface {
	// Set stores a key-value pair, unique per plugin.
	// Keys prefixed with `mmi_` are reserved for use by this package and will fail to be set.
	//
	// Returns (false, err) if DB error occurred
	// Returns (false, nil) if the value was not set
	// Returns (true, nil) if the value was set
	//
	// Minimum server version: 5.18
	Set(key string, value interface{}, options ...pluginapi.KVSetOption) (bool, error)

	// SetWithExpiry sets a key-value pair with the given expiration duration relative to now.
	//
	// Deprecated: SetWithExpiry exists to streamline adoption of this package for existing plugins.
	// Use Set with the appropriate options instead.
	//
	// Minimum server version: 5.18
	SetWithExpiry(key string, value interface{}, ttl time.Duration) error

	// CompareAndSet writes a key-value pair if the current value matches the given old value.
	//
	// Returns (false, err) if DB error occurred
	// Returns (false, nil) if the value was not set
	// Returns (true, nil) if the value was set
	//
	// Deprecated: CompareAndSet exists to streamline adoption of this package for existing plugins.
	// Use Set with the appropriate options instead.
	//
	// Minimum server version: 5.18
	CompareAndSet(key string, oldValue, value interface{}) (bool, error)

	// CompareAndDelete deletes a key-value pair if the current value matches the given old value.
	//
	// Returns (false, err) if DB error occurred
	// Returns (false, nil) if current value != oldValue or key does not exist when deleting
	// Returns (true, nil) if current value == oldValue and the key was deleted
	//
	// Deprecated: CompareAndDelete exists to streamline adoption of this package for existing plugins.
	// Use Set with the appropriate options instead.
	//
	// Minimum server version: 5.18
	CompareAndDelete(key string, oldValue interface{}) (bool, error)

	// Get gets the value for the given key into the given interface.
	//
	// An error is returned only if the value cannot be fetched. A non-existent key will return no
	// error, with nothing written to the given interface.
	//
	// Minimum server version: 5.2
	Get(key string, o interface{}) error

	// Delete deletes the given key-value pair.
	//
	// An error is returned only if the value failed to be deleted. A non-existent key will return
	// no error.
	//
	// Minimum server version: 5.18
	Delete(key string) error

	// DeleteAll removes all key-value pairs.
	//
	// Minimum server version: 5.6
	DeleteAll() error

	// ListKeys lists all keys for the plugin.
	//
	// Minimum server version: 5.6
	ListKeys(page, count int, options ...pluginapi.ListKeysOption) ([]string, error)
}
