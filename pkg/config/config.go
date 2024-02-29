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
	"os"

	"github.com/wasp-project/yazi/pkg/policy"
	"github.com/wasp-project/yazi/pkg/protocol"
	"github.com/wasp-project/yazi/pkg/storage"

	"github.com/mlycore/log"
	"gopkg.in/yaml.v2"
)

type ServerConfig struct {
	Port     int                  `json:"port,omitempty" default:"3479"`
	Protocol protocol.Protocol    `json:"protocol,omitempty" default:"naive"`
	Policy   policy.KeyPolicy     `json:"policy,omitempty" default:""`
	Storage  storage.StorageClass `json:"storage,omitempty" default:"memory"`
	Capacity int                  `json:"capacity,omitempty" default:"1024"`
	Loglevel string               `json:"loglevel,omitempty" default:"info"`
}

func Default() *ServerConfig {
	return &ServerConfig{
		Port:     3456,
		Protocol: protocol.ProtocolNaive,
		Capacity: 1024,
		Loglevel: "info",
	}
}

func (c *ServerConfig) Load(path string) *ServerConfig {
	if data, err := os.ReadFile(path); err != nil {
		log.Errorf("Read config file error: %s", err)
	} else {
		if err := yaml.Unmarshal(data, c); err != nil {
			log.Errorf("Read config file error: %s", err)
		}
	}

	return c
}
