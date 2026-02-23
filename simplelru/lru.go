package simplelru

import (
	"errors"

	"github.com/srivastavcodes/go-lru/dll"
)

// EvictCallback is used to get a callback when a cache entry is evicted.
type EvictCallback[K comparable, V any] func(key K, value V)

// LruCache is a simple LRU items implementation of fixed size and not thread-safe.
type LruCache[K comparable, V any] struct {
	size      int
	evictList *dll.LruList[K, V]
	items     map[K]*dll.Entry[K, V]
	onEvict   EvictCallback[K, V]
}

// NewLru initializes a LruCache of the given size.
func NewLru[K comparable, V any](size int, onEvict EvictCallback[K, V]) (*LruCache[K, V], error) {
	if size <= 0 {
		return nil, errors.New("size must be greater than 0")
	}
	return &LruCache[K, V]{
		size: size, evictList: dll.NewList[K, V](),
		items:   make(map[K]*dll.Entry[K, V]),
		onEvict: onEvict,
	}, nil
}

// Purge will completely clear the cache.
func (lru *LruCache[K, V]) Purge() {
	lru.evictList.Init()
	for k, v := range lru.items {
		if lru.onEvict != nil {
			lru.onEvict(k, v.Val)
		}
		delete(lru.items, k)
	}
}

// Add adds a new key-value pair to the cache. Returns true if an eviction occurred.
func (lru *LruCache[K, V]) Add(key K, val V) (evict bool) {
	if entry, ok := lru.items[key]; ok {
		lru.evictList.MoveToFront(entry)
		entry.Val = val
		return
	}
	entry := lru.evictList.PushFront(key, val)
	lru.items[key] = entry
	// check size is exceeded or not?
	evict = lru.evictList.Length() > lru.size
	if evict {
		lru.removeOldest()
	}
	return evict
}

// Get returns the value for the given key and true if exists. Also refreshes
// the recentness of the key.
func (lru *LruCache[K, V]) Get(key K) (val V, ok bool) {
	entry, ok := lru.items[key]
	if !ok {
		return
	}
	lru.evictList.MoveToFront(entry)
	return entry.Val, true
}

// GetOldest returns the oldest entry in the cache and true if exists.
func (lru *LruCache[K, V]) GetOldest() (key K, val V, ok bool) {
	if entry := lru.evictList.Back(); entry != nil {
		return entry.Key, entry.Val, true
	}
	return
}

// Contains returns true if the key exists in the cache.
func (lru *LruCache[K, V]) Contains(key K) (ok bool) {
	_, ok = lru.items[key]
	return ok
}

// Peek returns the value for the given key and true if exists.
func (lru *LruCache[K, V]) Peek(key K) (val V, ok bool) {
	entry, ok := lru.items[key]
	if !ok {
		return
	}
	return entry.Val, true
}

// Len returns the number of items in the cache.
func (lru *LruCache[K, V]) Len() int {
	return lru.evictList.Length()
}

// Cap returns the size of the cache.
func (lru *LruCache[K, V]) Cap() int {
	return lru.size
}

// Keys returns a slice of keys containing all the keys from the cache from
// oldest to newest.
func (lru *LruCache[K, V]) Keys() []K {
	keys := make([]K, lru.evictList.Length())
	i := 0
	entry := lru.evictList.Back()
	for entry != nil {
		keys[i] = entry.Key
		i++
		entry = entry.PrevEntry()
	}
	return keys
}

// Vals returns a slice of keys containing all the keys from the cache from
// oldest to newest.
func (lru *LruCache[K, V]) Vals() []V {
	vals := make([]V, lru.evictList.Length())
	i := 0
	entry := lru.evictList.Back()
	for entry != nil {
		vals[i] = entry.Val
		i++
		entry = entry.PrevEntry()
	}
	return vals
}

// Resize resizes the cache to the provided size; and returns the number of
// items evicted.
func (lru *LruCache[K, V]) Resize(size int) (evictCount int) {
	diff := lru.Len() - size
	if size < 0 {
		size = 0
	}
	for i := 0; i < diff; i++ {
		lru.removeOldest()
	}
	lru.size = size
	return lru.size
}

// Remove removes the key from the cache.
func (lru *LruCache[K, V]) Remove(key K) (present bool) {
	if entry, present := lru.items[key]; present {
		lru.removeElement(entry)
		return present
	}
	return
}

// RemoveOldest removes the oldest entry from the cache and returns key/val
// and true.
func (lru *LruCache[K, V]) RemoveOldest() (key K, val V, ok bool) {
	if entry := lru.evictList.Back(); entry != nil {
		lru.removeElement(entry)
		return entry.Key, entry.Val, true
	}
	return
}

// removeOldest removes the oldest entry from the cache.
func (lru *LruCache[K, V]) removeOldest() {
	if entry := lru.evictList.Back(); entry != nil {
		lru.removeElement(entry)
	}
}

// removeElement removes the given entry from the cache and executes the onEvict
// function if provided.
func (lru *LruCache[K, V]) removeElement(entry *dll.Entry[K, V]) {
	lru.evictList.Remove(entry)
	delete(lru.items, entry.Key)
	if lru.onEvict != nil {
		lru.onEvict(entry.Key, entry.Val)
	}
}
