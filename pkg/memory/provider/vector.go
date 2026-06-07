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
	"hash/fnv"
	"math"
	"sort"
	"strings"
)

// Reference "standard" providers: a dependency-free local embedder and an
// in-memory cosine vector index. They prove the embed->vector-search pipeline
// end-to-end and offline. The same interfaces accept real providers later — a
// hosted/Ollama embedder, or pgvector/Milvus Lite — without other changes.

// LocalEmbedder is a deterministic, offline bag-of-words hashing embedder. It is
// the local-first reference Embedder: no model download, no network, reproducible.
// It reports embed-token usage so cost metering is exercised (a real embedding
// model swaps in behind the same interface and reports real usage).
type LocalEmbedder struct {
	Dim int
}

// NewLocalEmbedder returns a local embedder with the given dimensionality
// (defaults to 256).
func NewLocalEmbedder(dim int) *LocalEmbedder {
	if dim <= 0 {
		dim = 256
	}
	return &LocalEmbedder{Dim: dim}
}

func (e *LocalEmbedder) Meta() Meta {
	return Meta{Name: "local-embed", Cost: "embed", Latency: "low"}
}
func (e *LocalEmbedder) Available() bool { return true }

func (e *LocalEmbedder) Embed(texts []string) ([][]float32, Usage, error) {
	var u Usage
	out := make([][]float32, len(texts))
	for i, t := range texts {
		out[i] = e.embedOne(t)
		u.EmbedTokens += estimateTokens(t)
	}
	return out, u, nil
}

func (e *LocalEmbedder) embedOne(text string) []float32 {
	v := make([]float32, e.Dim)
	for _, w := range wordRe.FindAllString(strings.ToLower(text), -1) {
		h := fnv.New32a()
		_, _ = h.Write([]byte(w))
		idx := int(h.Sum32()) % e.Dim
		if idx < 0 {
			idx += e.Dim
		}
		v[idx] += 1
	}
	// L2-normalize so cosine similarity is a dot product.
	var norm float64
	for _, x := range v {
		norm += float64(x) * float64(x)
	}
	if norm > 0 {
		inv := float32(1 / math.Sqrt(norm))
		for i := range v {
			v[i] *= inv
		}
	}
	return v
}

// VectorIndex is an in-memory cosine-similarity index implementing both Index and
// Retriever. It is the reference vector store; pgvector / Milvus Lite implement
// the same interfaces for durable/scalable deployments.
type VectorIndex struct {
	items []Item
}

// NewVectorIndex returns an empty in-memory vector index.
func NewVectorIndex() *VectorIndex { return &VectorIndex{} }

func (v *VectorIndex) Meta() Meta {
	return Meta{Name: "vector", Cost: "free", Latency: "low"}
}
func (v *VectorIndex) Available() bool { return true }

// Upsert stores the item (which must already carry a Vector from the embedder).
func (v *VectorIndex) Upsert(item Item) (Usage, error) {
	for i := range v.items {
		if v.items[i].ID == item.ID && item.ID != "" {
			v.items[i] = item
			return Usage{}, nil
		}
	}
	v.items = append(v.items, item)
	return Usage{}, nil
}

// Retrieve ranks stored items by cosine similarity to the query vector. If the
// query has no vector (no embedder configured), it returns nothing.
func (v *VectorIndex) Retrieve(q Query, topK int) ([]Hit, Usage, error) {
	if len(q.Vector) == 0 {
		return nil, Usage{}, nil
	}
	type scored struct {
		score float64
		item  Item
	}
	ranked := make([]scored, 0, len(v.items))
	for _, it := range v.items {
		if len(it.Vector) == 0 {
			continue
		}
		s := cosine(q.Vector, it.Vector)
		if s <= 0 {
			continue
		}
		ranked = append(ranked, scored{s, it})
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

func cosine(a, b []float32) float64 {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	var dot float64
	for i := 0; i < n; i++ {
		dot += float64(a[i]) * float64(b[i])
	}
	return dot // vectors are pre-normalized by the embedder
}
