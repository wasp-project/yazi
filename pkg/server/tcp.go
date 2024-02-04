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
)

type TCPServer struct {
	Listener net.Listener
	connCh   chan net.Conn
	errCh    chan error
}

func (s *TCPServer) Open(addr string) (net.Conn, error) {
	var err error
	if s.Listener, err = net.Listen("tcp", addr); err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *TCPServer) Listen() {
	for {
		if s.Listener == nil {
			break
		}

		if conn, err := s.Listener.Accept(); err != nil {
			s.errCh <- err
			break
		} else {
			s.connCh <- conn
		}
	}
}

func (s *TCPServer) Close() error {
	return s.Listener.Close()
}
