package constants

import "sync"

var (
	// Default values for Kademlia
	kValue = 1 // Bucket size, can be updated dynamically

	// Mutex for thread-safe access
	mu sync.RWMutex
)

// GetK returns the current value of k
func GetK() int {
	mu.RLock()
	defer mu.RUnlock()
	return kValue
}

// SetK allows updating the value of k dynamically
func SetK(value int) {
	mu.Lock()
	defer mu.Unlock()
	kValue = value
}
