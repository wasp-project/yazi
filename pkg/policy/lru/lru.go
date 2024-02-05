// Copyright 2024 mlycore. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package lru

import "time"

type MListNode[K comparable, V any] struct {
	metadata NodeMeta[K, V]
	data     NodeData[K, V]
}

type NodeData[K comparable, V any] struct {
	key K
	val V
}

type NodeMeta[K comparable, V any] struct {
	next *MListNode[K, V]
	prev *MListNode[K, V]
	ttl  *time.Time
}

type LRUCache[K comparable, V any] struct {
	cache    map[K]*MListNode[K, V]
	capacity int
	size     int
	head     *MListNode[K, V]
	tail     *MListNode[K, V]
}

func New[K comparable, V any](capacity int) LRUCache[K, V] {
	cache := LRUCache[K, V]{
		cache:    map[K]*MListNode[K, V]{},
		capacity: capacity,
		head:     &MListNode[K, V]{},
		tail:     &MListNode[K, V]{},
	}
	cache.head.metadata.next = cache.tail
	cache.tail.metadata.prev = cache.head
	return cache
}

func (c *LRUCache[K, V]) Get(key K) (val V, gotten bool) {
	if v, ok := c.cache[key]; ok {
		if v.metadata.ttl != nil && v.metadata.ttl.Before(time.Now()) {
			c.Remove(c.cache[key])
			delete(c.cache, key)
			return val, false
		} else {
			c.MoveToHead(c.cache[key])
			return v.data.val, ok
		}
	}

	return val, false
}

func (c *LRUCache[K, V]) Set(key K, value V, ttl time.Duration) (prev V, replaced bool) {
	if _, ok := c.cache[key]; !ok {
		p := &MListNode[K, V]{
			data: NodeData[K, V]{
				key: key,
				val: value,
			},
		}
		if ttl != 0 {
			c := time.Now().Add(ttl)
			p.metadata.ttl = &c
		}
		c.AddToHead(p)
		c.cache[key] = p
		c.size++
		if c.size > c.capacity {
			t := c.RemoveFromTail()
			delete(c.cache, t.data.key)
			c.size--
		}
		replaced = false
		return
	} else {
		prev = c.cache[key].data.val
		replaced = true
		c.cache[key].data.val = value
		c.MoveToHead(c.cache[key])
		return
	}
}

func (c *LRUCache[K, V]) MoveToHead(p *MListNode[K, V]) {
	c.Remove(p)
	c.AddToHead(p)
}

func (c *LRUCache[K, V]) RemoveFromTail() *MListNode[K, V] {
	p := c.tail.metadata.prev
	c.Remove(p)
	return p
}

func (c *LRUCache[K, V]) AddToHead(p *MListNode[K, V]) {
	p.metadata.next = c.head.metadata.next
	c.head.metadata.next = p
	p.metadata.next.metadata.prev = p
	p.metadata.prev = c.head
}

func (c *LRUCache[K, V]) Remove(p *MListNode[K, V]) {
	p.metadata.prev.metadata.next = p.metadata.next
	p.metadata.next.metadata.prev = p.metadata.prev
}
