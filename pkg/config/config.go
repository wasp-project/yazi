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

package config

import (
	"github.com/wasp-project/yazi/pkg/policy"
	"github.com/wasp-project/yazi/pkg/storage"
)

type ServerConfig struct {
	Port    int                  `json:"port,omitempty" default:3479`
	Policy  policy.KeyPolicy     `json:"policy,omitempty" default:"lru"`
	Storage storage.StorageClass `json:"storage,omitempty" default:"memory"`
}

func Default() *ServerConfig {
	return &ServerConfig{
		Port:    3479,
		Policy:  policy.KeyPolicyLRU,
		Storage: storage.StorageClassMemory,
	}
}
