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

package client

import (
	"bufio"
	"errors"
	"fmt"
	"net"

	"github.com/wasp-project/yazi/pkg/protocol"
)

func NewYaziClient(p protocol.Protocol) (Client, error) {
	switch p {
	case protocol.ProtocolNaive:
		return &naiveclient{}, nil
	case protocol.ProtocolGrpc:
		return &grpcclient{}, nil
	default:
		return nil, errors.New("unsupported protocol")
	}

}

type Client interface {
	Connect(host, port string) error
	Get(key string) (string, error)
	Set(key, value string) error
	Close() error
}

type naiveclient struct {
	conn net.Conn
}

func (c *naiveclient) Connect(host, port string) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *naiveclient) Close() error {
	c.conn.Close()
	return nil
}

func (c *naiveclient) Get(key string) (string, error) {
	req := fmt.Sprintf("+get %s;", key)
	_, err := c.conn.Write([]byte(req))
	if err != nil {
		return "", err
	}

	reader := bufio.NewReader(c.conn)
	resp, _, err := reader.ReadLine()
	if err != nil {
		return "", err
	}
	return string(resp), nil
}

func (c *naiveclient) Set(key, value string) error {
	req := fmt.Sprintf("+set %s %s;", key, value)
	_, err := c.conn.Write([]byte(req))
	if err != nil {
		return err
	}

	reader := bufio.NewReader(c.conn)
	_, _, err = reader.ReadLine()
	if err != nil {
		return err
	}
	return nil
}

type grpcclient struct{}

func (c *grpcclient) Connect(host, port string) error {
	return nil
}

func (c *grpcclient) Close() error {
	return nil
}

func (c *grpcclient) Get(key string) (string, error) {
	return "", nil
}

func (c *grpcclient) Set(key, value string) error {
	return nil
}
