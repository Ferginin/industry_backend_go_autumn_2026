package main

import (
	"container/list"
)

type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	// ll не инициализируем, он рабочий даже нулевым значением

	// в мапе храним ссылку на конкретный элемент списка, а так же в Value
	// лежит структура entry, через которую получаем как значение, так и ключ(для удаления из мапы)
	return &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*list.Element),
	}
}

func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	// при отключении сразу возвращаем
	if c.capacity <= 0 {
		return value, ok
	}

	// получаем нужный элемент списка, делаем его самым "свежим" и возвращаем значение
	// .(entry[K, V]) нужен, чтобы определить откуда именно мы хотим достать value
	if elem, ok := c.items[key]; ok {
		c.ll.MoveToFront(elem)
		return elem.Value.(entry[K, V]).value, true
	}

	// иначе вернем промах
	return value, false
}

func (c *LRUCache[K, V]) Set(key K, value V) {
	// при отключении нечего не делаем
	if c.capacity <= 0 {
		return
	}

	// просто меняем значение, если такой ключ был
	// делаем через создание переменной, чтобы смочь изменить значение value
	// иначе при обращении к elem.Value.(entry[K, V]).value мы получаем НЕадресованную копию
	// ее менять не получится
	if elem, ok := c.items[key]; ok {
		newel := elem.Value.(entry[K, V])
		newel.value = value
		elem.Value = newel
		return
	}

	// если мест не осталось, вытесняем самый старый элемент
	// Remove возвращает нам entry, из которого достанем ключ для удаления из мапы
	if c.capacity == len(c.items) {
		oldel := c.ll.Remove(c.ll.Back())
		delete(c.items, oldel.(entry[K, V]).key)
	}

	// и добавляем новый элемент в список и мапу
	newel := c.ll.PushFront(entry[K, V]{
		value: value,
		key:   key,
	})
	c.items[key] = newel
}

type LRU[K comparable, V any] interface {
	Get(K) (V, bool)
	Set(K, V)
}
