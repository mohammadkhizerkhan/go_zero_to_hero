package sync

import (
	"sync"
)

// Cache is a thread-safe in-memory key-value store protected by a sync.RWMutex.
type Cache struct {
	mu    sync.RWMutex
	store map[string]string
}

// NewCache creates and initializes a new Cache.
func NewCache() *Cache {
	return &Cache{store: make(map[string]string)}
}

// Get reads the value for a key using RLock (shared read lock).
// Multiple goroutines can execute Get simultaneously without blocking each other.
func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.store[key]
	return val, ok
}

// Set writes a key-value pair using Lock (exclusive write lock).
// Blocks all readers and writers until the operation completes.
func (c *Cache) Set(key, val string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = val
}

// MutexCache is a cache protected by a standard exclusive sync.Mutex.
// Every read and write acquires an exclusive lock.
type MutexCache struct {
	mu    sync.Mutex
	store map[string]string
}

func NewMutexCache() *MutexCache {
	return &MutexCache{store: make(map[string]string)}
}

func (c *MutexCache) Get(key string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, ok := c.store[key]
	return val, ok
}

func (c *MutexCache) Set(key, val string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = val
}

// NoLockCache is an unprotected cache with no synchronization.
// Useful for baseline single-threaded benchmarks (UNSAFE for concurrent writes!).
type NoLockCache struct {
	store map[string]string
}

func NewNoLockCache() *NoLockCache {
	return &NoLockCache{store: make(map[string]string)}
}

func (c *NoLockCache) Get(key string) (string, bool) {
	val, ok := c.store[key]
	return val, ok
}

func (c *NoLockCache) Set(key, val string) {
	c.store[key] = val
}