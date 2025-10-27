package pokecache

import (
	"sync"
	"time"
)

type Cache struct {
	mu      *sync.Mutex
	entries map[string]cacheEntry
}

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

func NewCache(interval time.Duration) Cache {
	cache := Cache{
		mu:      &sync.Mutex{},
		entries: map[string]cacheEntry{},
	}
	go cache.reapLoop(interval)
	return cache
}
func (cache Cache) Add(key string, val []byte) {
	(*cache.mu).Lock()
	defer (*cache.mu).Unlock()

	cache.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}
func (cache Cache) Get(key string) ([]byte, bool) {
	entry, exists := cache.entries[key]
	if exists {
		return entry.val, true
	}
	return []byte{}, false
}
func (cache Cache) Remove(key string) {
	(*cache.mu).Lock()
	defer (*cache.mu).Unlock()

	delete(cache.entries, key)
}
func (cache Cache) Contains(key string) bool {
	_, exists := cache.entries[key]
	return exists
}

func (cache Cache) reapLoop(interval time.Duration) {
	for {
		time.Sleep(interval)

		(*cache.mu).Lock()
		keysToRemove := []string{}
		for key, entry := range cache.entries {
			if time.Since(entry.createdAt) > interval {
				keysToRemove = append(keysToRemove, key)
			}
		}
		for _, key := range keysToRemove {
			delete(cache.entries, key)
		}
		(*cache.mu).Unlock()
	}
}
