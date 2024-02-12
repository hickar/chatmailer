package storage

import (
	"sync"

	"github.com/hickar/tg-remailer/internal/app/config"
)

type inmemoryClientStorage struct {
	data map[string]config.ClientConfig
	mu   sync.Mutex
}

// NewInMemoryStorage creates new inmem storage instance.
// Storage will be deleted and re-created on new deployments or on fail.
// That means we will have duplicate messages forwarded in Telegram.
// Need to consider some simple persistence for this reason.
func NewInMemoryStorage() *inmemoryClientStorage {
	return &inmemoryClientStorage{data: make(map[string]config.ClientConfig)}
}

// Retrieves the client configuration for the specified user from the storage.
// Returns the configuration and a boolean indicating whether the user exists in the storage.
func (s *inmemoryClientStorage) Get(user string) (config.ClientConfig, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.data[user]
	return item, ok
}

// Stores the client configuration for the specified user in the storage.
func (s *inmemoryClientStorage) Set(user string, data config.ClientConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[user] = data
}

// Removes the client configuration for the specified user from the storage.
// Returns a boolean indicating whether the user existed in the storage before removal.
func (s *inmemoryClientStorage) Remove(user string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data[user]
	delete(s.data, user)
	return ok
}
