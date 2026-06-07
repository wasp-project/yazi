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

import "testing"

func liteItems() []Item {
	return []Item{
		{ID: "p1", Text: "The user prefers concise answer-first responses", Tags: []string{"style"}},
		{ID: "p2", Text: "The user prefers dark mode in the UI", Tags: []string{"ui"}},
		{ID: "p3", Text: "The user wants code examples in Go", Tags: []string{"code"}},
	}
}

func ingestAll(t *testing.T, p *Pipeline, items []Item) {
	t.Helper()
	for _, it := range items {
		if _, err := p.Ingest(it); err != nil {
			t.Fatalf("ingest %s: %v", it.ID, err)
		}
	}
}

func TestStudentPipelineZeroCostAndCorrect(t *testing.T) {
	p, err := BuildPipeline(Config{Profile: "lite"}, nil, "")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	ingestAll(t, p, liteItems())

	hits, u, err := p.Recall(Query{Text: "what UI theme does the user want", Tag: "ui", TopK: 3})
	if err != nil {
		t.Fatalf("recall: %v", err)
	}
	if len(hits) == 0 || hits[0].ID != "p2" {
		t.Fatalf("expected p2 first, got %+v", hits)
	}
	if u.LLMInputTokens != 0 || u.LLMOutputTokens != 0 || u.EmbedTokens != 0 {
		t.Fatalf("lite recall should cost zero tokens, got %+v", u)
	}
}

func TestStandardPipelineEmbedsAndRetrieves(t *testing.T) {
	meter := NewMeter()
	p, err := BuildPipeline(Config{Profile: "standard", Pricing: DefaultPricing()}, meter, "acme")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	iu, err := p.Ingest(Item{ID: "p2", Text: "The user prefers dark mode in the UI", Class: "basic"})
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if iu.EmbedTokens == 0 {
		t.Fatalf("standard ingest should embed (non-zero embed tokens)")
	}
	_, _ = p.Ingest(Item{ID: "p3", Text: "The user wants code examples in Go", Class: "basic"})

	hits, ru, err := p.Recall(Query{Text: "dark mode theme in the UI", TopK: 2})
	if err != nil {
		t.Fatalf("recall: %v", err)
	}
	if len(hits) == 0 || hits[0].ID != "p2" {
		t.Fatalf("expected p2 via vector search, got %+v", hits)
	}
	if ru.EmbedTokens == 0 {
		t.Fatalf("standard recall should embed the query")
	}
	// metering attributed to the tenant
	if meter.Usage("acme", "basic").EmbedTokens == 0 {
		t.Fatalf("ingest embed usage not metered for tenant acme")
	}
	if cost := meter.CostFor("acme", "basic", DefaultPricing()); cost <= 0 {
		t.Fatalf("expected non-zero metered cost, got %v", cost)
	}
}

func TestRecallContextBudgetCaps(t *testing.T) {
	budget := Budget{RecallContextTokens: 8} // ~enough for one short hit only
	p, err := BuildPipeline(Config{Profile: "lite", Budget: budget}, nil, "")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	ingestAll(t, p, []Item{
		{ID: "a", Text: "alpha user preference about formatting style"},
		{ID: "b", Text: "beta user preference about formatting style"},
		{ID: "c", Text: "gamma user preference about formatting style"},
	})
	hits, _, err := p.Recall(Query{Text: "user preference formatting style", TopK: 5})
	if err != nil {
		t.Fatalf("recall: %v", err)
	}
	// budget of 8 tokens should admit fewer than all 3 matching hits
	total := 0
	for _, h := range hits {
		total += estimateTokens(h.Text)
	}
	if total > 8 {
		t.Fatalf("recall context %d tokens exceeds budget 8", total)
	}
	if len(hits) >= 3 {
		t.Fatalf("expected budget to drop some hits, got %d", len(hits))
	}
}

func TestIngestOverBudgetRejected(t *testing.T) {
	// Tiny ingest budget with the standard (embedding) profile -> reject.
	budget := Budget{IngestMaxUSD: 1e-12, Mode: EnforceReject}
	p, err := BuildPipeline(Config{Profile: "standard", Budget: budget, Pricing: DefaultPricing()}, nil, "")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	if _, err := p.Ingest(Item{ID: "x", Text: "some memory that would need embedding"}); err != ErrIngestOverBudget {
		t.Fatalf("expected ErrIngestOverBudget, got %v", err)
	}
}

func TestIngestOverBudgetDegrades(t *testing.T) {
	budget := Budget{IngestMaxUSD: 1e-12, Mode: EnforceDegrade}
	p, err := BuildPipeline(Config{Profile: "standard", Budget: budget, Pricing: DefaultPricing()}, nil, "")
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	u, err := p.Ingest(Item{ID: "x", Text: "some memory that would need embedding"})
	if err != nil {
		t.Fatalf("degrade should not error: %v", err)
	}
	if u.EmbedTokens != 0 {
		t.Fatalf("degrade should skip embedding, got embed tokens %d", u.EmbedTokens)
	}
}

func TestUsageAdd(t *testing.T) {
	a := Usage{LLMInputTokens: 1, EmbedTokens: 2, ContextTokens: 3, LatencyMs: 1.5}
	a.Add(Usage{LLMInputTokens: 4, EmbedTokens: 5, ContextTokens: 6, LatencyMs: 0.5})
	if a.LLMInputTokens != 5 || a.EmbedTokens != 7 || a.ContextTokens != 9 || a.LatencyMs != 2.0 {
		t.Fatalf("unexpected sum: %+v", a)
	}
}
