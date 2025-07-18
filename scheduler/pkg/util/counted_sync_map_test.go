package util

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCountedSyncMap_Store(t *testing.T) {
	t.Parallel()

	m := NewCountedSyncMap[string]()

	m.Store("key1", "value1")
	require.Equal(t, 1, m.Length(), "length should be 1 after storing new key")

	m.Store("key1", "updated_value1")
	require.Equal(t, 1, m.Length(), "length should remain 1 after updating existing key")

	m.Store("key2", "value2")
	require.Equal(t, 2, m.Length(), "length should be 2 after storing second key")
}

func TestCountedSyncMap_Load(t *testing.T) {
	t.Parallel()

	m := NewCountedSyncMap[string]()

	val, ok := m.Load("nonexistent")
	require.False(t, ok, "Load should return false for non-existent key")
	require.Nil(t, val, "Load should return nil for non-existent key")

	m.Store("key1", "value1")
	val, ok = m.Load("key1")
	require.True(t, ok, "Load should return true for existing key")
	require.Equal(t, "value1", *val, "Load should return correct value")

	m.Store("key1", "updated_value1")
	val, ok = m.Load("key1")
	require.True(t, ok, "Load should return true for updated key")
	require.Equal(t, "updated_value1", *val, "Load should return updated value")
}

func TestCountedSyncMap_Delete(t *testing.T) {
	t.Parallel()

	m := NewCountedSyncMap[string]()

	m.Delete("nonexistent")
	require.Equal(t, 0, m.Length(), "length should remain 0 after deleting non-existent key")

	m.Store("key1", "value1")
	m.Store("key2", "value2")

	m.Delete("key1")
	require.Equal(t, 1, m.Length(), "length should be 1 after deleting one key")

	_, ok := m.Load("key1")
	require.False(t, ok, "deleted key should not be found")

	val, ok := m.Load("key2")
	require.True(t, ok, "remaining key should still exist")
	require.Equal(t, "value2", *val, "remaining key should have correct value")

	m.Delete("key1")
	require.Equal(t, 1, m.Length(), "length should remain 1 after deleting same key twice")
}

func TestCountedSyncMap_Length(t *testing.T) {
	t.Parallel()

	m := NewCountedSyncMap[int]()

	require.Equal(t, 0, m.Length(), "initial length should be 0")

	for i := 0; i < 10; i++ {
		m.Store(fmt.Sprintf("key%d", i), i)
		require.Equal(t, i+1, m.Length(), "length should match number of stored items")
	}

	for i := 0; i < 5; i++ {
		m.Delete(fmt.Sprintf("key%d", i))
		expected := 10 - i - 1
		require.Equal(t, expected, m.Length(), "length should decrease after deletion")
	}
}

func TestCountedSyncMap_Range(t *testing.T) {
	t.Parallel()

	m := NewCountedSyncMap[string]()

	called := false
	m.Range(func(key string, value string) bool {
		called = true
		return true
	})
	require.False(t, called, "Range should not call function on empty map")

	expected := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	for k, v := range expected {
		m.Store(k, v)
	}

	found := make(map[string]string)
	m.Range(func(key string, value string) bool {
		found[key] = value
		return true
	})

	require.Equal(t, len(expected), len(found), "Range should iterate over all items")

	for k, v := range expected {
		require.Equal(t, v, found[k], "Range should return correct key-value pairs")
	}

	count := 0
	m.Range(func(key string, value string) bool {
		count++
		return count < 2 // Stop after 2 items
	})

	require.Equal(t, 2, count, "Range should stop when function returns false")
}

func TestCountedSyncMap_Concurrent(t *testing.T) {
	t.Parallel()

	m := NewCountedSyncMap[int]()

	const numGoroutines = 5
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				m.Store(key, id*numOperations+j)
			}
		}(i)
	}

	wg.Wait()

	expectedLength := numGoroutines * numOperations
	require.Equal(t, m.Length(), expectedLength, fmt.Sprintf("expected length %d after concurrent stores, got %d", expectedLength, m.Length()))

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				val, ok := m.Load(key)
				require.True(t, ok, "key should exist")

				expected := id*numOperations + j
				require.Equal(t, *val, expected)
			}
		}(i)
	}

	wg.Wait()

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations/2; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				m.Delete(key)
			}
		}(i)
	}

	wg.Wait()

	expectedLengthAfterDeletes := numGoroutines * numOperations / 2
	require.Equal(t, m.Length(), expectedLengthAfterDeletes)
}
