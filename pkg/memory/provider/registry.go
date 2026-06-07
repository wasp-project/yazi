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
	"fmt"
	"strings"
)

// Store is a backend that provides both Index and Retriever (they are inherently
// coupled — a retriever reads what its index wrote), so a profile selects them
// together as one "store".
type Store interface {
	Index
	Retriever
}

// Config selects a memory composition. The named Profile drives everything;
// when Profile == "custom", the per-stage fields are used (each defaulting to the
// cheap option). Pricing/Budget apply to any profile.
type Config struct {
	Profile   string  // student | standard | pro | custom
	Embedder  string  // none | local
	Store     string  // keyword | vector | pgvector
	Extractor string  // none | llm
	Reranker  string  // none | llm
	Budget    Budget
	Pricing   Pricing
}

// DefaultConfig is the cheap, deterministic student profile.
func DefaultConfig() Config {
	return Config{Profile: "student", Pricing: DefaultPricing()}
}

// Profiles lists the known named profiles.
func Profiles() []string { return []string{"student", "standard", "pro", "custom"} }

// resolveNames expands a profile into concrete per-stage provider names.
func resolveNames(c Config) (embedder, store, extractor, reranker string, err error) {
	switch strings.ToLower(c.Profile) {
	case "", "student":
		return "none", "keyword", "none", "none", nil
	case "standard":
		return "local", "vector", "none", "none", nil
	case "pro":
		return "local", "vector", "llm", "llm", nil
	case "custom":
		return orDefault(c.Embedder, "none"), orDefault(c.Store, "keyword"),
			orDefault(c.Extractor, "none"), orDefault(c.Reranker, "none"), nil
	default:
		return "", "", "", "", fmt.Errorf("provider: unknown profile %q (have %v)", c.Profile, Profiles())
	}
}

func orDefault(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}

func buildStore(name string) (Store, error) {
	switch name {
	case "keyword":
		return NewKeywordStore(), nil
	case "vector":
		return NewVectorIndex(), nil
	case "pgvector":
		return PgVectorIndex{}, nil
	default:
		return nil, fmt.Errorf("provider: unknown store %q", name)
	}
}

func buildEmbedder(name string) (Embedder, error) {
	switch name {
	case "none":
		return nil, nil
	case "local":
		return NewLocalEmbedder(0), nil
	default:
		return nil, fmt.Errorf("provider: unknown embedder %q", name)
	}
}

func buildExtractor(name string) (Extractor, error) {
	switch name {
	case "none":
		return nil, nil // pipeline skips a nil extractor (deterministic pass-through)
	case "llm":
		return LLMExtractor{}, nil
	default:
		return nil, fmt.Errorf("provider: unknown extractor %q", name)
	}
}

func buildReranker(name string) (Reranker, error) {
	switch name {
	case "none":
		return nil, nil // pipeline skips a nil reranker
	case "llm":
		return LLMReranker{}, nil
	default:
		return nil, fmt.Errorf("provider: unknown reranker %q", name)
	}
}

// requireAvailable fails fast when a selected provider's external requirements
// are unmet, so the system never starts serving with a broken pipeline.
func requireAvailable(role string, p Provider) error {
	if p == nil || p.Available() {
		return nil
	}
	m := p.Meta()
	return fmt.Errorf("provider: %s %q is unavailable; requires: %s",
		role, m.Name, strings.Join(m.Requires, ", "))
}

// BuildPipeline resolves a Config into a ready Pipeline, validating that every
// selected provider is available. meter/tenant may be empty for a standalone
// pipeline.
func BuildPipeline(c Config, meter *Meter, tenant string) (*Pipeline, error) {
	embName, storeName, extName, rerName, err := resolveNames(c)
	if err != nil {
		return nil, err
	}

	embedder, err := buildEmbedder(embName)
	if err != nil {
		return nil, err
	}
	store, err := buildStore(storeName)
	if err != nil {
		return nil, err
	}
	extractor, err := buildExtractor(extName)
	if err != nil {
		return nil, err
	}
	reranker, err := buildReranker(rerName)
	if err != nil {
		return nil, err
	}

	for role, p := range map[string]Provider{
		"embedder":  embedder,
		"store":     store,
		"extractor": extractor,
		"reranker":  reranker,
	} {
		if err := requireAvailable(role, p); err != nil {
			return nil, err
		}
	}

	pricing := c.Pricing
	if pricing == (Pricing{}) {
		pricing = DefaultPricing()
	}

	stages := Stages{
		Extractor: extractor,
		Distiller: NoopDistiller{},
		Embedder:  embedder,
		Index:     store,
		Retriever: store,
		Reranker:  reranker,
		Budgeter:  TopKBudgeter{},
	}
	return NewPipeline(stages, pricing, c.Budget, meter, tenant), nil
}
