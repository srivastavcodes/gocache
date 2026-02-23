package simplelru

// LruCacher is the interface for a simple Lru implementation.
type LruCacher[K comparable, V any] interface {
	// Add adds a value to the cache and returns true if an eviction occurred
	// and updates the recently usedness of the key.
	Add(key K, val V) bool

	// Get Returns key's value from the cache and updates the recently usedness
	// of the key. Returns value, isFound.
	Get(key K) (value V, ok bool)

	// Contains Checks if a key exists in the cache without updating the
	// recent-ness.
	Contains(key K) (ok bool)

	// Peek returns the key's value, (if)exits without updating the recently
	// usedness of the key.
	Peek(key K) (value V, ok bool)

	// Remove removes a key from the cache.
	Remove(key K) bool

	// RemoveOldest removes the oldest entry from the cache.
	RemoveOldest() (K, V, bool)

	// GetOldest returns the oldest entry from the cache. Returns key, value,
	// isFound
	GetOldest() (K, V, bool)

	// Keys returns a slice of the keys in the cache, from oldest to newest.
	Keys() []K

	// Values returns a slice of the values in the cache, from oldest to newest.
	Values() []V

	// Len returns the number of items in the cache.
	Len() int

	// Cap returns the capacity of the cache.
	Cap() int

	// Purge clears all cache entries.
	Purge()

	// Resize resizes cache, returning number evicted.
	Resize(int) int
}
