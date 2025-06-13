package models

import "sync"

// KeyValueStore represents a thread-safe key-value store
type KeyValueStore struct {
	mu    sync.RWMutex
	Store map[string]string
}

// NewKeyValueStore initializes a new KeyValueStore
func NewKeyValueStore() *KeyValueStore {
	return &KeyValueStore{
		Store: make(map[string]string),
	}
}

// Set stores a key-value pair
func (kv *KeyValueStore) Set(key, value string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.Store[key] = value
}

// Get retrieves the value for a given key
func (kv *KeyValueStore) Get(key string) (string, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()
	value, exists := kv.Store[key]
	return value, exists
}

func (kv *KeyValueStore) GetAll() map[string]string {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	// Create a copy to ensure thread safety
	copy := make(map[string]string, len(kv.Store))
	for key, value := range kv.Store {
		copy[key] = value
	}
	return copy
}
