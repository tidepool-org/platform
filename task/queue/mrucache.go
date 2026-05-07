package queue

import (
	"slices"
	"sync"
	"time"
)

const (
	// DefaultLRUCacheMaxSize is based on current-day production having 7,456 tasks in
	// either the pending or running states.
	DefaultLRUCacheMaxSize = 75000
)

// mruCache is a cache of most recently used key-value pairs.
type mruCache struct {
	// cache allows fast lookup by key.
	cache map[string]time.Time
	// sorted is an ordered slice of keys, such that the most recently used is last.
	sorted  []string
	maxSize int
	mu      sync.Mutex
}

func NewMRUCache(maxSize int) *mruCache {
	if maxSize < 1 {
		maxSize = DefaultLRUCacheMaxSize
	}
	return &mruCache{
		maxSize: maxSize,
		cache:   make(map[string]time.Time, max(100, maxSize/10)),
		sorted:  make([]string, 0, max(100, maxSize/10)),
	}
}

func (l *mruCache) Store(key string, value time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.cache[key] = value
	for len(l.cache) > l.maxSize {
		delete(l.cache, l.sorted[0])
		l.sorted = l.sorted[1:]
	}
	l.refresh(key)
}

// refresh the key's position in the sorted list of keys.
func (l *mruCache) refresh(key string) {
	l.sorted = slices.DeleteFunc(l.sorted, func(i string) bool {
		return i == key
	})
	l.sorted = append(l.sorted, key)
}

func (l *mruCache) Get(key string) time.Time {
	l.mu.Lock()
	defer l.mu.Unlock()
	value, found := l.cache[key]
	if !found {
		return time.Time{}
	}
	l.refresh(key)
	return value
}
