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

import (
	"strings"
	"testing"

	"github.com/wasp-project/yazi/pkg/policy"
	"github.com/wasp-project/yazi/pkg/storage"
)

// TestEmptyPrefixPreservesLayout verifies that NewStoreWithTenant(kv, "")
// stores keys under the original /_memory/... layout with no tenant prefix.
func TestEmptyPrefixPreservesLayout(t *testing.T) {
	kv := storage.NewKVStore(1024, policy.KeyPolicy(""))
	store := NewStoreWithTenant(kv, "")

	id, err := store.PutBasic(BasicMemory{Kind: BasicKindPreference, Scope: "user", Content: "x"})
	if err != nil {
		t.Fatalf("put: %v", err)
	}

	keys, err := kv.Keys()
	if err != nil {
		t.Fatalf("keys: %v", err)
	}
	wantKey := basicPrefix + id
	found := false
	for _, k := range keys {
		if k == wantKey {
			found = true
		}
		if strings.HasPrefix(k, "tenant/") {
			t.Fatalf("unexpected tenant-prefixed key with empty prefix: %q", k)
		}
	}
	if !found {
		t.Fatalf("expected un-prefixed key %q in %v", wantKey, keys)
	}
}

// TestTenantPrefixIsolation verifies two tenants on the same KV cannot see each
// other's records, and that the underlying keys are namespaced per tenant.
func TestTenantPrefixIsolation(t *testing.T) {
	kv := storage.NewKVStore(1024, policy.KeyPolicy(""))
	acme := NewStoreWithTenant(kv, "tenant/acme/")
	globex := NewStoreWithTenant(kv, "tenant/globex/")

	acmeID, err := acme.PutBasic(BasicMemory{Kind: BasicKindPreference, Scope: "user", Subject: "ui", Content: "dark", Tags: []string{"theme"}})
	if err != nil {
		t.Fatalf("acme put: %v", err)
	}
	if _, err := globex.PutBasic(BasicMemory{Kind: BasicKindPreference, Scope: "user", Subject: "ui", Content: "light", Tags: []string{"theme"}}); err != nil {
		t.Fatalf("globex put: %v", err)
	}

	// globex must not be able to read acme's record by id.
	if _, err := globex.GetBasic(acmeID); err == nil {
		t.Fatalf("globex should not read acme record %s", acmeID)
	}

	// Each tenant lists exactly one record, its own.
	acmeList, err := acme.ListBasic(BasicFilter{Tag: "theme"})
	if err != nil {
		t.Fatalf("acme list: %v", err)
	}
	if len(acmeList) != 1 || acmeList[0].Content != "dark" {
		t.Fatalf("acme list wrong: %+v", acmeList)
	}
	globexList, err := globex.ListBasic(BasicFilter{Tag: "theme"})
	if err != nil {
		t.Fatalf("globex list: %v", err)
	}
	if len(globexList) != 1 || globexList[0].Content != "light" {
		t.Fatalf("globex list wrong: %+v", globexList)
	}

	// Underlying KV keys must carry distinct tenant prefixes.
	keys, _ := kv.Keys()
	var sawAcme, sawGlobex bool
	for _, k := range keys {
		if strings.HasPrefix(k, "tenant/acme/"+basicPrefix) {
			sawAcme = true
		}
		if strings.HasPrefix(k, "tenant/globex/"+basicPrefix) {
			sawGlobex = true
		}
	}
	if !sawAcme || !sawGlobex {
		t.Fatalf("expected both tenant-namespaced keys, got %v", keys)
	}
}

// TestTenantDefaultScanIsolation verifies the no-filter list path (which scans
// all keys) also respects tenant isolation.
func TestTenantDefaultScanIsolation(t *testing.T) {
	kv := storage.NewKVStore(1024, policy.KeyPolicy(""))
	acme := NewStoreWithTenant(kv, "tenant/acme/")
	globex := NewStoreWithTenant(kv, "tenant/globex/")

	if _, err := acme.PutBasic(BasicMemory{Content: "a"}); err != nil {
		t.Fatalf("acme put: %v", err)
	}
	if _, err := globex.PutBasic(BasicMemory{Content: "b"}); err != nil {
		t.Fatalf("globex put: %v", err)
	}

	list, err := acme.ListBasic(BasicFilter{})
	if err != nil {
		t.Fatalf("acme list: %v", err)
	}
	if len(list) != 1 || list[0].Content != "a" {
		t.Fatalf("acme default scan leaked: %+v", list)
	}
}
