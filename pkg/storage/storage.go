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
	"strings"
	"sync"
	"time"

	"github.com/wasp-project/yazi/pkg/policy"
	"github.com/wasp-project/yazi/pkg/policy/lru"
	"github.com/wasp-project/yazi/pkg/storage/local"
	"github.com/wasp-project/yazi/pkg/utils"
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

	Encode() []byte
	Decode([]byte) error
}

type Store struct {
	cache      Cache
	persistent PersistentStorage
}

func NewKVStore(capacity int, kp policy.KeyPolicy) *Store {
	s := &Store{}

	// set key policy
	switch kp {
	case policy.KeyPolicyTTL:
		utils.TODO()
	case policy.KeyPolicyLRU:
		s.cache = lru.New[string, string](capacity)
	default:
		s.cache = &memcache{
			metadata: cachemeta{
				capacity:  capacity,
				size:      0,
				maxmemory: 2048,
			},
			data: map[string]string{},
			lock: sync.Mutex{},
		}
	}

	return s
}

type PersistentStorage interface {
	Write(data []byte) (int, error)
	Read(data []byte) (int, error)
}

func NewKVStoreWithPersistent(capacity int, kp policy.KeyPolicy, ps PersistentStorage) *Store {
	s := NewKVStore(capacity, kp)

	if ps != nil {
		s.persistent = ps
	}

	return s
}

func (s *Store) Get(key string) (string, error) {
	v, ok := s.cache.Get(strings.TrimSpace(key))
	if !ok {
		return v, errors.New("key not found")
	}
	return v, nil
}

func (s *Store) Set(key, value string) error {
	s.cache.Set(strings.TrimSpace(key), value)
	if s.persistent != nil {
		_, err := s.persistent.Write(s.cache.Encode())
		return err
	}
	return nil
}

func (s *Store) Expire(key string, ttl time.Duration) error {
	s.cache.Expire(strings.TrimSpace(key), ttl)
	return nil
}

func (s *Store) Encode() []byte {
	return s.cache.Encode()
}

func (s *Store) Decode(data []byte) error {
	return s.cache.Decode(data)
}

type Persister interface {
	Write(data []byte) (int, error)
}

var _ Persister = &local.DiskWriter{}

type PersistentPolicy string

const (
	PersistentPolicyAppend    PersistentPolicy = "append"
	PersistentPolicyScheduled PersistentPolicy = "scheduled"
)

type Loader interface {
	Read([]byte) (int, error)
}

var _ Loader = &local.DiskReader{}
