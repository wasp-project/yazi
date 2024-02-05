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

package memory

import (
	"errors"

	"github.com/wasp-project/yazi/pkg/policy"
)

type Cache interface {
	Get(key string) (val string, gotten bool)
	Set(key, val string) (prev string, replaced bool)
}

var _ Cache = &memcache{}

type memcache struct {
	metadata cachemeta
	data     map[string]string
}

type cachemeta struct {
	capacity  int
	size      int
	maxmemory int
}

const (
	defaultCapacity = 1024
)

func (c *memcache) SetCapacity(cap int) {
	c.metadata.capacity = cap
}

func (c *memcache) Get(key string) (val string, gotten bool) {
	v, ok := c.data[key]
	return v, ok
}

func (c *memcache) Set(key string, val string) (prev string, replaced bool) {
	prev, replaced = c.data[key]
	c.data[key] = val
	return prev, replaced
}

type Store struct {
	policy policy.KeyPolicy
	cache  Cache
}

func New() *Store {
	return &Store{
		cache: &memcache{
			metadata: cachemeta{
				capacity: defaultCapacity,
			},
			data: map[string]string{},
		},
	}
}

func (s *Store) SetPolicy(p policy.KeyPolicy) {
	s.policy = p
}

func (s *Store) SetStorageCapacity(cap int) {
}

func (s *Store) Get(key string) (string, error) {
	v, ok := s.cache.Get(key)
	if !ok {
		return v, errors.New("key not found")
	}
	return v, nil
}

func (s *Store) Set(key, value string) error {
	s.cache.Set(key, value)
	return nil
}
