package storage

import (
	"sync"

	"github.com/hickar/tg-remailer/internal/app/config"
)

type inmemoryClientStorage struct {
	data map[string]config.ClientConfig
	mu   sync.Mutex
}

func NewInMemoryStorage() *inmemoryClientStorage {
	return &inmemoryClientStorage{data: make(map[string]config.ClientConfig)}
}

func (s *inmemoryClientStorage) Get(user string) (config.ClientConfig, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.data[user]
	return item, ok
}

func (s *inmemoryClientStorage) Set(user string, data config.ClientConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[user] = data
}

func (s *inmemoryClientStorage) Remove(user string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data[user]
	delete(s.data, user)
	return ok
}
