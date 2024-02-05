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

package naive

import (
	"reflect"
	"strings"

	"github.com/mlycore/log"
)

type Codec interface {
	Decode(data []byte) *Command
	Encode(cmd *Command) []byte
}

type NaiveCodec struct{}

var _ Codec = &NaiveCodec{}

// +set k v
// +get k
func (c *NaiveCodec) Decode(data []byte) *Command {
	cmd := &Command{}
	if data[0] == byte(CommandKindString) {
		cmd.Kind = CommandKindString
	}

	if reflect.DeepEqual(data[1:4], []byte(CommandVerbGet)) {
		cmd.Verb = CommandVerbGet
		ops := data[5 : len(data)-1]
		key := strings.Split(string(ops), " ")[0]
		cmd.Request.Data = KVPair{
			Key: key,
		}
	}

	if reflect.DeepEqual(data[1:4], []byte(CommandVerbSet)) {
		cmd.Verb = CommandVerbSet
		ops := data[5 : len(data)-1]
		key := strings.Split(string(ops), " ")[0]
		val := strings.Split(string(ops), " ")[1]
		cmd.Request.Data = KVPair{
			Key:   key,
			Value: val,
		}
	}
	log.Debugf("cmd.kind: %s, verb: %s, data: %#v", cmd.Kind, cmd.Verb, cmd.Request.Data)
	return cmd
}

func (c *NaiveCodec) Encode(cmd *Command) []byte {
	data := []byte{}

	if cmd.Response.Error != nil {
		data = append(data, []byte(cmd.Response.Error.Error())...)
		data = append(data, []byte("\n")...)
		return data
	}

	switch cmd.Verb {
	case CommandVerbGet:
		{
			kv := cmd.Response.Data.(KVPair)
			data = append(data, []byte(kv.Value)...)
		}
	case CommandVerbSet:
		{
			data = append(data, []byte("OK")...)
		}
	}
	data = append(data, []byte("\n")...)
	return data
}
