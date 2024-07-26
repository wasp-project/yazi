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

import (
	"time"

	"github.com/wasp-project/yazi/pkg/utils"
)

type MListNode[K comparable, V any] struct {
	metadata NodeMeta[K, V]
	data     NodeData[K, V]
}

func (n *MListNode[K, V]) isExpired() bool {
	return n.metadata.ttl != nil && n.metadata.ttl.Before(time.Now())
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

func New[K comparable, V any](capacity int) *LRUCache[K, V] {
	cache := &LRUCache[K, V]{
		cache:    map[K]*MListNode[K, V]{},
		capacity: capacity,
		head:     &MListNode[K, V]{},
		tail:     &MListNode[K, V]{},
	}
	cache.head.metadata.next = cache.tail
	cache.tail.metadata.prev = cache.head
	return cache
}

func (c *LRUCache[K, V]) get(key K) *MListNode[K, V] {
	if n, ok := c.cache[key]; ok {
		if n.isExpired() {
			c.Remove(n)
			delete(c.cache, key)
			return nil
		} else {
			return n
		}
	}
	return nil
}

func (c *LRUCache[K, V]) Get(key K) (val V, gotten bool) {
	n := c.get(key)
	if n != nil {
		c.MoveToHead(n)
		return n.data.val, true
	}

	return val, false
}

func (c *LRUCache[K, V]) Set(key K, value V) (prev V, replaced bool) {
	n := c.get(key)
	if n != nil {
		prev = n.data.val
		replaced = true
		n.data.val = value
		c.MoveToHead(n)
		return
	} else {
		p := &MListNode[K, V]{
			data: NodeData[K, V]{
				key: key,
				val: value,
			},
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
	}
}

func (c *LRUCache[K, V]) Expire(key K, ttl time.Duration) (updated bool) {
	if n, ok := c.cache[key]; ok {
		updated = n.metadata.ttl != nil
		cur := time.Now().Add(ttl)
		c.cache[key].metadata.ttl = &cur
	}
	return
}

func (c *LRUCache[K, V]) Encode() []byte {
	utils.TODO()
	return nil
}

func (c *LRUCache[K, V]) Decode(data []byte) error {
	utils.TODO()
	return nil
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
