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
	"net"
	"strconv"
	"time"

	"github.com/wasp-project/yazi/pkg/config"
	"github.com/wasp-project/yazi/pkg/storage"
	"github.com/wasp-project/yazi/pkg/storage/memory"

	"github.com/mlycore/log"
)

type Server struct {
	conf  *config.ServerConfig
	store storage.KVStore

	core   *TCPServer
	connCh chan net.Conn
	errCh  chan error
	// codec protocol.Codec
}

func NewServer(conf *config.ServerConfig) *Server {
	return &Server{
		conf:   conf,
		connCh: make(chan net.Conn),
		errCh:  make(chan error),
	}
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

	s.core = &TCPServer{
		connCh: s.connCh,
		errCh:  s.errCh,
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatalf("Server cannot listen and serve...")
	}

	log.Infoln("Server is running...")
}

func (s *Server) ListenAndServe() error {
	s.core.Open(net.JoinHostPort("127.0.0.1", strconv.Itoa(s.conf.Port)))
	defer s.core.Close()

	go s.core.Listen()

	var err error

	for {
		select {
		case conn := <-s.connCh:
			go s.handle(conn)
		case err = <-s.errCh:
			log.Errorf("Listening error: %s", err)
			return err
		// FIXME: remove after the server is running good
		default:
			time.Sleep(time.Second)
			log.Infof("Waiting for connection...")
		}
	}
}

func (s *Server) handle(conn net.Conn) error {
	defer conn.Close()
	conn.Write([]byte("hello"))
	return nil
}

func (s *Server) execute() error {
	return nil
}
