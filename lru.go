package golru

import (
	"sync"

	"github.com/srivastavcodes/gocache/simplelru"
)

const (
	// DefaultEvictedBufferSize defines the default buffer size to store
	// evicted key/val pairs.
	DefaultEvictedBufferSize = 16
)

// Cache is a thread-safe fixed size lru-cache.
type Cache[K comparable, V any] struct {
	rwm sync.RWMutex
	lru *simplelru.LruCache[K, V]

	// buffers to store evicted key/val pairs.
	// used when onEvictCb is provided.
	keysEvicted []K
	valsEvicted []V

	onEvictCb func(key K, val V)
}

// NewLru Creates an Lru of the given size.
func NewLru[K comparable, V any](size int) (*Cache[K, V], error) {
	return NewLruWithEvict[K, V](size, nil)
}

// NewLruWithEvict initializes a fixed size cache with the given eviction callback.
func NewLruWithEvict[K comparable, V any](size int, onEvict func(key K, val V)) (*Cache[K, V], error) {
	var err error
	lru := &Cache[K, V]{
		onEvictCb: onEvict,
	}
	if onEvict != nil {
		lru.initEvictBuffer()
		onEvict = lru.onEvict
	}
	lru.lru, err = simplelru.NewLru(size, onEvict)
	return lru, err
}

// initEvictBuffer initializes the key/valsEvicted fields of the cache.
func (c *Cache[K, V]) initEvictBuffer() {
	c.keysEvicted = make([]K, 0, DefaultEvictedBufferSize)
	c.valsEvicted = make([]V, 0, DefaultEvictedBufferSize)
}

// onEvict is the default callback used for filling the evicted key and val buffers.
func (c *Cache[K, V]) onEvict(key K, val V) {
	c.keysEvicted = append(c.keysEvicted, key)
	c.valsEvicted = append(c.valsEvicted, val)
}

// Purge is used to completely clear the cache.
func (c *Cache[K, V]) Purge() {
	var keys []K
	var vals []V
	c.rwm.Lock()
	c.lru.Purge()
	if c.onEvictCb != nil && len(c.keysEvicted) > 0 {
		keys = c.keysEvicted
		vals = c.valsEvicted
		c.initEvictBuffer() // reset the buffer
	}
	c.rwm.Unlock()
	if c.onEvictCb != nil {
		for i := 0; i < len(keys); i++ {
			c.onEvictCb(keys[i], vals[i])
		}
	}
}

// Add adds a new key-value pair to the cache and calls the onEvict callback
// if an element was evicted.
func (c *Cache[K, V]) Add(key K, val V) bool {
	var k K
	var v V
	c.rwm.Lock()
	evicted := c.lru.Add(key, val)
	if c.onEvictCb != nil && evicted {
		k = c.keysEvicted[0]
		v = c.valsEvicted[0]
		c.keysEvicted = c.keysEvicted[:0]
		c.valsEvicted = c.valsEvicted[:0]
	}
	c.rwm.Unlock()
	if c.onEvictCb != nil && evicted {
		c.onEvictCb(k, v)
	}
	return evicted
}

// ContainsOrAdd first checks if the key exists within the cache without updating
// the recentness of the key; and if not, adds the key/val pair.
// Returns whether the key was found and if an eviction occurred.
func (c *Cache[K, V]) ContainsOrAdd(key K, val V) (ok, evicted bool) {
	var k K
	var v V
	c.rwm.Lock()
	if ok = c.lru.Contains(key); ok {
		c.rwm.Unlock()
		return true, false
	}
	evicted = c.lru.Add(key, val)
	if c.onEvictCb != nil && evicted {
		k = c.keysEvicted[0]
		v = c.valsEvicted[0]
		c.keysEvicted = c.keysEvicted[:0]
		c.valsEvicted = c.valsEvicted[:0]
	}
	c.rwm.Unlock()
	if c.onEvictCb != nil && evicted {
		c.onEvictCb(k, v)
	}
	return false, evicted
}

