package store

import (
	"strconv"
	"sync"
	"time"
)

type KVStore struct {
	store      map[string]string
	mutex      sync.RWMutex
	expiries   map[string]time.Time
	gcInterval time.Duration
}

func runGCRoutine(store *KVStore) {
	for {

		store.mutex.RLock()
		expiredKeys := make([]string, 0, len(store.expiries))
		for key, expiry := range store.expiries {
			if store.isExpired(expiry) {
				expiredKeys = append(expiredKeys, key)
			}
		}

		store.mutex.RUnlock()

		if len(expiredKeys) > 0 {
			store.mutex.Lock()

			for _, key := range expiredKeys {
				if expiry, exists := store.expiries[key]; exists && store.isExpired(expiry) {
					delete(store.store, key)
					delete(store.expiries, key)
				}
			}

			store.mutex.Unlock()
		}

		time.Sleep(store.gcInterval)
	}
}

func NewKVStore() *KVStore {
	store := &KVStore{
		store:      make(map[string]string),
		expiries:   make(map[string]time.Time),
		gcInterval: 1 * time.Second,
	}

	go runGCRoutine(store)

	return store
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

	if s.checkAndRemoveIfExpired(key) {
		return "", false
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()
	value, exists := s.store[key]
	return value, exists
}

func (s *KVStore) Delete(key string) bool {
	if !s.Has(key) {
		return false
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.store, key)
	delete(s.expiries, key)
	return true
}

func (s *KVStore) Add(key string, x int) (int, error) {

	s.checkAndRemoveIfExpired(key)

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

func (s *KVStore) Keys() []string {
	s.mutex.RLock()

	keys := make([]string, 0, len(s.store))
	for key := range s.store {
		keys = append(keys, key)
	}

	s.mutex.RUnlock()

	validKeys := make([]string, 0, len(keys))
	now := time.Now()
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	for _, key := range keys {
		if expiry, exists := s.expiries[key]; !exists || !expiry.Before(now) {
			validKeys = append(validKeys, key)
		}
	}

	return validKeys
}

func (s *KVStore) checkAndRemoveIfExpired(key string) bool {
	s.mutex.RLock()
	expiry, exists := s.expiries[key]
	if !exists {
		s.mutex.RUnlock()
		return false
	}
	isExpired := expiry.Before(time.Now())
	s.mutex.RUnlock()

	if !isExpired {
		return false
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	expiry, exists = s.expiries[key]
	if !exists || !expiry.Before(time.Now()) {
		return false
	}

	delete(s.store, key)
	delete(s.expiries, key)
	return true
}

func (s *KVStore) Expire(key string, ttl int) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.store[key]; !exists {
		return false
	}

	if expiry, hasExpiry := s.expiries[key]; hasExpiry && expiry.Before(time.Now()) {
		delete(s.store, key)
		delete(s.expiries, key)
		return false
	}

	s.expiries[key] = time.Now().Add(time.Duration(ttl) * time.Second)
	return true
}

func (s *KVStore) TTL(key string) int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if _, exists := s.store[key]; !exists {
		return -2
	}

	expiry, hasExpiry := s.expiries[key]

	if !hasExpiry {
		return -1
	}

	remaining := time.Until(expiry).Seconds()
	ttl := int(remaining)
	if remaining <= 0 {
		return -2
	}

	return ttl
}

func (s *KVStore) isExpired(expiry time.Time) bool {
	return expiry.Before(time.Now())
}

func (s *KVStore) Persist(key string) bool {
	s.mutex.RLock()
	defer s.mutex.Unlock()

	if _, exists := s.store[key]; !exists {
		return false
	}

	if _, hasExpiry := s.expiries[key]; !hasExpiry {
		return false
	}

	if s.isExpired(s.expiries[key]) {
		delete(s.store, key)
		delete(s.expiries, key)
		return false
	}

	delete(s.expiries, key)

	return true
}
