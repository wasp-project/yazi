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
	"github.com/wasp-project/yazi/pkg/storage"
	"github.com/wasp-project/yazi/pkg/storage/memory"

	"github.com/mlycore/log"
)

type Server struct {
	conf  *config.ServerConfig
	store storage.KVStore
}

func NewServer(conf *config.ServerConfig) *Server {
	return &Server{conf: conf}
}

func (s *Server) Run() {
	log.Infof("Server is configured with storage: %s", s.conf.Storage)
	log.Infof("Server is configured with policy: %s", s.conf.Policy)

	switch s.conf.Storage {
	case storage.StorageClassMemory:
		fallthrough
	default:
		s.store = memory.New()
	}

	log.Infoln("Server is running...")
}
