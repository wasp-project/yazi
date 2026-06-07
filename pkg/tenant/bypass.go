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

package tenant

var (
	_ Service = (*BypassService)(nil)
	_ Service = (*StaticService)(nil)
)

// BypassService is the default no-op tenant service. It resolves every request
// to an empty TenantContext (no id, empty key prefix), so memory keys keep the
// original /_memory/... layout and behavior is identical to the pre-tenant
// system. It is the fallback whenever no tenant is configured.
type BypassService struct{}

// NewBypassService returns the default bypass service.
func NewBypassService() *BypassService { return &BypassService{} }

// Resolve always returns the empty (bypass) context, ignoring the request.
func (s *BypassService) Resolve(req Request) (TenantContext, error) {
	return TenantContext{}, nil
}

// StaticService pins all operations to a single configured tenant id. It is the
// simplest non-bypass implementation: useful for a single-tenant-but-namespaced
// deployment or for routing a CLI/process to one tenant. A request's TenantID,
// when set, overrides the configured default so the same process can act for
// different tenants per call.
type StaticService struct {
	defaultTenant string
}

// NewStaticService returns a Service that resolves to the given tenant id. If
// the id is empty it behaves like the bypass service.
func NewStaticService(tenantID string) *StaticService {
	return &StaticService{defaultTenant: tenantID}
}

// New builds a Service from configuration. Mode "" or "bypass" returns the
// default bypass service; mode "static" returns a StaticService pinned to id.
// Unknown modes fall back to bypass so misconfiguration never silently isolates
// or leaks data unexpectedly.
func New(mode, id string) Service {
	switch mode {
	case "static":
		return NewStaticService(id)
	default:
		return NewBypassService()
	}
}

// Resolve uses req.TenantID when present, otherwise the configured default,
// and derives a deterministic prefix from it.
func (s *StaticService) Resolve(req Request) (TenantContext, error) {
	id := req.TenantID
	if id == "" {
		id = s.defaultTenant
	}
	return NewContext(id), nil
}
