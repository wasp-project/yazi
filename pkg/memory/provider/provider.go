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

// Package provider is Yazi's cost-aware memory composition layer.
//
// The memory engine runs writes and reads through a Pipeline of swappable
// provider stages. Each provider reports a Usage (tokens, latency) so cost can
// be metered and bounded, and declares its requirements so the system can refuse
// to start with an unavailable provider. The default providers are deterministic
// and zero-token, reproducing Yazi's original behavior; richer (and costlier)
// providers are opt-in via profiles. Yazi orchestrates these capabilities — it
// wraps external systems (embedders, vector stores, LLMs) rather than
// reimplementing them.
package provider

// Usage is the resource a single provider operation consumed. It is the common
// currency the Pipeline aggregates and the Meter prices. Deterministic providers
// report zero tokens (latency may still be non-zero). The field shapes mirror
// benchmark/framework/metrics.py so engine and benchmark numbers are comparable.
type Usage struct {
	LLMInputTokens  int     `json:"llmInputTokens"`
	LLMOutputTokens int     `json:"llmOutputTokens"`
	EmbedTokens     int     `json:"embedTokens"`
	ContextTokens   int     `json:"contextTokens"`
	LatencyMs       float64 `json:"latencyMs"`
}

// Add accumulates another Usage into this one.
func (u *Usage) Add(o Usage) {
	u.LLMInputTokens += o.LLMInputTokens
	u.LLMOutputTokens += o.LLMOutputTokens
	u.EmbedTokens += o.EmbedTokens
	u.ContextTokens += o.ContextTokens
	u.LatencyMs += o.LatencyMs
}

// Item is a unit of memory flowing through the write pipeline.
type Item struct {
	ID      string
	Text    string
	Kind    string // preference | history | decision | context
	Scope   string
	Subject string
	Tags    []string
	Class   string    // basic | advanced | policy (used for per-class metering)
	Vector  []float32 // populated by an Embedder stage, if any
}

// Query is a recall request flowing through the read pipeline.
type Query struct {
	Text   string
	Tag    string
	TopK   int
	Vector []float32 // populated by an Embedder stage during recall, if any
}

// Hit is a retrieved memory.
type Hit struct {
	ID    string
	Text  string
	Score float64
}

// Meta describes a provider's cost characteristics for reporting and selection.
type Meta struct {
	Name     string   // e.g. "noop", "keyword", "local-embed", "vector", "llm"
	Cost     string   // qualitative: "free" | "embed" | "llm"
	Latency  string   // "low" | "medium" | "high"
	Requires []string // external services needed to run (empty = none)
}

// Provider is embedded by every stage interface.
type Provider interface {
	Meta() Meta
	// Available reports whether this provider can run in the current environment
	// (its Requires are satisfied). Unavailable providers must not be selected.
	Available() bool
}

// --- Write-path stages ---

// Extractor turns raw input into one or more structured memory items. The
// default is a no-op that passes the item through unchanged at zero cost; a real
// implementation calls an LLM (and reports tokens).
type Extractor interface {
	Provider
	Extract(raw Item) ([]Item, Usage, error)
}

// Distiller summarizes a set of items into higher-level memories (e.g. advanced
// memory). The default returns nothing at zero cost.
type Distiller interface {
	Provider
	Distill(items []Item) ([]Item, Usage, error)
}

// Embedder encodes texts into vectors. Absent (nil) in deterministic profiles.
type Embedder interface {
	Provider
	Embed(texts []string) ([][]float32, Usage, error)
}

// Index stores an item for later retrieval.
type Index interface {
	Provider
	Upsert(item Item) (Usage, error)
}

// --- Read-path stages ---

// Retriever returns candidate hits for a query.
type Retriever interface {
	Provider
	Retrieve(q Query, topK int) ([]Hit, Usage, error)
}

// Reranker reorders candidate hits. The default returns them unchanged at zero
// cost; a real implementation uses a cross-encoder or an LLM.
type Reranker interface {
	Provider
	Rerank(q Query, hits []Hit) ([]Hit, Usage, error)
}

// ContextBudgeter trims hits to fit a downstream prompt token budget.
type ContextBudgeter interface {
	Provider
	Budget(hits []Hit, maxContextTokens int) ([]Hit, Usage)
}

// estimateTokens is the shared ~4-chars/token heuristic used for context-token
// accounting and the local embedder.
func estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	n := (len(text) + 3) / 4
	if n < 1 {
		return 1
	}
	return n
}
