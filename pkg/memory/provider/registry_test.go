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

package provider

import (
	"strings"
	"testing"
)

func TestBuildKnownProfiles(t *testing.T) {
	for _, name := range []string{"lite", "standard", ""} {
		if _, err := BuildPipeline(Config{Profile: name}, nil, ""); err != nil {
			t.Fatalf("profile %q should build: %v", name, err)
		}
	}
}

func TestProProfileFailsFastOnUnavailableProvider(t *testing.T) {
	_, err := BuildPipeline(Config{Profile: "pro"}, nil, "")
	if err == nil {
		t.Fatalf("pro profile should fail fast (llm providers unavailable)")
	}
	if !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("expected an unavailable-provider error, got: %v", err)
	}
}

func TestUnknownProfileErrors(t *testing.T) {
	if _, err := BuildPipeline(Config{Profile: "platinum"}, nil, ""); err == nil {
		t.Fatalf("unknown profile should error")
	}
}

func TestCustomOverridesSelectedStagesOnly(t *testing.T) {
	// custom: local embedder + vector store, deterministic extractor/reranker.
	p, err := BuildPipeline(Config{Profile: "custom", Embedder: "local", Store: "vector"}, nil, "")
	if err != nil {
		t.Fatalf("custom build: %v", err)
	}
	if p.stages.Embedder == nil {
		t.Fatalf("custom embedder should be set")
	}
	if p.stages.Extractor != nil {
		t.Fatalf("unspecified custom extractor should default to none (nil)")
	}
}

func TestCustomUnavailableStoreFailsFast(t *testing.T) {
	_, err := BuildPipeline(Config{Profile: "custom", Store: "pgvector"}, nil, "")
	if err == nil || !strings.Contains(err.Error(), "Postgres") {
		t.Fatalf("expected pgvector requirement error, got: %v", err)
	}
}

func TestRepricingFromRecordedUsage(t *testing.T) {
	m := NewMeter()
	m.Record("acme", "basic", Usage{EmbedTokens: 1_000_000})
	cheap := m.CostFor("acme", "basic", Pricing{EmbedPerM: 0.02})
	pricey := m.CostFor("acme", "basic", Pricing{EmbedPerM: 0.20})
	if !(pricey > cheap && cheap > 0) {
		t.Fatalf("re-pricing should scale cost: cheap=%v pricey=%v", cheap, pricey)
	}
}
