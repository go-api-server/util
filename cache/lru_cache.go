package cache

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

type LRUCache struct {
	mutex    sync.Mutex
	list     *list.List
	table    map[string]*list.Element
	size     int64
	capacity int64
}

type Value interface {
	Size() int
}

type Item struct {
	Key   string
	Value Value
}

type entry struct {
	key      string
	value    Value
	size     int64
	accessAt time.Time
	expireAt int64
}

func NewLRUCache(capacity int64) *LRUCache {
	return &LRUCache{
		list:     list.New(),
		table:    make(map[string]*list.Element),
		capacity: capacity,
	}
}

func (lru *LRUCache) Get(key string) (v Value, ok bool) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	element := lru.table[key]
	if element == nil {
		return nil, false
	}

	res := element.Value.(*entry)
	if res.expireAt > 0 && res.expireAt < time.Now().Unix() {
		lru.list.Remove(element)
		delete(lru.table, key)
		lru.size -= res.size
		return nil, false
	}

	lru.moveToFront(element)
	return res.value, true
}

func (lru *LRUCache) Set(key string, value Value) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if element := lru.table[key]; element != nil {
		lru.updateInplace(element, value)
	} else {
		lru.addNew(key, value, -1)
	}
}

func (lru *LRUCache) SetEX(key string, value Value, cacheTime int64) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if element := lru.table[key]; element != nil {
		lru.updateInplace(element, value)
	} else {
		lru.addNew(key, value, cacheTime)
	}
}

func (lru *LRUCache) SetIfAbsent(key string, value Value) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if element := lru.table[key]; element != nil {
		lru.moveToFront(element)
	} else {
		lru.addNew(key, value, -1)
	}
}

func (lru *LRUCache) SetIfAbsentEX(key string, value Value, cacheTime int64) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	if element := lru.table[key]; element != nil {
		lru.moveToFront(element)
	} else {
		lru.addNew(key, value, cacheTime)
	}
}

func (lru *LRUCache) Delete(key string) bool {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	element := lru.table[key]
	if element == nil {
		return false
	}

	lru.list.Remove(element)
	delete(lru.table, key)
	lru.size -= element.Value.(*entry).size
	return true
}

func (lru *LRUCache) Clear() {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	lru.list.Init()
	lru.table = make(map[string]*list.Element)
	lru.size = 0
}

func (lru *LRUCache) SetCapacity(capacity int64) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	lru.capacity = capacity
	lru.checkCapacity()
}

func (lru *LRUCache) Stats() (length, size, capacity int64, oldest time.Time) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()
	if lastElem := lru.list.Back(); lastElem != nil {
		oldest = lastElem.Value.(*entry).accessAt
	}
	return int64(lru.list.Len()), lru.size, lru.capacity, oldest
}

func (lru *LRUCache) StatsJSON() string {
	if lru == nil {
		return "{}"
	}
	l, s, c, o := lru.Stats()
	return fmt.Sprintf("{\"Length\": %v, \"Size\": %v, \"Capacity\": %v, \"OldestAccess\": \"%v\"}", l, s, c, o)
}

func (lru *LRUCache) Length() int64 {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()
	return int64(lru.list.Len())
}

func (lru *LRUCache) Size() int64 {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()
	return lru.size
}

func (lru *LRUCache) Capacity() int64 {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()
	return lru.capacity
}

func (lru *LRUCache) Oldest() (oldest time.Time) {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()
	if lastElem := lru.list.Back(); lastElem != nil {
		oldest = lastElem.Value.(*entry).accessAt
	}
	return
}

func (lru *LRUCache) Keys() []string {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	keys := make([]string, 0, lru.list.Len())
	for e := lru.list.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*entry).key)
	}
	return keys
}

func (lru *LRUCache) Items() []Item {
	lru.mutex.Lock()
	defer lru.mutex.Unlock()

	items := make([]Item, 0, lru.list.Len())
	for e := lru.list.Front(); e != nil; e = e.Next() {
		v := e.Value.(*entry)
		items = append(items, Item{Key: v.key, Value: v.value})
	}
	return items
}

func (lru *LRUCache) updateInplace(element *list.Element, value Value) {
	valueSize := int64(value.Size())
	sizeDiff := valueSize - element.Value.(*entry).size
	element.Value.(*entry).value = value
	element.Value.(*entry).size = valueSize
	lru.size += sizeDiff
	lru.moveToFront(element)
	lru.checkCapacity()
}

func (lru *LRUCache) moveToFront(element *list.Element) {
	lru.list.MoveToFront(element)
	element.Value.(*entry).accessAt = time.Now()
}

func (lru *LRUCache) addNew(key string, value Value, cacheTime int64) {
	var expireAt int64 = -1
	if cacheTime > 0 {
		expireAt = time.Now().Unix() + cacheTime
	}
	newEntry := &entry{key, value, int64(value.Size()), time.Now(), expireAt}
	element := lru.list.PushFront(newEntry)
	lru.table[key] = element
	lru.size += newEntry.size
	lru.checkCapacity()
}

func (lru *LRUCache) checkCapacity() {
	for lru.size > lru.capacity {
		delElem := lru.list.Back()
		delValue := delElem.Value.(*entry)
		lru.list.Remove(delElem)
		delete(lru.table, delValue.key)
		lru.size -= delValue.size
	}
}
