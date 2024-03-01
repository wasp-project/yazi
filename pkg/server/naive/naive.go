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

package naive

import (
	"bufio"
	"net"
	"strconv"
	"time"

	cn "github.com/wasp-project/yazi/pkg/protocol/naive"
	"github.com/wasp-project/yazi/pkg/storage"
	"github.com/wasp-project/yazi/pkg/utils"

	"github.com/mlycore/log"
)

func NewService(codec cn.Codec, port int) *naive {
	return &naive{
		codec: codec,
		core: &TCPServer{
			connCh: make(chan net.Conn),
			errCh:  make(chan error),
			port:   port,
		},
	}
}

type naive struct {
	core  *TCPServer
	codec cn.Codec
	store storage.KVStore
}

func (s *naive) Run() {
	if err := s.listenAndServe(); err != nil {
		log.Fatalf("Server cannot listen and serve...")
	}
}

func (s *naive) SetStorage(stor storage.KVStore) {
	s.store = stor
}

func (s *naive) listenAndServe() error {
	s.core.Open(net.JoinHostPort("127.0.0.1", strconv.Itoa(s.core.port)))
	defer s.core.Close()

	go s.core.Listen()

	var err error

	for {
		select {
		case conn := <-s.core.connCh:
			go s.handle(conn)
		case err = <-s.core.errCh:
			log.Errorf("Listening error: %s", err)
			return err
		// FIXME: remove after the server is running good
		default:
			time.Sleep(time.Second)
			log.Infof("Waiting for connection...")
		}
	}
}

func (s *naive) handle(conn net.Conn) error {
	defer conn.Close()

	for {
		reader := bufio.NewReader(conn)
		data, err := reader.ReadBytes(byte(';'))
		if err != nil {
			return err
		}
		log.Debugf("received data: %s", data)

		switch cmd := s.codec.Decode(data); cmd.GetKind() {
		case cn.CommandKindString:
			{
				switch cmd.GetVerb() {
				case cn.CommandVerbSet:
					{
						kv := cmd.Request.Data.(cn.KVPair)
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
				case cn.CommandVerbGet:
					{
						var err error
						kv := cmd.Request.Data.(cn.KVPair)
						if kv.Value, err = s.store.Get(kv.Key); err != nil {
							log.Errorf("Error getting key %s: %s", kv.Key, err)
							cmd.SetResponse(nil, err)
							resp := s.codec.Encode(cmd)
							conn.Write(resp)
						} else {
							cmd.SetResponse(cn.KVPair{
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

func (s *naive) GetMeta(key string) (string, error) {
	utils.TODO()
	return "", nil
}

func (s *naive) SetMeta(key, value string) error {
	utils.TODO()
	return nil
}

func (s *naive) DelMeta(key string) error {
	utils.TODO()
	return nil
}

func (s *naive) GetRaft(key string) (string, error) {
	utils.TODO()
	return "", nil
}

func (s *naive) SetRaft(key, value string) error {
	utils.TODO()
	return nil
}

func (s *naive) DelRaft(key string) error {
	utils.TODO()
	return nil
}
