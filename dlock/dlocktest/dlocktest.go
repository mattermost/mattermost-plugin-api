// Package dlocktest is a testing helper for you to unit test your packages that using dlock.
// simply fake dlock's store(network layer) by creating an InMemoryStore with NewStore().
package dlocktest

import (
	"reflect"
	"sync"
	"time"

	pluginapi "github.com/lieut-data/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
)

// InMemoryStore is an in-memory KV store.
type InMemoryStore struct {
	m    sync.Mutex
	data map[string]Value
}

type Value struct {
	data      interface{}
	ttl       time.Duration
	createdAt time.Time
}

// NewStore creates a new in-memory KV Store.
// TODO(ilgooz): improve InMemoryStore to simulate error cases.
func NewStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]Value),
	}
}

// Set implements a fake in memory Store. Store is defined at the dlock pkg.
func (s *InMemoryStore) Set(key string, value interface{}, options ...pluginapi.KVSetOption) (bool, error) {
	s.m.Lock()
	defer s.m.Unlock()

	opts := model.PluginKVSetOptions{}
	for _, o := range options {
		o(&opts)
	}

	if value == nil {
		delete(s.data, key)
		return true, nil
	}

	v, ok := s.data[key]
	if ok && time.Since(v.createdAt) > v.ttl {
		v.data = nil
	}

	if opts.Atomic && !reflect.DeepEqual(v.data, opts.OldValue) {
		return false, nil
	}

	s.data[key] = Value{
		data:      value,
		ttl:       time.Duration(opts.ExpireInSeconds) * time.Second,
		createdAt: time.Now(),
	}

	return true, nil
}
