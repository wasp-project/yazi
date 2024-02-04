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
	"time"

	"github.com/wasp-project/yazi/pkg/policy"
	"github.com/wasp-project/yazi/pkg/policy/lru"
)

type Store struct {
	policy policy.KeyPolicy
	cache  lru.LRUCache[string, string]
}

func New() *Store {
	return &Store{
		policy: policy.KeyPolicyLRU,
		cache:  lru.New[string, string](8192),
	}
}

func (s *Store) Get(key string) (string, error) {
	v, ok := s.cache.Get(key)
	if !ok {
		return v, errors.New("key not found")
	}
	return v, nil
}

func (s *Store) Set(key, value string) error {
	s.cache.Set(key, value, 0)
	return nil
}

func (s *Store) SetWithTTL(key, value string, ttl time.Duration) error {
	s.cache.Set(key, value, ttl)
	return nil
}