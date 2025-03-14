package store

import "sync"

type KVStore struct {
	store map[string]string
	mutex sync.RWMutex
}

func NewKVStore() *KVStore {
	return &KVStore{
		store: make(map[string]string),
	}
}

func (s *KVStore) Set(key string, value string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.store[key] = value
}

func (s *KVStore) Has(key string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	_, exists := s.store[key]
	return exists
}

func (s *KVStore) Get(key string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	value, exists := s.store[key]
	return value, exists
}

func (s *KVStore) Delete(key string) bool {
	if s.Has(key) {
		s.mutex.Lock()
		defer s.mutex.Unlock()

		delete(s.store, key)
		return true
	}

	return false
}
