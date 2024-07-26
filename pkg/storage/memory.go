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
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/wasp-project/yazi/pkg/utils"
)

type Cache interface {
	Get(key string) (val string, gotten bool)
	Set(key, val string) (prev string, replaced bool)
	Expire(key string, ttl time.Duration) (updated bool)

	Encode() []byte
	Decode([]byte) error
}

type memcache struct {
	metadata cachemeta
	data     map[string]string
	lock     sync.Mutex
}

type cachemeta struct {
	capacity  int
	size      int
	maxmemory int
}

func (c *memcache) SetCapacity(capacity int) {
	c.metadata.capacity = capacity
}

func (c *memcache) Get(key string) (val string, gotten bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	v, ok := c.data[key]
	return v, ok
}

func (c *memcache) Set(key string, val string) (prev string, replaced bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	prev, replaced = c.data[key]
	c.data[key] = val
	return prev, replaced
}

func (c *memcache) Expire(key string, ttl time.Duration) bool {
	utils.TODO()
	return false
}

func (c *memcache) Encode() []byte {
	c.lock.Lock()
	defer c.lock.Unlock()
	d, _ := json.Marshal(c.data)
	return d
}

func (c *memcache) Decode(data []byte) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	// FIXME: the data loaded from file will contains invisible char "\x00"
	// they are replaced by the line below temporarily
	data = []byte(strings.ReplaceAll(string(data), "\x00", ""))
	return json.Unmarshal(data, &c.data)
}
