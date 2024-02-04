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
	"time"

	"github.com/wasp-project/yazi/pkg/storage/memory"
)

type StorageClass string

const (
	StorageClassMemory StorageClass = "memory"
	StorageClassS3     StorageClass = "s3"
)

var _ KVStore = &memory.Store{}

type KVStore interface {
	Get(key string) (string, error)
	Set(key, val string) error
	SetWithTTL(key, val string, ttl time.Duration) error
}
