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

import "testing"

func TestBypassReturnsEmptyPrefix(t *testing.T) {
	s := NewBypassService()
	ctx, err := s.Resolve(Request{TenantID: "acme"})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !ctx.IsBypass() {
		t.Fatalf("expected bypass context")
	}
	if ctx.KeyPrefix() != "" {
		t.Fatalf("expected empty prefix, got %q", ctx.KeyPrefix())
	}
	if ctx.ID() != "" {
		t.Fatalf("expected empty id, got %q", ctx.ID())
	}
}

func TestStaticPrefixIsStable(t *testing.T) {
	s := NewStaticService("acme")
	a, err := s.Resolve(Request{})
	if err != nil {
		t.Fatalf("resolve a: %v", err)
	}
	b, err := s.Resolve(Request{})
	if err != nil {
		t.Fatalf("resolve b: %v", err)
	}
	if a.KeyPrefix() == "" {
		t.Fatalf("expected non-empty prefix for named tenant")
	}
	if a.KeyPrefix() != b.KeyPrefix() {
		t.Fatalf("prefix not stable: %q vs %q", a.KeyPrefix(), b.KeyPrefix())
	}
	if a.KeyPrefix() != "tenant/acme/" {
		t.Fatalf("unexpected prefix: %q", a.KeyPrefix())
	}
}

func TestDistinctTenantsGetDistinctPrefixes(t *testing.T) {
	acme := NewContext("acme")
	globex := NewContext("globex")
	if acme.KeyPrefix() == globex.KeyPrefix() {
		t.Fatalf("expected distinct prefixes, both %q", acme.KeyPrefix())
	}
}

func TestRequestTenantOverridesDefault(t *testing.T) {
	s := NewStaticService("acme")
	ctx, err := s.Resolve(Request{TenantID: "globex"})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if ctx.KeyPrefix() != "tenant/globex/" {
		t.Fatalf("expected override to globex, got %q", ctx.KeyPrefix())
	}
}

func TestSanitizeUnsafeCharacters(t *testing.T) {
	ctx := NewContext("a/b c")
	// '/' and ' ' must not survive as-is; the prefix stays well-formed.
	if got := ctx.KeyPrefix(); got != "tenant/a_b_c/" {
		t.Fatalf("unexpected sanitized prefix: %q", got)
	}
}

func TestEmptyIDIsBypass(t *testing.T) {
	if got := NewContext("   ").KeyPrefix(); got != "" {
		t.Fatalf("expected empty prefix for blank id, got %q", got)
	}
}

func TestNewDefaultsToBypass(t *testing.T) {
	// No tenant block configured (zero-value mode/id) must yield bypass.
	ctx, err := New("", "").Resolve(Request{})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if !ctx.IsBypass() {
		t.Fatalf("expected bypass for empty config")
	}
}

func TestNewStaticMode(t *testing.T) {
	ctx, err := New("static", "acme").Resolve(Request{})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if ctx.KeyPrefix() != "tenant/acme/" {
		t.Fatalf("unexpected prefix: %q", ctx.KeyPrefix())
	}
}
