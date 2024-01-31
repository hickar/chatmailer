package main

import "sync"

type inmemoryClientStorage struct {
	data map[string]ClientConfig
	mu   sync.Mutex
}

func newInMemoryStorage() *inmemoryClientStorage {
	return &inmemoryClientStorage{data: make(map[string]ClientConfig)}
}

func (s *inmemoryClientStorage) Get(user string) (ClientConfig, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.data[user]
	return item, ok
}

func (s *inmemoryClientStorage) Set(user string, data ClientConfig) {
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
