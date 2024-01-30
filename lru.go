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

package main

import "time"

type MListNode[K comparable, V any] struct {
	Key  K
	Val  V
	Next *MListNode[K, V]
	Prev *MListNode[K, V]
}

type LRUCache[K comparable, V any] struct {
	cache    map[K]*MListNode[K, V]
	capacity int
	size     int
	head     *MListNode[K, V]
	tail     *MListNode[K, V]
}

func Constructor[K comparable, V any](capacity int) LRUCache[K, V] {
	cache := LRUCache[K, V]{
		cache:    map[K]*MListNode[K, V]{},
		capacity: capacity,
		head:     &MListNode[K, V]{},
		tail:     &MListNode[K, V]{},
	}
	cache.head.Next = cache.tail
	cache.tail.Prev = cache.head
	return cache
}

func (c *LRUCache[K, V]) Get(key K) (val V, gotten bool) {
	if _, ok := c.cache[key]; ok {
		c.MoveToHead(c.cache[key])
		return c.cache[key].Val, ok
	}

	return val, false
}

// TODO: implement ttl
func (c *LRUCache[K, V]) Set(key K, value V, ttl time.Duration) (prev V, replaced bool) {
	if _, ok := c.cache[key]; !ok {
		p := &MListNode[K, V]{
			Key: key,
			Val: value,
		}
		c.AddToHead(p)
		c.cache[key] = p
		c.size++
		if c.size > c.capacity {
			t := c.RemoveFromTail()
			delete(c.cache, t.Key)
			c.size--
		}
		replaced = false
		return
	} else {
		prev = c.cache[key].Val
		replaced = true
		c.cache[key].Val = value
		c.MoveToHead(c.cache[key])
		return
	}
}

func (c *LRUCache[K, V]) MoveToHead(p *MListNode[K, V]) {
	c.Remove(p)
	c.AddToHead(p)
}

func (c *LRUCache[K, V]) RemoveFromTail() *MListNode[K, V] {
	p := c.tail.Prev
	c.Remove(p)
	return p
}

func (c *LRUCache[K, V]) AddToHead(p *MListNode[K, V]) {
	p.Next = c.head.Next
	c.head.Next = p
	p.Next.Prev = p
	p.Prev = c.head
}

func (c *LRUCache[K, V]) Remove(p *MListNode[K, V]) {
	p.Prev.Next = p.Next
	p.Next.Prev = p.Prev
}
