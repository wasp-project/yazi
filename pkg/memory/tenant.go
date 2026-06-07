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

package memory

import "strings"

var _ KV = (*prefixKV)(nil)

// prefixKV decorates a KV so every key is transparently namespaced under a
// tenant prefix. Writes and point reads prepend the prefix; Keys lists only the
// keys belonging to this tenant and strips the prefix on the way out, so the
// memory Store continues to see plain /_memory/... key shapes and its index
// scans (which match on those shapes) keep working unchanged.
type prefixKV struct {
	inner  KV
	prefix string
}

func newPrefixKV(inner KV, prefix string) *prefixKV {
	return &prefixKV{inner: inner, prefix: prefix}
}

func (p *prefixKV) Get(key string) (string, error) {
	return p.inner.Get(p.prefix + key)
}

func (p *prefixKV) Set(key, value string) error {
	return p.inner.Set(p.prefix+key, value)
}

func (p *prefixKV) Del(key string) error {
	return p.inner.Del(p.prefix + key)
}

func (p *prefixKV) MGet(keys []string) ([]string, error) {
	return p.inner.MGet(p.addPrefix(keys))
}

func (p *prefixKV) MSet(keys, values []string) error {
	return p.inner.MSet(p.addPrefix(keys), values)
}

// Keys returns only the keys under this tenant's prefix, with the prefix
// removed, so callers observe the same key space a single-tenant store would.
func (p *prefixKV) Keys() ([]string, error) {
	all, err := p.inner.Keys()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(all))
	for _, k := range all {
		if strings.HasPrefix(k, p.prefix) {
			out = append(out, strings.TrimPrefix(k, p.prefix))
		}
	}
	return out, nil
}

func (p *prefixKV) addPrefix(keys []string) []string {
	out := make([]string, len(keys))
	for i, k := range keys {
		out[i] = p.prefix + k
	}
	return out
}
