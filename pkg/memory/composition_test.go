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
	"testing"

	"github.com/wasp-project/yazi/pkg/memory/provider"
	"github.com/wasp-project/yazi/pkg/policy"
	"github.com/wasp-project/yazi/pkg/storage"
)

// The deterministic pipeline, when backed by a real Store, persists through the
// KV/storage path and retrieves with zero token cost.
func TestStorePipelinePersistsAndRecalls(t *testing.T) {
	kv := storage.NewKVStore(1024, policy.KeyPolicy(""))
	store := NewStore(kv)
	meter := provider.NewMeter()
	p := NewStorePipeline(store, provider.Budget{}, provider.DefaultPricing(), meter, "acme")

	if _, err := p.Ingest(provider.Item{ID: "m1", Class: "basic", Kind: "preference", Scope: "user", Subject: "ui", Text: "prefers dark mode in the UI", Tags: []string{"ui"}}); err != nil {
		t.Fatalf("ingest: %v", err)
	}

	// The record is persisted in the underlying typed Store (durable path).
	got, err := store.GetBasic("m1")
	if err != nil {
		t.Fatalf("record not persisted to Store: %v", err)
	}
	if got.Content != "prefers dark mode in the UI" {
		t.Fatalf("unexpected persisted content: %q", got.Content)
	}

	hits, u, err := p.Recall(provider.Query{Text: "what UI theme", Tag: "ui", TopK: 3})
	if err != nil {
		t.Fatalf("recall: %v", err)
	}
	if len(hits) == 0 || hits[0].ID != "m1" {
		t.Fatalf("expected m1, got %+v", hits)
	}
	if u.EmbedTokens != 0 || u.LLMInputTokens != 0 {
		t.Fatalf("student/store pipeline must be zero-token, got %+v", u)
	}
}
