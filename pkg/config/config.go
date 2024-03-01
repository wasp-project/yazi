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
	"reflect"
	"strconv"

	"github.com/wasp-project/yazi/pkg/policy"
	"github.com/wasp-project/yazi/pkg/protocol"
	"github.com/wasp-project/yazi/pkg/storage"

	"github.com/mlycore/log"
	"gopkg.in/yaml.v2"
)

type ServerConfig struct {
	Port       int                      `json:"port,omitempty" default:"3456"`
	Protocol   protocol.Protocol        `json:"protocol,omitempty" default:"naive"`
	Policy     policy.KeyPolicy         `json:"policy,omitempty" default:"none"`
	Storage    storage.StorageClass     `json:"storage,omitempty" default:"memory"`
	Persistent storage.PersistentPolicy `json:"persistent,omitempty" default:"append"`
	Capacity   int                      `json:"capacity,omitempty" default:"1024"`
	Loglevel   string                   `json:"loglevel,omitempty" default:"info"`
	RaftPort   int                      `json:"raftPort,omitempty"`
	RaftNode   string                   `json:"raftNode,omitempty"`
}

func Default() *ServerConfig {
	conf := &ServerConfig{}
	for i := 0; i < reflect.TypeOf(ServerConfig{}).NumField(); i++ {
		field := reflect.TypeOf(ServerConfig{}).Field(i)
		if tag := field.Tag.Get("default"); tag != "" {
			switch field.Name {
			case "Port":
				p, _ := strconv.ParseInt(tag, 10, 64)
				conf.Port = int(p)
			case "Protocol":
				conf.Protocol = protocol.Protocol(tag)
			case "Policy":
				conf.Policy = policy.KeyPolicy(tag)
			case "Storage":
				conf.Storage = storage.StorageClass(tag)
			case "Capacity":
				c, _ := strconv.ParseInt(tag, 10, 64)
				conf.Capacity = int(c)
			case "Loglevel":
				conf.Loglevel = tag
			case "Persistent":
				conf.Persistent = storage.PersistentPolicy(tag)
			}
		}
	}

	return conf
}

func (c *ServerConfig) Load(path string) *ServerConfig {
	if data, err := os.ReadFile(path); err != nil {
		log.Errorf("Read config file error: %s", err)
	} else {
		if err := yaml.Unmarshal(data, c); err != nil {
			log.Errorf("Read config file error: %s", err)
		}
	}

	// check if the config is overrided by env
	if os.Getenv("PORT") != "" {
		if p, err := strconv.ParseInt(os.Getenv("PORT"), 10, 64); err == nil {
			c.Port = int(p)
		}
	}

	if os.Getenv("RAFT_PORT") != "" {
		if p, err := strconv.ParseInt(os.Getenv("RAFT_PORT"), 10, 64); err == nil {
			c.RaftPort = int(p)
		}
	}

	if os.Getenv("RAFT_NODE") != "" {
		c.RaftNode = os.Getenv("RAFT_NODE")
	}

	return c
}
