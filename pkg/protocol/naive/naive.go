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

type CommandKind byte

type Command struct {
	Kind     CommandKind
	Verb     CommandVerb
	Request  Request
	Response Response
}

type Request struct {
	Data any
}

type Response struct {
	Data  any
	Error error
}

type KVPair struct {
	Key   string
	Value string
}

func (c *Command) GetKind() CommandKind {
	return c.Kind
}

type CommandVerb string

const (
	CommandVerbGet CommandVerb = "get"
	CommandVerbSet CommandVerb = "set"
)

func (c *Command) GetVerb() CommandVerb {
	return c.Verb
}

func (c *Command) SetRequest(data any) {
	c.Request = Request{
		Data: data,
	}
}

func (c *Command) GetRequest() Request {
	return c.Request
}

func (c *Command) SetResponse(data any, err error) {
	c.Response = Response{
		Data:  data,
		Error: err,
	}
}

func (c *Command) GetResponse() Response {
	return c.Response
}

const (
	CommandKindString CommandKind = '+'
)
