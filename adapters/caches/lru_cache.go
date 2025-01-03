package caches

import (
	"container/list"
	"context"
	"sync"

	"github.com/kevin07696/receipt-processor/domain"
)

type LRUCache struct {
	cache    *sync.Map
	lruList  *LRUList
	capacity int
}

func NewLRUCache(capacity int) LRUCache {
	return LRUCache{
		cache:    &sync.Map{},
		lruList:  NewLRUList(),
		capacity: capacity,
	}
}

type entry struct {
	key   string
	value interface{}
}

func (c LRUCache) Get(ctx context.Context, key string) (interface{}, domain.StatusCode) {
	if elem, ok := c.cache.Load(key); ok {
		// Move the accessed item to the front of the LRU list
		c.lruList.MoveToFront(elem.(*list.Element))
		return elem.(*list.Element).Value.(*entry).value, domain.StatusOK
	}
	return nil, domain.ErrNotFound
}

func (c *LRUCache) Set(ctx context.Context, key string, value interface{}) domain.StatusCode {
	if elem, ok := c.cache.Load(key); ok {
		// Update the value and move the item to the front of the LRU list
		c.lruList.MoveToFront(elem.(*list.Element))
		elem.(*list.Element).Value.(*entry).value = value
		return domain.StatusOK
	}

	// Add the new item to the front of the LRU list
	newElem := c.lruList.PushFront(&entry{key: key, value: value})
	c.cache.Store(key, newElem)

	// Evict the least recently used item if the cache exceeds its capacity
	if c.lruList.Len() > c.capacity {
		backElem := c.lruList.Back()
		if backElem != nil {
			c.lruList.Remove(backElem)
			c.cache.Delete(backElem.Value.(*entry).key)
		}
	}
	return domain.StatusOK
}

type LRUList struct {
	mu      sync.Mutex
	lruList *list.List
}

func NewLRUList() *LRUList {
	return &LRUList{
		lruList: list.New(),
	}
}

func (l *LRUList) Lock() {
	l.mu.Lock()
}

func (l *LRUList) Unlock() {
	l.mu.Unlock()
}

func (l *LRUList) PushFront(value interface{}) *list.Element {
	l.Lock()
	defer l.Unlock()
	return l.lruList.PushFront(value)
}

func (l *LRUList) MoveToFront(elem *list.Element) {
	l.Lock()
	defer l.Unlock()
	l.lruList.MoveToFront(elem)
}

func (l *LRUList) Remove(elem *list.Element) {
	l.Lock()
	defer l.Unlock()
	l.lruList.Remove(elem)
}

func (l *LRUList) Back() *list.Element {
	l.Lock()
	defer l.Unlock()
	return l.lruList.Back()
}

func (l *LRUList) Len() int {
	l.Lock()
	defer l.Unlock()
	return l.lruList.Len()
}
