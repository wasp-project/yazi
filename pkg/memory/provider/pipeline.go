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

import "errors"

// ErrIngestOverBudget is returned when an ingest exceeds its cost budget and the
// enforcement mode is reject.
var ErrIngestOverBudget = errors.New("provider: ingest exceeds cost budget")

// Stages holds one provider per pipeline stage. Index and Retriever are required;
// every other stage is optional (nil = skipped, which is the deterministic
// default behavior).
type Stages struct {
	Extractor Extractor
	Distiller Distiller
	Embedder  Embedder
	Index     Index
	Retriever Retriever
	Reranker  Reranker
	Budgeter  ContextBudgeter
}

// Pipeline runs the write and read paths over a set of Stages, accumulating
// Usage, enforcing budgets, and recording cost to a Meter.
type Pipeline struct {
	stages  Stages
	pricing Pricing
	budget  Budget
	meter   *Meter
	tenant  string
}

// NewPipeline builds a pipeline. A nil meter disables metering; nil budgeter
// falls back to a plain top-k trim.
func NewPipeline(stages Stages, pricing Pricing, budget Budget, meter *Meter, tenant string) *Pipeline {
	if stages.Budgeter == nil {
		stages.Budgeter = TopKBudgeter{}
	}
	return &Pipeline{stages: stages, pricing: pricing, budget: budget, meter: meter, tenant: tenant}
}

// Ingest runs the write path: extract -> distill -> (budget check) -> embed ->
// index, returning the aggregated usage. When the projected cost exceeds the
// ingest budget, it rejects or degrades (skips embedding) per the budget mode.
func (p *Pipeline) Ingest(item Item) (Usage, error) {
	var u Usage
	if item.Class == "" {
		item.Class = "basic"
	}

	items := []Item{item}
	if p.stages.Extractor != nil {
		out, eu, err := p.stages.Extractor.Extract(item)
		u.Add(eu)
		if err != nil {
			return u, err
		}
		if len(out) > 0 {
			items = out
		}
	}

	if p.stages.Distiller != nil {
		extra, du, err := p.stages.Distiller.Distill(items)
		u.Add(du)
		if err != nil {
			return u, err
		}
		items = append(items, extra...)
	}

	// Budget check before the expensive embedding step.
	embed := p.stages.Embedder != nil
	if embed && p.budget.IngestMaxUSD > 0 {
		projected := p.pricing.Cost(u) + p.pricing.Cost(embedProjection(items))
		if projected > p.budget.IngestMaxUSD {
			if p.budget.mode() == EnforceReject {
				p.record(item.Class, u)
				return u, ErrIngestOverBudget
			}
			embed = false // degrade: skip embedding, index deterministically
		}
	}

	if embed {
		texts := make([]string, len(items))
		for i, it := range items {
			texts[i] = it.Text
		}
		vecs, emu, err := p.stages.Embedder.Embed(texts)
		u.Add(emu)
		if err != nil {
			return u, err
		}
		for i := range items {
			if i < len(vecs) {
				items[i].Vector = vecs[i]
			}
		}
	}

	for _, it := range items {
		iu, err := p.stages.Index.Upsert(it)
		u.Add(iu)
		if err != nil {
			return u, err
		}
	}

	p.record(item.Class, u)
	return u, nil
}

// Recall runs the read path: (embed query) -> retrieve -> rerank -> budget,
// returning the hits and aggregated usage (including the downstream context
// tokens of the returned hits).
func (p *Pipeline) Recall(q Query) ([]Hit, Usage, error) {
	var u Usage
	topK := q.TopK
	if topK <= 0 {
		topK = 5
	}

	if p.stages.Embedder != nil {
		vecs, eu, err := p.stages.Embedder.Embed([]string{q.Text})
		u.Add(eu)
		if err != nil {
			return nil, u, err
		}
		if len(vecs) > 0 {
			q.Vector = vecs[0]
		}
	}

	// Over-fetch a little so the reranker/budgeter have candidates to work with.
	fetch := topK
	if p.stages.Reranker != nil {
		fetch = topK * 3
	}
	hits, ru, err := p.stages.Retriever.Retrieve(q, fetch)
	u.Add(ru)
	if err != nil {
		return nil, u, err
	}

	if p.stages.Reranker != nil {
		reranked, rru, err := p.stages.Reranker.Rerank(q, hits)
		u.Add(rru)
		if err != nil {
			return nil, u, err
		}
		hits = reranked
	}

	if len(hits) > topK {
		hits = hits[:topK]
	}

	hits, bu := p.stages.Budgeter.Budget(hits, p.budget.RecallContextTokens)
	u.Add(bu)

	ctx := 0
	for _, h := range hits {
		ctx += estimateTokens(h.Text)
	}
	u.ContextTokens += ctx

	p.record("recall", u)
	return hits, u, nil
}

func (p *Pipeline) record(class string, u Usage) {
	if p.meter != nil {
		p.meter.Record(p.tenant, class, u)
	}
}

// embedProjection estimates the embedding usage of a set of items for the
// pre-embedding budget check.
func embedProjection(items []Item) Usage {
	var u Usage
	for _, it := range items {
		u.EmbedTokens += estimateTokens(it.Text)
	}
	return u
}
