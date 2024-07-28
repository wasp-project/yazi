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
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/wasp-project/yazi/pkg/protocol"
	"github.com/wasp-project/yazi/pkg/server/pb"
	"github.com/wasp-project/yazi/pkg/utils"
	"google.golang.org/grpc"
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
	Del(key string) error
	MGet(keys []string) ([]string, error)
	MSet(keys, values []string) error
	Keys() ([]string, error)
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

func (c *naiveclient) Del(key string) error {
	utils.TODO()
	return nil
}
func (c *naiveclient) MSet(keys, values []string) error {
	utils.TODO()
	return nil
}
func (c *naiveclient) MGet(keys []string) ([]string, error) {
	utils.TODO()
	return []string{}, nil
}
func (c *naiveclient) Keys() ([]string, error) {
	utils.TODO()
	return []string{}, nil
}

type grpcclient struct {
	conn *grpc.ClientConn
}

func (c *grpcclient) Connect(host, port string) error {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", host, port), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}

func (c *grpcclient) Close() error {
	c.conn.Close()
	return nil
}

func (c *grpcclient) Get(key string) (string, error) {
	cli := pb.NewRPCServerClient(c.conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := cli.Get(ctx, &pb.KVRequest{Key: key})
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}

func (c *grpcclient) Set(key, value string) error {
	cli := pb.NewRPCServerClient(c.conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := cli.Set(ctx, &pb.KVRequest{Key: key, Value: value})
	return err
}

func (c *grpcclient) Del(key string) error {
	cli := pb.NewRPCServerClient(c.conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := cli.Del(ctx, &pb.KVRequest{Key: key})
	return err
}

func (c *grpcclient) MGet(keys []string) ([]string, error) {
	cli := pb.NewRPCServerClient(c.conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := cli.MGet(ctx, &pb.MKVRequest{Keys: keys})
	if err != nil {
		return []string{}, err
	}
	return resp.Values, nil
}

func (c *grpcclient) MSet(keys, values []string) error {
	cli := pb.NewRPCServerClient(c.conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err := cli.MSet(ctx, &pb.MKVRequest{Keys: keys, Values: values})
	return err
}

func (c *grpcclient) Keys() ([]string, error) {
	cli := pb.NewRPCServerClient(c.conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := cli.Keys(ctx, &pb.MKVRequest{})
	if err != nil {
		return []string{}, err
	}
	return resp.Values, nil
}
