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

package server

import (
	"encoding/json"

	"github.com/wasp-project/yazi/pkg/config"
	"github.com/wasp-project/yazi/pkg/protocol"
	"github.com/wasp-project/yazi/pkg/protocol/naive"
	sg "github.com/wasp-project/yazi/pkg/server/grpc"
	sn "github.com/wasp-project/yazi/pkg/server/naive"
	"github.com/wasp-project/yazi/pkg/storage"
	"github.com/wasp-project/yazi/pkg/storage/local"

	"github.com/mlycore/log"
)

type Server struct {
	conf    *config.ServerConfig
	manager *storage.Manager
	ncore   Service
}

func NewServer(conf *config.ServerConfig) *Server {
	s := &Server{conf: conf}

	// init protocol
	switch s.conf.Protocol {
	case protocol.ProtocolNaive:
		s.ncore = sn.NewService(&naive.NaiveCodec{}, s.conf.Port)
	case protocol.ProtocolGrpc:
		fallthrough
	default:
		s.ncore = sg.NewService(s.conf.Port)
	}

	return s
}

func (s *Server) Run() {
	log.SetLevel(s.conf.Loglevel)

	log.Infof("Server is configured with storage: %s", s.conf.Storage)
	log.Infof("Server is configured with policy: %s", s.conf.Policy)
	log.Infof("Server is configured with protocol: %s", s.conf.Protocol)
	log.Infof("Server is configured with port: %d", s.conf.Port)
	log.Infof("Server is configured with raft port: %d", s.conf.RaftPort)
	log.Infof("Server is configured with raft node: %s", s.conf.RaftNode)

	var (
		store      storage.KVStore
		persistent storage.PersistentStorage
	)

	s.manager = storage.NewManager()

	// init storage
	if len(s.conf.Storage) != 0 {
		// decide where to persist
		switch s.conf.Storage {
		case storage.StorageClassLocal:
			persistent = local.NewLocalStorage()
		default:
			log.Infof("Unsupported storage class")
		}

		// set persistent data policy
		// decide how to persist
		// - Append: will persist by kvstore itself
		// - Scheduled: will persist by background task
		switch s.conf.Persistent {
		case storage.PersistentPolicyAppend:
			store = storage.NewKVStoreWithPersistent(s.conf.Capacity, s.conf.Policy, persistent)
		case storage.PersistentPolicyScheduled:
			store = storage.NewKVStore(s.conf.Capacity, s.conf.Policy)
			s.manager.SetTask("persistent", s.manager.Persistent)
		}

		s.manager.SetPersistentStorage(persistent).SetStore(store)
	}

	s.ncore.SetStorage(store)
	s.initServerMetadata()
	s.initRaftMetadata()

	s.manager.SetTask("memory-check", storage.TaskMemoryCheck)

	go s.manager.Run()
	s.ncore.Run()
}

type meta struct {
	Port int `json:"port,omitempty"`
}

func (s *Server) initServerMetadata() error {
	if data, err := json.Marshal(meta{Port: s.conf.Port}); err != nil {
		return err
	} else {
		return s.ncore.SetMeta("", string(data))
	}
}

type raft struct {
	Port int    `json:"port,omitempty"`
	Node string `json:"node,omitempty"`
}

func (s *Server) initRaftMetadata() error {
	if data, err := json.Marshal(raft{Port: s.conf.RaftPort, Node: s.conf.RaftNode}); err != nil {
		return err
	} else {
		return s.ncore.SetRaft("", string(data))
	}
}
