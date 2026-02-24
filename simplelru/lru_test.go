package simplelru

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLru(t *testing.T) {
	var evicts int
	onEvict := func(key int, val int) {
		require.Equal(t, key, val)
		evicts++
	}
	lru, err := NewLru(128, onEvict)
	require.NoError(t, err)
	for i := 0; i < 256; i++ {
		lru.Add(i, i)
	}
	require.Equal(t, 128, lru.Cap())
	require.Equal(t, 128, evicts)
	require.Equal(t, 128, lru.Len())
	for i, key := range lru.Keys() {
		val, ok := lru.Get(key)
		require.True(t, ok)
		require.Equal(t, key, val)
		require.Equalf(t, val, 128+i, "val=%d, index=%d", val, 128+i)
	}
	for i, val := range lru.Vals() {
		key, ok := lru.Get(val)
		require.True(t, ok)
		require.Equal(t, key, val)
		require.Equalf(t, key, 128+i, "val=%d, index=%d", key, 128+i)
	}
	for i := 0; i < 128; i++ {
		_, ok := lru.Get(i)
		require.False(t, ok, "should be evicted")
	}
	for i := 128; i < 256; i++ {
		_, ok := lru.Get(i)
		require.Truef(t, ok, "should not be evicted. i=%d", i)
	}
	for i := 128; i < 200; i++ {
		require.True(t, lru.Remove(i), "should be removed")
		require.False(t, lru.Remove(i), "should have been removed")
		_, ok := lru.Get(i)
		require.False(t, ok, "should not be in cache")
	}
	// refreshing the cache for the key
	lru.Get(200)
	for i, key := range lru.Keys() {
		require.Condition(t, func() bool {
			return (i < 55 && key != 200+i) || (i == 55 && key == 200)
		}, "key out of order")
	}
	lru.Purge()
	require.Zero(t, lru.Len())
	_, ok := lru.Get(0)
	require.False(t, ok, "should not be in cache")
}
