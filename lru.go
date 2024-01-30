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

type MListNode struct {
	Key  string
	Val  int
	Next *MListNode
	Prev *MListNode
}

type LRUCache struct {
	cache    map[string]*MListNode
	capacity int
	size     int
	head     *MListNode
	tail     *MListNode
}

func Constructor(capacity int) LRUCache {
	cache := LRUCache{
		cache:    map[string]*MListNode{},
		capacity: capacity,
		head:     &MListNode{},
		tail:     &MListNode{},
	}
	cache.head.Next = cache.tail
	cache.tail.Prev = cache.head
	return cache

}

func (c *LRUCache) Get(key string) int {
	if _, ok := c.cache[key]; ok {
		c.MoveToHead(c.cache[key])
		return c.cache[key].Val
	}
	return -1
}

func (c *LRUCache) Set(key string, value int) {
	if _, ok := c.cache[key]; !ok {
		p := &MListNode{
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
	} else {
		c.cache[key].Val = value
		c.MoveToHead(c.cache[key])
	}
}

func (c *LRUCache) MoveToHead(p *MListNode) {
	c.Remove(p)
	c.AddToHead(p)
}

func (c *LRUCache) RemoveFromTail() *MListNode {
	p := c.tail.Prev
	c.Remove(p)
	return p
}

func (c *LRUCache) AddToHead(p *MListNode) {
	p.Next = c.head.Next
	c.head.Next = p
	p.Next.Prev = p
	p.Prev = c.head
}

func (c *LRUCache) Remove(p *MListNode) {
	p.Prev.Next = p.Next
	p.Next.Prev = p.Prev
}
