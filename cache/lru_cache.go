package cache

import (
	"container/list"
	"os"
	"sync"
	"time"
)

// LRUCache structure
type LRUCache struct {
	mu       sync.Mutex
	capacity int
	cache    map[string]*list.Element
	eviction *list.List
}

// CacheEntry stores key-value and expiration time
type CacheEntry struct {
	key        string
	value      []byte
	expiration time.Time
}

// NewLRUCache initializes an LRU cache
func NewLRUCache(capa int) *LRUCache {
	return &LRUCache{
		capacity: capa,
		cache:    make(map[string]*list.Element),
		eviction: list.New(),
	}
}

// Get retrieves from cache
func (l *LRUCache) Get(key string) ([]byte, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if element, found := l.cache[key]; found {
		entry := element.Value.(*CacheEntry)

		// Check if expired
		if time.Now().After(entry.expiration) {
			l.removeElement(element)
			return nil, false
		}

		// Move accessed item to front (recently used)
		l.eviction.MoveToFront(element)
		return entry.value, true
	}
	return nil, false
}

// Put inserts an item into the cache
func (l *LRUCache) Put(key string, value []byte, ttl time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// If already exists, update & move to front
	if element, found := l.cache[key]; found {
		l.eviction.MoveToFront(element)
		entry := element.Value.(*CacheEntry)
		entry.value = value
		entry.expiration = time.Now().Add(ttl)
		return
	}

	// If cache is full, evict the oldest entry
	if l.eviction.Len() >= l.capacity {
		l.removeOldest()
	}

	// Insert new entry at front
	entry := &CacheEntry{key, value, time.Now().Add(ttl)}
	element := l.eviction.PushFront(entry)
	l.cache[key] = element
}

// removeOldest removes least recently used item
func (l *LRUCache) removeOldest() {
	element := l.eviction.Back()
	if element != nil {
		l.removeElement(element)
	}
}

// removeElement deletes an element from cache
func (l *LRUCache) removeElement(element *list.Element) {
	entry := element.Value.(*CacheEntry)
	delete(l.cache, entry.key)
	l.eviction.Remove(element)
}

// Cache structure using LRUCache
type Cache struct {
	lru *LRUCache
	ttl time.Duration
}

func NewCache(capacity int, ttl time.Duration) *Cache {
	return &Cache{
		lru: NewLRUCache(capacity),
		ttl: ttl,
	}
}

// Retrieve a file from the cache or fetch it from the disk
func (c *Cache) GetFile(path string) ([]byte, error) {
	// Try to get from LRU cache
	if data, found := c.lru.Get(path); found {
		return data, nil
	}

	// Load from disk
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.lru.Put(path, data, c.ttl)
	return data, nil
}
