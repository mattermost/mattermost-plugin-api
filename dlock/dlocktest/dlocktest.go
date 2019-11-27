// Package dlocktest is a testing helper for you to unit test your packages that using dlock.
// simply fake dlock's store(network layer) by using InMemoryStore.
package dlocktest

import (
	"reflect"
	"sync"
	"time"

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
func NewStore() *InMemoryStore {
	return &InMemoryStore{
		data: make(map[string]Value),
	}
}

func (s *InMemoryStore) KVSetWithOptions(key string, newValue interface{}, options model.PluginKVSetOptions) (bool, *model.AppError) {
	s.m.Lock()
	defer s.m.Unlock()
	v, ok := s.data[key]
	if ok && time.Since(v.createdAt) > v.ttl {
		v.data = nil
	}
	if options.Atomic && !reflect.DeepEqual(v.data, options.OldValue) {
		return false, nil
	}
	s.data[key] = Value{newValue, time.Duration(options.ExpireInSeconds) * time.Second, time.Now()}
	return true, nil
}

func (s *InMemoryStore) KVDelete(key string) *model.AppError {
	s.m.Lock()
	defer s.m.Unlock()
	delete(s.data, key)
	return nil
}
