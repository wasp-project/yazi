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

import "github.com/wasp-project/yazi/pkg/storage"

type Service interface {
	Run()
	SetStorage(s storage.KVStore)

	GetMeta(key string) (string, error)
	SetMeta(key, value string) error
	DelMeta(key string) error

	GetRaft(key string) (string, error)
	SetRaft(key, value string) error
	DelRaft(key string) error
}
