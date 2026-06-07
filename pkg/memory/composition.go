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
	"regexp"
	"sort"
	"strings"

	"github.com/wasp-project/yazi/pkg/memory/provider"
)

// StoreBackedStore adapts a *Store to the provider.Store contract (Index +
// Retriever), so the deterministic memory pipeline persists through Yazi's real
// KV -> LSM -> S3 path instead of an ephemeral in-memory index. This is how the
// `student` profile keeps Yazi's zero-token, durable behavior while still being
// expressed as a pipeline. Generic provider.Item maps to the typed basic /
// advanced / policy records by Class.
type StoreBackedStore struct {
	store *Store
}

var (
	_ provider.Index     = (*StoreBackedStore)(nil)
	_ provider.Retriever = (*StoreBackedStore)(nil)
)

// NewStoreBackedStore wraps a memory Store as a provider store.
func NewStoreBackedStore(s *Store) *StoreBackedStore { return &StoreBackedStore{store: s} }

func (a *StoreBackedStore) Meta() provider.Meta {
	return provider.Meta{Name: "yazi-store", Cost: "free", Latency: "low"}
}
func (a *StoreBackedStore) Available() bool { return true }

// Upsert persists the item as the typed record matching its Class.
func (a *StoreBackedStore) Upsert(it provider.Item) (provider.Usage, error) {
	var err error
	switch it.Class {
	case "advanced":
		_, err = a.store.PutAdvanced(AdvancedMemory{ID: it.ID, Title: it.Subject, Summary: it.Text, Tags: it.Tags})
	case "policy":
		_, err = a.store.PutPolicy(CognitivePolicy{ID: it.ID, Policy: it.Text, Tags: it.Tags})
	default:
		_, err = a.store.PutBasic(BasicMemory{
			ID: it.ID, Kind: BasicKind(it.Kind), Scope: it.Scope,
			Subject: it.Subject, Content: it.Text, Tags: it.Tags,
		})
	}
	return provider.Usage{}, err
}

// Retrieve pulls candidate basic memories via Yazi's deterministic filters (by
// tag, else all) and keyword-ranks them client-side — zero tokens, the way an
// agent uses Yazi today.
func (a *StoreBackedStore) Retrieve(q provider.Query, topK int) ([]provider.Hit, provider.Usage, error) {
	var (
		list []BasicMemory
		err  error
	)
	if q.Tag != "" {
		list, err = a.store.ListBasic(BasicFilter{Tag: q.Tag, Limit: 50})
	} else {
		list, err = a.store.ListBasic(BasicFilter{Limit: 50})
	}
	if err != nil {
		return nil, provider.Usage{}, err
	}

	qt := compTokens(q.Text)
	type scored struct {
		score float64
		m     BasicMemory
	}
	ranked := make([]scored, 0, len(list))
	for _, m := range list {
		doc := compTokens(m.Content + " " + m.Subject + " " + strings.Join(m.Tags, " "))
		overlap := 0
		for w := range qt {
			if _, ok := doc[w]; ok {
				overlap++
			}
		}
		if overlap == 0 {
			continue
		}
		ranked = append(ranked, scored{float64(overlap) / float64(len(qt)+1), m})
	}
	sort.SliceStable(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })
	if topK > 0 && len(ranked) > topK {
		ranked = ranked[:topK]
	}
	hits := make([]provider.Hit, len(ranked))
	for i, r := range ranked {
		hits[i] = provider.Hit{ID: r.m.ID, Text: r.m.Content, Score: r.score}
	}
	return hits, provider.Usage{}, nil
}

// NewStorePipeline builds a deterministic, persistence-backed pipeline over the
// given Store. This is the `student`-profile composition wired to durable storage.
func NewStorePipeline(s *Store, budget provider.Budget, pricing provider.Pricing, meter *provider.Meter, tenant string) *provider.Pipeline {
	a := NewStoreBackedStore(s)
	stages := provider.Stages{Index: a, Retriever: a}
	if pricing == (provider.Pricing{}) {
		pricing = provider.DefaultPricing()
	}
	return provider.NewPipeline(stages, pricing, budget, meter, tenant)
}

var compWordRe = regexp.MustCompile(`[a-z0-9]+`)

func compTokens(text string) map[string]struct{} {
	out := map[string]struct{}{}
	for _, w := range compWordRe.FindAllString(strings.ToLower(text), -1) {
		out[w] = struct{}{}
	}
	return out
}
