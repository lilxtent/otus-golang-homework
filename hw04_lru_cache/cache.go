package hw04lrucache

import "sync"

type Key string

type cacheEntry struct {
	key   Key
	value any
}

type Cache interface {
	Set(key Key, value any) bool
	Get(key Key) (any, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
	mutex    sync.Mutex
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
		mutex:    sync.Mutex{},
	}
}

func (cache *lruCache) Set(key Key, value any) bool {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if item, ok := cache.items[key]; ok {
		entry := item.Value.(cacheEntry)
		entry.value = value
		item.Value = entry
		cache.queue.MoveToFront(item)

		return true
	}

	if cache.capacity == 0 {
		return false
	}

	if cache.queue.Len() == cache.capacity {
		lastItem := cache.queue.Back()
		lastKey := lastItem.Value.(cacheEntry).key

		cache.queue.Remove(lastItem)
		delete(cache.items, lastKey)
	}

	newItem := cache.queue.PushFront(cacheEntry{key: key, value: value})
	cache.items[key] = newItem

	return false
}

func (cache *lruCache) Get(key Key) (any, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if item, ok := cache.items[key]; ok {
		cache.queue.MoveToFront(item)
		entry := item.Value.(cacheEntry)

		return entry.value, true
	}

	return nil, false
}

func (cache *lruCache) Clear() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.items = make(map[Key]*ListItem, cache.capacity)
	cache.queue = NewList()
}
