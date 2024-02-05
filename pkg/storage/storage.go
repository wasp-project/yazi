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

package storage

import (
	"errors"
	"time"

	"github.com/wasp-project/yazi/pkg/policy"
	"github.com/wasp-project/yazi/pkg/policy/lru"
)

type StorageClass string

const (
	StorageClassLocal StorageClass = "local"
	StorageClassS3    StorageClass = "s3"
)

var _ KVStore = &Store{}

type KVStore interface {
	Get(key string) (string, error)
	Set(key, val string) error
	Expire(key string, ttl time.Duration) error
}

type Store struct {
	cache Cache
}

func NewKVStore(cap int, p policy.KeyPolicy) *Store {
	s := &Store{}

	switch p {
	case policy.KeyPolicyLRU:
		s.cache = lru.New[string, string](cap)
	default:
		s.cache = &memcache{
			metadata: cachemeta{
				capacity: cap,
			},
			data: map[string]string{},
		}
	}

	return s
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

func (s *Store) Expire(key string, ttl time.Duration) error {
	s.cache.Expire(key, ttl)
	return nil
}

type Manager struct {
}

func NewManager() *Manager {
	return &Manager{}
}
