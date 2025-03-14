package store

import (
	"strconv"
	"sync"
)

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
	_, exists := s.Get(key)
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

func (s *KVStore) Add(key string, x int) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	value, exists := s.store[key]

	if !exists {
		value = "0"
	}

	i, err := strconv.Atoi(value)

	if err != nil {
		return 0, err
	}

	i += x

	s.store[key] = strconv.Itoa(i)

	return i, nil
}

func (s *KVStore) Incr(key string) (int, error) {
	return s.Add(key, 1)
}

func (s *KVStore) Decr(key string) (int, error) {
	return s.Add(key, -1)
}
