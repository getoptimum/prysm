package utils_test

import (
	"github.com/OffchainLabs/prysm/v6/utils"
	"github.com/dgraph-io/ristretto"
	"github.com/stretchr/testify/require"
	"strconv"
	"sync"
	"testing"
	"time"
)

func BenchmarkTTLMap(b *testing.B) {
	maxTTL := 1 * time.Second
	cleanupInterval := 1 * time.Second
	m := utils.NewTTLMap[string, int](maxTTL, cleanupInterval)

	for i := 0; i < b.N; i++ {
		key := "key" + strconv.Itoa(i)
		m.Put(key, i)
	}
}

func BenchmarkRistretto(b *testing.B) {
	b.Skip("it gives not so big difference in performance, but it additional dependency")
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	require.NoError(b, err)
	defer cache.Close()
	a := 10 * time.Second

	for i := 0; i < b.N; i++ {
		key := "key" + strconv.Itoa(i)
		cache.SetWithTTL(key, i, 1, a)
	}
}

func TestTTLMap(t *testing.T) {
	t.Run("basic put and get operations", func(t *testing.T) {
		maxTTL := 1 * time.Second
		cleanupInterval := 1 * time.Second
		m := utils.NewTTLMap[string, int](maxTTL, cleanupInterval)

		// Put a value
		m.Put("key", 42)
		require.Equal(t, 1, m.Len(), "expected TTLMap length to be 1")

		// Get the value
		val, exists := m.Get("key")
		require.True(t, exists, "expected value to exist for key 'key', but it doesn't")
		require.Equal(t, 42, val, "expected value 42, got %d", val)

		// Wait for the cleanup routine to run
		// Check if the value is still present after cleanup
		require.Eventually(t, func() bool {
			_, exists = m.Get("key")
			return !exists
		}, 5*cleanupInterval, cleanupInterval, "expected value to be deleted after TTL expiration, but it still exists")
	})

	t.Run("concurrent put and get operations", func(t *testing.T) {
		maxTTL := 2 * time.Second
		cleanupInterval := 1 * time.Second
		m := utils.NewTTLMap[string, int](maxTTL, cleanupInterval)

		// Use a wait group to wait for goroutines to finish
		var wg sync.WaitGroup

		// Concurrent put operations
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				key := "key" + strconv.Itoa(index)
				m.Put(key, index)
			}(i)
		}

		// Concurrent get operations
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				key := "key" + strconv.Itoa(index)
				val, exists := m.Get(key)
				if exists {
					require.Equal(t, index, val, "expected value %d, got %d", index, val)
				}
			}(i)
		}

		wg.Wait()
	})
}
