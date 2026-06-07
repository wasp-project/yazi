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

// Interface-only stubs for capabilities Yazi will WRAP rather than reimplement.
// They report Available() == false until wired up to a real backend + credentials,
// so the profile resolver refuses to select them rather than failing mid-request.
// This keeps the capability surface honest: declared, not faked.

// ErrProviderUnavailable is returned when an unwired stub is invoked directly.
var ErrProviderUnavailable = errors.New("provider: not wired up (interface-only stub)")

// LLMExtractor wraps an LLM that extracts/consolidates facts from raw input (C1).
// TODO: implement against a provider SDK; report real LLM token usage.
type LLMExtractor struct{}

func (LLMExtractor) Meta() Meta {
	return Meta{Name: "llm-extractor", Cost: "llm", Latency: "high", Requires: []string{"LLM provider + key"}}
}
func (LLMExtractor) Available() bool { return false }
func (LLMExtractor) Extract(_ Item) ([]Item, Usage, error) {
	return nil, Usage{}, ErrProviderUnavailable
}

// GraphRetriever wraps a graph/temporal store (e.g. Graphiti/Neo4j) for relational
// and time-aware recall.
type GraphRetriever struct{}

func (GraphRetriever) Meta() Meta {
	return Meta{Name: "graph", Cost: "embed", Latency: "medium", Requires: []string{"graph DB"}}
}
func (GraphRetriever) Available() bool { return false }
func (GraphRetriever) Retrieve(_ Query, _ int) ([]Hit, Usage, error) {
	return nil, Usage{}, ErrProviderUnavailable
}

// LLMReranker wraps a cross-encoder/LLM reranker over candidate hits (C4).
type LLMReranker struct{}

func (LLMReranker) Meta() Meta {
	return Meta{Name: "llm-rerank", Cost: "llm", Latency: "high", Requires: []string{"LLM provider + key"}}
}
func (LLMReranker) Available() bool { return false }
func (LLMReranker) Rerank(_ Query, _ []Hit) ([]Hit, Usage, error) {
	return nil, Usage{}, ErrProviderUnavailable
}

// PgVectorIndex / MilvusIndex are placeholders for durable vector backends. They
// share the Index+Retriever interfaces with the in-memory VectorIndex; only the
// storage changes. Unavailable until configured.
type PgVectorIndex struct{}

func (PgVectorIndex) Meta() Meta {
	return Meta{Name: "pgvector", Cost: "free", Latency: "low", Requires: []string{"Postgres + pgvector"}}
}
func (PgVectorIndex) Available() bool                          { return false }
func (PgVectorIndex) Upsert(_ Item) (Usage, error)            { return Usage{}, ErrProviderUnavailable }
func (PgVectorIndex) Retrieve(_ Query, _ int) ([]Hit, Usage, error) {
	return nil, Usage{}, ErrProviderUnavailable
}
