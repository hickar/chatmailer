package kvstore

import (
	"sync"
)

type KVStore[K comparable, V any] struct {
	data map[K]V
	mu   sync.RWMutex
}

// New creates new KVStore instance.
func New[K comparable, V any]() *KVStore[K, V] {
	return &KVStore[K, V]{data: make(map[K]V)}
}

// Get returns value by key.
func (s *KVStore[K, V]) Get(key K) (V, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.data[key]
	return item, ok
}

// Set stores value in storage making it accessible by key.
func (s *KVStore[K, V]) Set(key K, data V) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = data
}

// Remove entry by key.
func (s *KVStore[K, V]) Remove(key K) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data[key]
	delete(s.data, key)
	return ok
}
