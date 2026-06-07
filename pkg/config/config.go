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
	Port            int                      `json:"port,omitempty" default:"3456"`
	Protocol        protocol.Protocol        `json:"protocol,omitempty" default:"naive"`
	Policy          policy.KeyPolicy         `json:"policy,omitempty" default:"none"`
	Storage         storage.StorageClass     `json:"storage,omitempty" default:"memory"`
	Engine          storage.Engine           `json:"engine,omitempty" default:"mem"`
	Persistent      storage.PersistentPolicy `json:"persistent,omitempty" default:"append"`
	ScheduledPeriod int                      `json:"scheduledPeriod,omitempty" yaml:"scheduledPeriod" default:"10"`
	Capacity        int                      `json:"capacity,omitempty" default:"1024"`
	Loglevel        string                   `json:"loglevel,omitempty" default:"info"`
	RaftPort        int                      `json:"raftPort,omitempty"`
	RaftNode        string                   `json:"raftNode,omitempty"`
	Experimental    ExperimentalConfig       `json:"experimental,omitempty"`
	LSM             LSMConfig                `json:"lsm,omitempty" yaml:"lsm"`
	Replication     ReplicationConfig        `json:"replication,omitempty" yaml:"replication"`
	S3              S3Config                 `json:"s3,omitempty" yaml:"s3"`
	Tenant          TenantConfig             `json:"tenant,omitempty" yaml:"tenant"`
}

// TenantConfig configures the tenant-service layer. When absent (the zero
// value), the system uses the no-op bypass service and behaves as a
// single-tenant store. Set Mode to "static" with a Tenant id to namespace all
// memory under one tenant; an external project may supply richer modes.
type TenantConfig struct {
	// Mode selects the tenant service: "" or "bypass" (default) or "static".
	Mode string `json:"mode,omitempty" yaml:"mode"`
	// Tenant is the tenant id used when Mode is "static".
	Tenant string `json:"tenant,omitempty" yaml:"tenant"`
}

// S3Config configures the AWS S3 (or S3-compatible) cloud persistence backend.
// It is read only when Storage is "s3".
type S3Config struct {
	Bucket    string `json:"bucket,omitempty" yaml:"bucket"`
	Region    string `json:"region,omitempty" yaml:"region"`
	Key       string `json:"key,omitempty" yaml:"key"`
	Endpoint  string `json:"endpoint,omitempty" yaml:"endpoint"`
	AccessKey string `json:"accessKey,omitempty" yaml:"accessKey"`
	SecretKey string `json:"secretKey,omitempty" yaml:"secretKey"`
}

type ExperimentalConfig struct {
	RaftPort int    `json:"raftPort,omitempty" yaml:"raftPort"`
	RaftNode string `json:"raftNode,omitempty" yaml:"raftNode"`
	Capacity int    `json:"capacity,omitempty" default:"1024"`
	Buffer   int    `json:"buffer,omitempty" default:"1024"`
}

type LSMConfig struct {
	Dir                  string `json:"dir,omitempty" yaml:"dir"`
	MemtableMaxEntries   int    `json:"memtableMaxEntries,omitempty" yaml:"memtableMaxEntries"`
	CompactionMaxTables  int    `json:"compactionMaxTables,omitempty" yaml:"compactionMaxTables"`
	WALMaxSegmentEntries int    `json:"walMaxSegmentEntries,omitempty" yaml:"walMaxSegmentEntries"`
}

type ReplicationConfig struct {
	Enabled       bool     `json:"enabled,omitempty" yaml:"enabled"`
	Peers         []string `json:"peers,omitempty" yaml:"peers"`
	WriteQuorum   int      `json:"writeQuorum,omitempty" yaml:"writeQuorum"`
	ReadQuorum    int      `json:"readQuorum,omitempty" yaml:"readQuorum"`
	TimeoutMillis int      `json:"timeoutMillis,omitempty" yaml:"timeoutMillis"`
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
			case "Engine":
				conf.Engine = storage.Engine(tag)
			case "Capacity":
				c, _ := strconv.ParseInt(tag, 10, 64)
				conf.Capacity = int(c)
			case "Loglevel":
				conf.Loglevel = tag
			case "Persistent":
				conf.Persistent = storage.PersistentPolicy(tag)
			case "ScheduledPeriod":
				sp, _ := strconv.ParseInt(tag, 10, 64)
				conf.ScheduledPeriod = int(sp)
			}
		}
	}

	if conf.LSM.Dir == "" {
		conf.LSM.Dir = "data/lsm"
	}
	if conf.LSM.MemtableMaxEntries == 0 {
		conf.LSM.MemtableMaxEntries = 1024
	}
	if conf.LSM.CompactionMaxTables == 0 {
		conf.LSM.CompactionMaxTables = 4
	}
	if conf.LSM.WALMaxSegmentEntries == 0 {
		conf.LSM.WALMaxSegmentEntries = 0
	}
	if conf.Replication.TimeoutMillis == 0 {
		conf.Replication.TimeoutMillis = 500
	}
	if conf.Replication.WriteQuorum == 0 {
		conf.Replication.WriteQuorum = 1
	}
	if conf.Replication.ReadQuorum == 0 {
		conf.Replication.ReadQuorum = 1
	}

	if conf.S3.Region == "" {
		conf.S3.Region = "us-east-1"
	}
	if conf.S3.Key == "" {
		conf.S3.Key = "yazi.data"
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

	// S3 credentials may come from the standard AWS environment variables so
	// they need not be written into the config file.
	if c.S3.AccessKey == "" {
		c.S3.AccessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	if c.S3.SecretKey == "" {
		c.S3.SecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}

	return c
}
