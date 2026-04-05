package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCacheEmptyCache(t *testing.T) {
	c := NewCache(10)

	_, ok := c.Get("aaa")
	require.False(t, ok)

	_, ok = c.Get("bbb")
	require.False(t, ok)
}

func TestCacheSimple(t *testing.T) {
	c := NewCache(5)

	wasInCache := c.Set("aaa", 100)
	require.False(t, wasInCache)

	wasInCache = c.Set("bbb", 200)
	require.False(t, wasInCache)

	val, ok := c.Get("aaa")
	require.True(t, ok)
	require.Equal(t, 100, val)

	val, ok = c.Get("bbb")
	require.True(t, ok)
	require.Equal(t, 200, val)

	wasInCache = c.Set("aaa", 300)
	require.True(t, wasInCache)

	val, ok = c.Get("aaa")
	require.True(t, ok)
	require.Equal(t, 300, val)

	val, ok = c.Get("ccc")
	require.False(t, ok)
	require.Nil(t, val)
}

func TestCachePurgeLogic(t *testing.T) {
	cache := NewCache(2)

	cache.Set("a", 1)
	cache.Set("b", 2)

	cache.Clear()

	val, ok := cache.Get("a")
	require.False(t, ok)
	require.Nil(t, val)

	val, ok = cache.Get("b")
	require.False(t, ok)
	require.Nil(t, val)

	cache.Set("c", 3)

	val, ok = cache.Get("c")
	require.True(t, ok)
	require.Equal(t, 3, val)
}

func TestCacheZeroCapacity(t *testing.T) {
	cache := NewCache(0)

	cache.Set("a", 1)

	val, ok := cache.Get("a")
	require.False(t, ok)
	require.Nil(t, val)
}

func TestCacheExistedValue(t *testing.T) {
	cache := NewCache(3)

	alreadyExisted := cache.Set("a", 1)
	require.False(t, alreadyExisted)

	alreadyExisted = cache.Set("a", 2)
	require.True(t, alreadyExisted)
}

func TestCacheSetSameValue(t *testing.T) {
	cache := NewCache(3)

	cache.Set("a", 1)
	cache.Set("a", 2)

	value, ok := cache.Get("a")
	require.True(t, ok)
	require.Equal(t, 2, value)
}

func TestCacheCapacityOverflow(t *testing.T) {
	cache := NewCache(2)

	cache.Set("a", 1)
	cache.Set("b", 2)
	cache.Set("c", 3)

	_, ok := cache.Get("a")
	require.False(t, ok)

	value, ok := cache.Get("b")
	require.True(t, ok)
	require.Equal(t, 2, value)

	value, ok = cache.Get("c")
	require.True(t, ok)
	require.Equal(t, 3, value)
}

func TestCacheMultithreading(_ *testing.T) {
	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := range 1_000_000 {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for range 1_000_000 {
			c.Get(Key(strconv.Itoa(rand.Intn(1_000_000)))) //nolint:gosec
		}
	}()

	wg.Wait()
}
