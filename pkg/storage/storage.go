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
	"runtime"
	"sync"
	"time"

	"github.com/wasp-project/yazi/pkg/policy"
	"github.com/wasp-project/yazi/pkg/policy/lru"
	"github.com/wasp-project/yazi/pkg/storage/local"

	"github.com/mlycore/log"
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
}

type Store struct {
	cache Cache
}

func TODO() {
	log.Warnf("Not implemented yet")
}

func NewKVStore(capacity int, p policy.KeyPolicy) *Store {
	s := &Store{}

	switch p {
	case policy.KeyPolicyTTL:
		TODO()
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

func (s *Store) Encode() []byte {
	return s.cache.Encode()
}

type Persister interface {
	Write(data []byte) (int, error)
}

var _ Persister = &local.DiskWriter{}

type Manager struct {
	tasks map[string]func()
	p     Persister
	store KVStore
}

func NewManager() *Manager {
	return &Manager{
		tasks: map[string]func(){},
	}
}

func (m *Manager) SetPersistenter(p Persister) *Manager {
	m.p = p
	return m
}

func (m *Manager) SetTask(name string, f func()) *Manager {
	m.tasks[name] = f
	return m
}

func (m *Manager) SetStore(s KVStore) *Manager {
	m.store = s
	return m
}

func (m *Manager) Persistent() {
	ticker := time.NewTicker(10 * time.Second)

	for range ticker.C {
		{
			data := m.store.Encode()
			if n, err := m.p.Write(data); err != nil {
				log.Warnf("Persistent data error: %s", err)
			} else {
				log.Tracef("Persistent data %d bytes", n)
			}
		}
	}
}

func (m *Manager) Run() {
	for _, task := range m.tasks {
		go task()
	}
}

var TaskMemoryCheck = func() {
	ticker := time.NewTicker(1 * time.Second)

	bToMb := func(b uint64) uint64 {
		return b / 1024 / 1024
	}

	for range ticker.C {
		{
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			log.Tracef("Alloc = %v MiB\tTotalAlloc = %v MiB\tSys = %v MiB\tNumGC = %v", bToMb(m.Alloc), bToMb(m.TotalAlloc), bToMb(m.Sys), m.NumGC)
		}
	}
}