// PeekOrAdd checks if a key exists in the cache and returns the value if it does
// without updating the recentness of it; and if not exist, then adds the key/val
// pair to the cache.
// Returns previous value, whether a key was found, and if an eviction occurred.
func (c *Cache[K, V]) PeekOrAdd(key K, val V) (prev V, ok, evicted bool) {
	var k K
	var v V
	c.rwm.Lock()
	prev, ok = c.lru.Peek(key)
	if ok {
		c.rwm.Unlock()
		return prev, true, false
	}
	evicted = c.lru.Add(key, val)
	if c.onEvictCb != nil && evicted {
		k = c.keysEvicted[0]
		v = c.valsEvicted[0]
		c.keysEvicted = c.keysEvicted[:0]
		c.valsEvicted = c.valsEvicted[:0]
	}
	c.rwm.Unlock()
	if c.onEvictCb != nil && evicted {
		c.onEvictCb(k, v)
	}
	return prev, false, evicted
}

// Remove removes the key from the cache if it exists. Returns whether the cache
// existed or not. Calls the callback function if an eviction occurred.
func (c *Cache[K, V]) Remove(key K) bool {
	var k K
	var v V
	c.rwm.Lock()
	present := c.lru.Remove(key)
	if c.onEvictCb != nil && present {
		k = c.keysEvicted[0]
		v = c.valsEvicted[0]
		c.keysEvicted = c.keysEvicted[:0]
		c.valsEvicted = c.valsEvicted[:0]
	}
	c.rwm.Unlock()
	if c.onEvictCb != nil && present {
		c.onEvictCb(k, v)
	}
	return present
}

// Resize resizes the cache to the given size and returns the number of keys
// evicted if any.
func (c *Cache[K, V]) Resize(size int) int {
	var keys []K
	var vals []V
	c.rwm.Lock()
	evicted := c.lru.Resize(size)
	if c.onEvictCb != nil && evicted > 0 {
		keys = c.keysEvicted
		vals = c.valsEvicted
		c.initEvictBuffer()
	}
	c.rwm.Unlock()
	if c.onEvictCb != nil && evicted > 0 {
		for i := 0; i < len(keys); i++ {
			c.onEvictCb(keys[i], vals[i])
		}
	}
	return evicted
}

func (c *Cache[K, V]) RemoveOldest() (K, V, bool) {
	var k K
	var v V
	c.rwm.Lock()
	key, val, ok := c.lru.RemoveOldest()
	if c.onEvictCb != nil && ok {
		k = c.keysEvicted[0]
		v = c.valsEvicted[0]
		c.keysEvicted = c.keysEvicted[:0]
		c.valsEvicted = c.valsEvicted[:0]
	}
	c.rwm.Unlock()
	if c.onEvictCb != nil && ok {
		c.onEvictCb(k, v)
	}
	return key, val, ok
}

// Get returns the value and true for the given key if it exists in the cache,
// or nil and false otherwise.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.rwm.Lock()
	defer c.rwm.Unlock()
	return c.lru.Get(key)
}

// Contains returns true if the key exists in the cache.
func (c *Cache[K, V]) Contains(key K) bool {
	c.rwm.RLock()
	defer c.rwm.RUnlock()
	return c.lru.Contains(key)
}

// Peek returns the value for the given key without updating the recentness
// of the key.
func (c *Cache[K, V]) Peek(key K) (V, bool) {
	c.rwm.RLock()
	defer c.rwm.RUnlock()
	return c.lru.Peek(key)
}

// GetOldest returns the oldest entry.
func (c *Cache[K, V]) GetOldest() (K, V, bool) {
	c.rwm.RLock()
	c.rwm.RUnlock()
	return c.lru.GetOldest()

}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *Cache[K, V]) Keys() []K {
	c.rwm.RLock()
	c.rwm.RUnlock()
	return c.lru.Keys()
}

// Vals returns a slice of the values in the cache, from oldest to newest.
func (c *Cache[K, V]) Vals() []V {
	c.rwm.RLock()
	c.rwm.RUnlock()
	return c.lru.Vals()
}

// Len returns the number of items in the cache.
func (c *Cache[K, V]) Len() int {
	c.rwm.RLock()
	c.rwm.RUnlock()
	return c.lru.Len()
}

// Cap returns the capacity of the cache
func (c *Cache[K, V]) Cap() int {
	return c.lru.Cap()
}
