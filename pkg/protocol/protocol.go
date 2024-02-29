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

package protocol

import "github.com/wasp-project/yazi/pkg/protocol/naive"

type Protocol string

const (
	ProtocolNaive Protocol = "naive"
	ProtocolGrpc  Protocol = "grpc"
)

type EmptyCodec struct{}

func (c *EmptyCodec) Decode(data []byte) *naive.Command { return nil }
func (c *EmptyCodec) Encode(cmd *naive.Command) []byte  { return nil }
