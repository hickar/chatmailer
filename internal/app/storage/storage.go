package storage

import (
	"sync"
)

type inmemoryClientStorage[K comparable, V any] struct {
	data map[K]V
	mu   sync.RWMutex
}

// NewInMemoryStorage creates new inmem storage instance.
// Storage will be deleted and re-created on new deployments or on fail.
// That means we will have duplicate messages forwarded in Telegram.
// Need to consider some simple persistence for this reason.
func NewInMemoryStorage[K comparable, V any]() *inmemoryClientStorage[K, V] {
	return &inmemoryClientStorage[K, V]{data: make(map[K]V)}
}

// Retrieves the client configuration for the specified user from the storage.
// Returns the configuration and a boolean indicating whether the user exists in the storage.
func (s *inmemoryClientStorage[K, V]) Get(key K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.data[key]
	return item, ok
}

// Stores the client configuration for the specified user in the storage.
func (s *inmemoryClientStorage[K, V]) Set(key K, data V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = data
}

// Removes the client configuration for the specified user from the storage.
// Returns a boolean indicating whether the user existed in the storage before removal.
func (s *inmemoryClientStorage[K, V]) Remove(key K) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data[key]
	delete(s.data, key)
	return ok
}
