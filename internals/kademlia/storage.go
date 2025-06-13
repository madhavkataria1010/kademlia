package kademlia

import "github.com/Aradhya2708/kademlia/pkg/models"

// NewKeyValueStore creates a new thread-safe KeyValueStore.
func NewKeyValueStore() *models.KeyValueStore {
	return models.NewKeyValueStore()
}

// StoreKeyValue stores a key-value pair in the KeyValueStore.
func StoreKeyValue(kvs *models.KeyValueStore, key, value string) {
	kvs.Set(key, value)
}

// FindValue retrieves the value for a given key from the KeyValueStore.
func FindValue(kvs *models.KeyValueStore, key string) (string, bool) {
	return kvs.Get(key)
}
