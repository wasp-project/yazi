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
	"runtime"
	"time"

	"github.com/mlycore/log"
)

type Manager struct {
	tasks  map[string]func()
	p      PersistentStorage
	config PersistentConfig
	store  KVStore
}

func NewManager() *Manager {
	return &Manager{
		tasks: map[string]func(){},
	}
}

func (m *Manager) SetPersistentStorage(ps PersistentStorage) *Manager {
	if ps != nil {
		m.p = ps
	}
	return m
}

func (m *Manager) SetPersistentConfig(conf *PersistentConfig) *Manager {
	m.config = *conf
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
	ticker := time.NewTicker(time.Duration(m.config.ScheduledPeriod) * time.Second)

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

func (m *Manager) Load() {
	var data = []byte{}

	data = make([]byte, m.config.BufferInKiloBytes*1024)
	if n, err := m.p.Read(data); err != nil {
		log.Warnf("Persistent load data error: %s", err)
	} else {
		log.Tracef("Persistent load data %d bytes", n)
	}
	m.store.Decode(data)
}

type PersistentConfig struct {
	ScheduledPeriod   int
	BufferInKiloBytes int
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
