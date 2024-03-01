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

	// init storage
	store := storage.NewKVStore(s.conf.Capacity, s.conf.Policy)
	s.ncore.SetStorage(store)

	switch s.conf.Storage {
	case storage.StorageClassLocal:
		w := local.NewWriter()
		s.manager = storage.NewManager().SetPersistenter(w).SetStore(store)
		s.manager.SetTask("persistent", s.manager.Persistent)
	default:
		s.manager = storage.NewManager().SetStore(store)
	}

	s.manager.SetTask("memory-check", storage.TaskMemoryCheck)

	go s.manager.Run()
	s.ncore.Run()
}
