// Copyright 2024 mlycore. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpc

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/wasp-project/yazi/pkg/server/pb"
	"github.com/wasp-project/yazi/pkg/storage"
	"github.com/wasp-project/yazi/pkg/utils"

	"github.com/mlycore/log"
	"google.golang.org/grpc"
)

func NewService(port int) *gr {
	return &gr{
		port: port,
	}
}

type gr struct {
	port  int
	store storage.KVStore
	pb.UnimplementedRPCServerServer
}

func (s *gr) Run() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	gs := grpc.NewServer()
	pb.RegisterRPCServerServer(gs, s)
	if err := gs.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *gr) SetStorage(storage storage.KVStore) {
	s.store = storage
}

const (
	MetadataKeyPrefix = "/_meta"
	RaftKeyPrefix     = "/_raft"
	DataKeyPrefix     = "/_data/"
)

func (s *gr) Set(ctx context.Context, req *pb.KVRequest) (resp *pb.KVResponse, err error) {
	if err = s.store.Set(DataKeyPrefix+req.Key, req.Value); err != nil {
		log.Errorf("Error setting key %s: %s", req.Key, err)
	}

	return
}

func (s *gr) Get(ctx context.Context, req *pb.KVRequest) (*pb.KVResponse, error) {
	resp := &pb.KVResponse{}
	var err error

	if resp.Value, err = s.store.Get(DataKeyPrefix + req.Key); err != nil {
		log.Errorf("Error getting key %s: %s", req.Key, err)
	}

	return resp, err
}

func (s *gr) Del(ctx context.Context, req *pb.KVRequest) (*pb.KVResponse, error) {
	resp := &pb.KVResponse{}
	var err error

	if err = s.store.Del(DataKeyPrefix + req.Key); err != nil {
		log.Errorf("Error deleting key %s: %s", req.Key, err)
	}
	return resp, err
}

func (s *gr) MSet(ctx context.Context, req *pb.MKVRequest) (*pb.MKVResponse, error) {
	var (
		err   error
		nkeys []string        = make([]string, len(req.Keys))
		resp  *pb.MKVResponse = &pb.MKVResponse{}
	)
	for i := 0; i < len(req.Keys); i++ {
		nkeys[i] = DataKeyPrefix + req.Keys[i]
	}
	if err = s.store.MSet(nkeys, req.Values); err != nil {
		log.Errorf("Error mset %d keys: %s", len(req.Keys), err)
	}
	return resp, err
}

func (s *gr) MGet(ctx context.Context, req *pb.MKVRequest) (*pb.MKVResponse, error) {
	var (
		err   error
		nkeys []string        = make([]string, len(req.Keys))
		resp  *pb.MKVResponse = &pb.MKVResponse{}
	)
	for i := 0; i < len(req.Keys); i++ {
		nkeys[i] = DataKeyPrefix + req.Keys[i]
	}
	if resp.Values, err = s.store.MGet(nkeys); err != nil {
		log.Errorf("Error mget %d keys: %s", len(req.Keys), err)
	}
	return resp, err
}

func (s *gr) Keys(ctx context.Context, req *pb.MKVRequest) (*pb.MKVResponse, error) {
	var (
		err  error
		keys []string
		resp *pb.MKVResponse = &pb.MKVResponse{}
	)

	if keys, err = s.store.Keys(); err != nil {
		log.Errorf("Error keys: %s", err)
	}
	resp.Values = make([]string, len(keys))
	for i := 0; i < len(keys); i++ {
		resp.Values[i] = strings.TrimPrefix(keys[i], DataKeyPrefix)
	}

	return resp, err
}

func (s *gr) GetMeta(key string) (string, error) {
	utils.TODO()
	return "", nil
}

func (s *gr) SetMeta(key, value string) error {
	return s.store.Set(MetadataKeyPrefix+key, value)
}

func (s *gr) DelMeta(key string) error {
	utils.TODO()
	return nil
}

func (s *gr) GetRaft(key string) (string, error) {
	utils.TODO()
	return "", nil
}

func (s *gr) SetRaft(key, value string) error {
	return s.store.Set(RaftKeyPrefix+key, value)
}

func (s *gr) DelRaft(key string) error {
	utils.TODO()
	return nil
}
