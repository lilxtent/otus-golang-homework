package hw04lrucache

import "sync"

type Key string

type Cache interface {
	Set(key Key, value any) bool
	Get(key Key) (any, bool)
	Clear()
}

type lruCache struct {
	capacity  int
	queue     List
	items     map[Key]*ListItem
	itemToKey map[*ListItem]Key
	mutex     sync.Mutex
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity:  capacity,
		queue:     NewList(),
		items:     make(map[Key]*ListItem, capacity),
		itemToKey: make(map[*ListItem]Key, capacity),
		mutex:     sync.Mutex{},
	}
}

func (cache *lruCache) Set(key Key, value any) bool {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	existedValue, keyAlreadyExisted := cache.items[key]

	if keyAlreadyExisted {
		existedValue.Value = value
		cache.queue.MoveToFront(existedValue)

		return true
	}

	newItem := cache.queue.PushFront(value)
	cache.items[key] = newItem
	cache.itemToKey[newItem] = key

	if cache.queue.Len() > cache.capacity {
		lastItem := cache.queue.Back()
		lastItemKey := cache.itemToKey[lastItem]

		cache.queue.Remove(lastItem)
		delete(cache.itemToKey, lastItem)
		delete(cache.items, lastItemKey)
	}

	return false
}

func (cache *lruCache) Get(key Key) (any, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if item, ok := cache.items[key]; ok {
		cache.queue.MoveToFront(item)

		return item.Value, true
	}

	return nil, false
}

func (cache *lruCache) Clear() {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.items = make(map[Key]*ListItem, cache.capacity)
	cache.itemToKey = make(map[*ListItem]Key, cache.capacity)
	cache.queue = NewList()
}
