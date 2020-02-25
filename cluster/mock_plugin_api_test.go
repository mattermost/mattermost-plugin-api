package cluster

import (
	"bytes"
	"sync"
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
)

type mockPluginAPI struct {
	t *testing.T

	lock      sync.Mutex
	keyValues map[string][]byte
	failing   bool
}

func newMockPluginAPI(t *testing.T) *mockPluginAPI {
	return &mockPluginAPI{
		t:         t,
		keyValues: make(map[string][]byte),
	}
}

func (pluginAPI *mockPluginAPI) setFailing(failing bool) {
	pluginAPI.lock.Lock()
	defer pluginAPI.lock.Unlock()

	pluginAPI.failing = failing
}

func (pluginAPI *mockPluginAPI) KVGet(key string) ([]byte, *model.AppError) {
	pluginAPI.lock.Lock()
	defer pluginAPI.lock.Unlock()

	if pluginAPI.failing {
		return nil, &model.AppError{Message: "fake error"}
	}

	return pluginAPI.keyValues[key], nil
}

func (pluginAPI *mockPluginAPI) KVSet(key string, value []byte) *model.AppError {
	pluginAPI.lock.Lock()
	defer pluginAPI.lock.Unlock()

	if pluginAPI.failing {
		return &model.AppError{Message: "fake error"}
	}

	pluginAPI.keyValues[key] = value

	return nil
}

func (pluginAPI *mockPluginAPI) KVDelete(key string) *model.AppError {
	pluginAPI.lock.Lock()
	defer pluginAPI.lock.Unlock()

	if pluginAPI.failing {
		return &model.AppError{Message: "fake error"}
	}

	delete(pluginAPI.keyValues, key)

	return nil
}

func (pluginAPI *mockPluginAPI) KVCompareAndSet(key string, oldValue []byte, value []byte) (bool, *model.AppError) {
	pluginAPI.lock.Lock()
	defer pluginAPI.lock.Unlock()

	if pluginAPI.failing {
		return false, &model.AppError{Message: "fake error"}
	}

	if actualValue := pluginAPI.keyValues[key]; !bytes.Equal(actualValue, oldValue) {
		return false, nil
	}

	pluginAPI.keyValues[key] = value

	return true, nil
}

func (pluginAPI *mockPluginAPI) KVCompareAndDelete(key string, oldValue []byte) (bool, *model.AppError) {
	pluginAPI.lock.Lock()
	defer pluginAPI.lock.Unlock()

	if pluginAPI.failing {
		return false, &model.AppError{Message: "fake error"}
	}

	if actualValue := pluginAPI.keyValues[key]; !bytes.Equal(actualValue, oldValue) {
		return false, nil
	}

	delete(pluginAPI.keyValues, key)

	return true, nil
}

func (pluginAPI *mockPluginAPI) LogError(msg string, keyValuePairs ...interface{}) {
	if pluginAPI.t == nil {
		return
	}

	pluginAPI.t.Helper()

	params := []interface{}{msg}
	params = append(params, keyValuePairs...)

	pluginAPI.t.Log(params...)
}
