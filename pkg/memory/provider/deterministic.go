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
	"regexp"
	"sort"
	"strings"
)

// Deterministic, zero-token providers. These reproduce Yazi's original memory
// behavior and form the `lite` profile. They report zero LLM/embed usage.

// --- no-op write/read stages ---

// NoopExtractor passes the raw item through unchanged.
type NoopExtractor struct{}

func (NoopExtractor) Meta() Meta            { return Meta{Name: "noop", Cost: "free", Latency: "low"} }
func (NoopExtractor) Available() bool       { return true }
func (NoopExtractor) Extract(raw Item) ([]Item, Usage, error) {
	return []Item{raw}, Usage{}, nil
}

// NoopDistiller produces no distilled memories.
type NoopDistiller struct{}

func (NoopDistiller) Meta() Meta      { return Meta{Name: "noop", Cost: "free", Latency: "low"} }
func (NoopDistiller) Available() bool { return true }
func (NoopDistiller) Distill(items []Item) ([]Item, Usage, error) {
	return nil, Usage{}, nil
}

// NoopReranker returns hits unchanged.
type NoopReranker struct{}

func (NoopReranker) Meta() Meta      { return Meta{Name: "noop", Cost: "free", Latency: "low"} }
func (NoopReranker) Available() bool { return true }
func (NoopReranker) Rerank(_ Query, hits []Hit) ([]Hit, Usage, error) {
	return hits, Usage{}, nil
}

// TopKBudgeter trims hits to fit a context-token budget (0 = unlimited). Hits are
// kept in order until adding the next would exceed the budget.
type TopKBudgeter struct{}

func (TopKBudgeter) Meta() Meta      { return Meta{Name: "topk", Cost: "free", Latency: "low"} }
func (TopKBudgeter) Available() bool { return true }
func (TopKBudgeter) Budget(hits []Hit, maxContextTokens int) ([]Hit, Usage) {
	if maxContextTokens <= 0 {
		return hits, Usage{}
	}
	out := make([]Hit, 0, len(hits))
	used := 0
	for _, h := range hits {
		t := estimateTokens(h.Text)
		if used+t > maxContextTokens {
			break
		}
		used += t
		out = append(out, h)
	}
	return out, Usage{}
}

// --- deterministic Index + Retriever (keyword/tag, zero tokens) ---

var wordRe = regexp.MustCompile(`[a-z0-9]+`)

func tokenizeSet(text string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, w := range wordRe.FindAllString(strings.ToLower(text), -1) {
		out[w] = struct{}{}
	}
	return out
}

// KeywordStore is an in-process index providing both Index and Retriever. It
// ranks by query/term overlap with a light boost for a matching tag, mirroring
// how an agent uses Yazi's deterministic filters. It reports zero token usage.
type KeywordStore struct {
	items []Item
}

// NewKeywordStore returns an empty deterministic store.
func NewKeywordStore() *KeywordStore { return &KeywordStore{} }

func (s *KeywordStore) Meta() Meta {
	return Meta{Name: "keyword", Cost: "free", Latency: "low"}
}
func (s *KeywordStore) Available() bool { return true }

// Upsert stores or replaces an item by id.
func (s *KeywordStore) Upsert(item Item) (Usage, error) {
	for i := range s.items {
		if s.items[i].ID == item.ID && item.ID != "" {
			s.items[i] = item
			return Usage{}, nil
		}
	}
	s.items = append(s.items, item)
	return Usage{}, nil
}

// Retrieve ranks items by keyword overlap with the query, boosting a matching tag.
func (s *KeywordStore) Retrieve(q Query, topK int) ([]Hit, Usage, error) {
	qt := tokenizeSet(q.Text)
	type scored struct {
		score float64
		item  Item
	}
	ranked := make([]scored, 0, len(s.items))
	for _, it := range s.items {
		doc := tokenizeSet(it.Text + " " + it.Subject + " " + strings.Join(it.Tags, " "))
		overlap := 0
		for w := range qt {
			if _, ok := doc[w]; ok {
				overlap++
			}
		}
		if overlap == 0 {
			continue
		}
		score := float64(overlap) / float64(max(len(qt), 1))
		if q.Tag != "" {
			for _, t := range it.Tags {
				if t == q.Tag {
					score += 0.25
					break
				}
			}
		}
		ranked = append(ranked, scored{score, it})
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })
	if topK > 0 && len(ranked) > topK {
		ranked = ranked[:topK]
	}
	hits := make([]Hit, len(ranked))
	for i, r := range ranked {
		hits[i] = Hit{ID: r.item.ID, Text: r.item.Text, Score: r.score}
	}
	return hits, Usage{}, nil
}
