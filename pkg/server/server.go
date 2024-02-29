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
	"bufio"
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/wasp-project/yazi/pkg/config"
	"github.com/wasp-project/yazi/pkg/protocol"
	"github.com/wasp-project/yazi/pkg/protocol/naive"
	"github.com/wasp-project/yazi/pkg/server/pb"
	"github.com/wasp-project/yazi/pkg/storage"
	"github.com/wasp-project/yazi/pkg/storage/local"
	"google.golang.org/grpc"

	"github.com/mlycore/log"
)

type Server struct {
	conf    *config.ServerConfig
	manager *storage.Manager
	store   storage.KVStore

	core   *TCPServer
	connCh chan net.Conn
	errCh  chan error
	codec  naive.Codec

	pb.UnimplementedRPCServerServer
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
	log.Infof("Server is configured with protocol: %s", s.conf.Protocol)

	// init storage
	store := storage.NewKVStore(s.conf.Capacity, s.conf.Policy)
	switch s.conf.Storage {
	case storage.StorageClassLocal:
		w := local.NewWriter()
		s.manager = storage.NewManager().SetPersistenter(w).SetStore(store)
		s.manager.SetTask("persistent", s.manager.Persistent)
	default:
		s.manager = storage.NewManager().SetStore(store)
	}

	s.store = store
	s.manager.SetTask("memory-check", storage.TaskMemoryCheck)

	go s.manager.Run()

	// init protocol
	switch s.conf.Protocol {
	case protocol.ProtocolNaive:
		// grpc does not need an independent codec
		s.codec = &naive.NaiveCodec{}
		s.startNaiveServer()
	case protocol.ProtocolGrpc:
		fallthrough
	default:
		s.codec = &protocol.EmptyCodec{}
		s.startGrpcServer()
	}
}

func (s *Server) startNaiveServer() {
	// init tcp server
	s.core = &TCPServer{
		connCh: s.connCh,
		errCh:  s.errCh,
	}

	// start server
	if err := s.listenAndServe(); err != nil {
		log.Fatalf("Server cannot listen and serve...")
	}
}

func (s *Server) listenAndServe() error {
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

	for {
		reader := bufio.NewReader(conn)
		data, err := reader.ReadBytes(byte(';'))
		if err != nil {
			return err
		}
		log.Debugf("received data: %s", data)

		switch cmd := s.codec.Decode(data); cmd.GetKind() {
		case naive.CommandKindString:
			{
				switch cmd.GetVerb() {
				case naive.CommandVerbSet:
					{
						kv := cmd.Request.Data.(naive.KVPair)
						if err := s.store.Set(kv.Key, kv.Value); err != nil {
							log.Errorf("Error setting key %s: %s", kv.Key, err)
							cmd.SetResponse(nil, err)
							resp := s.codec.Encode(cmd)
							conn.Write(resp)
						} else {
							cmd.SetResponse(nil, nil)
							resp := s.codec.Encode(cmd)
							conn.Write(resp)
						}
					}
				case naive.CommandVerbGet:
					{
						var err error
						kv := cmd.Request.Data.(naive.KVPair)
						if kv.Value, err = s.store.Get(kv.Key); err != nil {
							log.Errorf("Error getting key %s: %s", kv.Key, err)
							cmd.SetResponse(nil, err)
							resp := s.codec.Encode(cmd)
							conn.Write(resp)
						} else {
							cmd.SetResponse(naive.KVPair{
								Key:   kv.Key,
								Value: kv.Value,
							}, nil)
							resp := s.codec.Encode(cmd)
							conn.Write(resp)
						}
					}
				}
			}
		default:
			log.Warnln("Unrecognized command kind")
		}

	}

	return nil
}

func (s *Server) Set(ctx context.Context, req *pb.KVRequest) (resp *pb.KVResponse, err error) {
	if err = s.store.Set(req.Key, req.Value); err != nil {
		log.Errorf("Error setting key %s: %s", req.Key, err)
	}

	return
}

func (s *Server) Get(ctx context.Context, req *pb.KVRequest) (*pb.KVResponse, error) {
	resp := &pb.KVResponse{}
	var err error

	if resp.Value, err = s.store.Get(req.Key); err != nil {
		log.Errorf("Error getting key %s: %s", req.Key, err)
	}

	return resp, err
}

func (s *Server) startGrpcServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.conf.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	gs := grpc.NewServer()
	pb.RegisterRPCServerServer(gs, s)
	if err := gs.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
