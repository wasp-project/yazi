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

// Package tenant is the tenant-service layer of the memory system.
//
// It is a thin seam between the user-interface layer and the storage-engine
// layer. Every memory operation resolves a TenantContext through a Service and
// uses its KeyPrefix to namespace keys. The default Service is a no-op bypass
// (see bypass.go) that preserves single-tenant behavior; a separate project can
// supply a real multi-tenant implementation by satisfying the Service interface
// without touching the memory or storage layers.
package tenant

import "strings"

// Request carries the inputs a Service uses to resolve a tenant. It is kept
// deliberately storage-agnostic: it never references storage-engine types, so
// external implementations stay decoupled from internal persistence details.
type Request struct {
	// TenantID is an explicit tenant identifier supplied by the caller
	// (e.g. a CLI flag or an RPC header). Empty means "no tenant".
	TenantID string
	// Metadata holds optional auxiliary identity inputs (headers, claims, ...)
	// for richer external implementations. The bypass service ignores it.
	Metadata map[string]string
}

// TenantContext is the resolved tenant for a single operation. The memory layer
// asks it for a KeyPrefix and prepends that prefix to every key it constructs.
type TenantContext struct {
	id     string
	prefix string
}

// ID returns the resolved tenant identifier, or "" in bypass mode.
func (c TenantContext) ID() string { return c.id }

// KeyPrefix returns the prefix to prepend to every memory/index key for this
// tenant. It is "" in bypass mode, which reproduces the original un-prefixed
// /_memory/... layout.
func (c TenantContext) KeyPrefix() string { return c.prefix }

// IsBypass reports whether this context applies no tenant namespacing.
func (c TenantContext) IsBypass() bool { return c.prefix == "" }

// Service resolves a Request into a TenantContext. Implementations MUST be
// deterministic: the same tenant identifier MUST yield the same key prefix.
type Service interface {
	Resolve(req Request) (TenantContext, error)
}

// prefixFor derives the deterministic, sanitized key prefix for a tenant id.
// An empty id yields an empty prefix (bypass). A non-empty id yields
// "tenant/<sanitized-id>/", which is unique and stable per id.
func prefixFor(id string) string {
	id = strings.TrimSpace(id)
	if id == "" {
		return ""
	}
	return "tenant/" + sanitize(id) + "/"
}

// sanitize maps a tenant id to a key-safe token. It keeps alphanumerics, '-',
// '_' and '.' as-is and replaces any other rune with '_', so distinct ids that
// differ only in unsafe characters remain distinguishable by their safe parts
// while never producing keys that collide with the reserved '/' separator.
func sanitize(id string) string {
	var b strings.Builder
	b.Grow(len(id))
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

// NewContext builds a TenantContext for the given id using the standard prefix
// derivation. Exposed so external Service implementations can reuse the same
// deterministic namespacing instead of inventing their own.
func NewContext(id string) TenantContext {
	return TenantContext{id: strings.TrimSpace(id), prefix: prefixFor(id)}
}
