package utils

import "sync"

// RWMap provides a thread-safe map implementation using read-write locks.
type RWMap[K comparable, V any] struct {
	mu       sync.RWMutex
	Internal map[K]V
}

// NewRWMap creates a new empty thread-safe map.
func NewRWMap[K comparable, V any]() *RWMap[K, V] {
	return &RWMap[K, V]{
		Internal: make(map[K]V),
	}
}

// NewRWMapFromStdMap creates a new thread-safe map from an existing standard map.
func NewRWMapFromStdMap[K comparable, V any](internal map[K]V) *RWMap[K, V] {
	return &RWMap[K, V]{
		Internal: internal,
	}
}

// Replace replaces the internal map with the provided one and returns the old one.
func (rm *RWMap[K, V]) Replace(internal map[K]V) map[K]V {
	rm.mu.Lock()
	current := rm.Internal
	rm.Internal = internal
	rm.mu.Unlock()
	return current
}

// Load retrieves a value from the map by its key.
func (rm *RWMap[K, V]) Load(key K) (value V, ok bool) {
	rm.mu.RLock()
	result, ok := rm.Internal[key]
	rm.mu.RUnlock()
	return result, ok
}

// LoadAll returns a copy of the entire map.
func (rm *RWMap[K, V]) LoadAll() map[K]V {
	rm.mu.RLock()
	actual := make(map[K]V, len(rm.Internal))
	for key, value := range rm.Internal {
		actual[key] = value
	}
	rm.mu.RUnlock()
	return actual
}

// Delete removes a key-value pair from the map.
func (rm *RWMap[K, V]) Delete(key K) {
	rm.mu.Lock()
	delete(rm.Internal, key)
	rm.mu.Unlock()
}

// DeleteAll removes all key-value pairs from the map.
func (rm *RWMap[K, V]) DeleteAll() {
	rm.mu.Lock()
	for key := range rm.Internal {
		delete(rm.Internal, key)
	}
	rm.mu.Unlock()
}

// Store adds or updates a key-value pair in the map.
func (rm *RWMap[K, V]) Store(key K, value V) {
	rm.mu.Lock()
	rm.Internal[key] = value
	rm.mu.Unlock()
}

// Keys returns a slice of all keys in the map.
func (rm *RWMap[K, V]) Keys() []K {
	rm.mu.Lock()
	result := make([]K, 0, len(rm.Internal))
	for key := range rm.Internal {
		result = append(result, key)
	}
	rm.mu.Unlock()
	return result
}

// Range calls the provided function for each key/value pair in the map.
// If the function returns false, iteration stops.
func (rm *RWMap[K, V]) Range(f func(key K, value V) bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	for k, v := range rm.Internal {
		if !f(k, v) {
			break
		}
	}
}

// Do executes a function on a value under read lock. Returns false if the key is not found.
func (rm *RWMap[K, V]) Do(key K, fn func(value V)) bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	value, ok := rm.Internal[key]
	if !ok {
		return false
	}
	fn(value)
	return true
}

// LoadAllAndErase returns a copy of the entire map.
func (rm *RWMap[K, V]) LoadAllAndErase() map[K]V {
	rm.mu.Lock()
	actual := make(map[K]V, len(rm.Internal))
	for key, value := range rm.Internal {
		actual[key] = value
	}
	rm.Internal = make(map[K]V) // Clear the map
	rm.mu.Unlock()
	return actual
}
