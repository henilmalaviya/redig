// Package store provides an in-memory key-value store with expiration and background cleanup.

package store

import (
	"strconv"
	"sync"
	"time"
)

// KVStore is a thread-safe key-value store with expiration and GC.
type KVStore struct {
	store    map[string]string
	mutex    sync.RWMutex
	expiries map[string]time.Time

	// this defines the frequency of GC routine
	gcInterval time.Duration
}

// runGCRoutine cleans up expired keys in the background every gcInterval
func runGCRoutine(store *KVStore) {
	for {
		// acquire read lock to collect expired keys
		// instead of acquiring full lock and checking every iteration
		// this specific operation is what we call RFCL (Read First, Check Later)
		// the operation is meant to simplify the dead-lock situations and reduce the full-lock duration

		store.mutex.RLock()
		expiredKeys := make([]string, 0, len(store.expiries))
		for key, expiry := range store.expiries {
			if store.isExpired(expiry) {
				expiredKeys = append(expiredKeys, key)
			}
		}

		store.mutex.RUnlock()

		// if any expired keys were found, acquire full lock and delete them
		if len(expiredKeys) > 0 {
			store.mutex.Lock()

			for _, key := range expiredKeys {
				// Recheck avoids race where key’s expiry changes mid-flight.
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

// NewKVStore spins up a store and starts GC with a 1-second interval.
func NewKVStore() *KVStore {
	store := &KVStore{
		store:      make(map[string]string),
		expiries:   make(map[string]time.Time),
		gcInterval: 1 * time.Second,
	}

	go runGCRoutine(store)

	return store
}

// Set sets a key-value pair into the store.
func (s *KVStore) Set(key string, value string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.store[key] = value
}

// Has checks if a key’s alive and not expired.
func (s *KVStore) Has(key string) bool {
	// the reason we don't lock here is because we use Get call which internally handles the lock
	// and because Get already tells us if the key is alive or not
	// we just fetch the exists boolean returned by Get
	_, exists := s.Get(key)
	return exists
}

// Get grabs a value if the key’s there and not expired.
func (s *KVStore) Get(key string) (string, bool) {

	// lazy expiration check
	// every-time Get is called, we first check if the key is expired
	// if the key is expired, treat the key as non-existent
	if s.checkAndRemoveIfExpired(key) {
		return "", false
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, exists := s.store[key]
	return value, exists
}

// Delete wipes a key if it exists and not expired
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

// Add tweaks a numeric value by x, starts at 0 if key’s new.
func (s *KVStore) Add(key string, x int) (int, error) {

	s.checkAndRemoveIfExpired(key)

	// acquire full lock for atomic operation
	// if we acquired read lock until the increment operation,
	// there is a potential race condition
	s.mutex.Lock()
	defer s.mutex.Unlock()

	value, exists := s.store[key]

	if !exists {
		value = "0"
	}

	i, err := strconv.Atoi(value)

	// string to int conversion can fail, if the value is not an integer
	if err != nil {
		return 0, err
	}

	i += x

	s.store[key] = strconv.Itoa(i)

	return i, nil
}

// Incr bumps a value by 1.
func (s *KVStore) Incr(key string) (int, error) {
	return s.Add(key, 1)
}

// Decr drops a value by 1.
func (s *KVStore) Decr(key string) (int, error) {
	return s.Add(key, -1)
}

// Keys lists all non-expired keys.
func (s *KVStore) Keys() []string {
	// we are performing RFCL here, read above in runGCRoutine
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

// checkAndRemoveIfExpired deletes a key if it’s expired, instant fallback for GC.
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
	// double check the expiry
	if !exists || !expiry.Before(time.Now()) {
		return false
	}

	delete(s.store, key)
	delete(s.expiries, key)
	return true
}

// Expire sets a TTL on a key, bails if key’s gone or expired.
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

// TTL shows seconds left for a key: -2 if non-existent/expired, -1 if exists but no expiry.
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

// isExpired checks if time is before now
func (s *KVStore) isExpired(expiry time.Time) bool {
	return expiry.Before(time.Now())
}

// Persist yanks a key’s expiration if it’s still good.
func (s *KVStore) Persist(key string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

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
