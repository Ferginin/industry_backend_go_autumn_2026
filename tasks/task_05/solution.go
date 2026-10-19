package main

type Cache[K comparable, V any] struct {
	capacity int
	items    map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	return &Cache[K, V]{capacity: capacity, items: make(map[K]V)}
}

func (c *Cache[K, V]) Get(k K) (v V, ok bool) {
	val, ok := c.items[k]
	return val, ok
}

func (c *Cache[K, V]) Set(k K, v V) bool {
	if _, ok := c.items[k]; ok {
		c.items[k] = v
		return true
	} else if c.capacity <= len(c.items) {
		return false
	} else {
		c.items[k] = v
		return true
	}
}
