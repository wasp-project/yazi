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

import "sync"

// Pricing converts Usage into a monetary estimate (USD per 1M tokens). It is
// kept separate from Usage so totals can be re-priced without re-running ops.
type Pricing struct {
	LLMInputPerM  float64 `json:"llmInputPerM" yaml:"llmInputPerM"`
	LLMOutputPerM float64 `json:"llmOutputPerM" yaml:"llmOutputPerM"`
	EmbedPerM     float64 `json:"embedPerM" yaml:"embedPerM"`
}

// DefaultPricing approximates a small hosted model plus embeddings.
func DefaultPricing() Pricing {
	return Pricing{LLMInputPerM: 0.15, LLMOutputPerM: 0.60, EmbedPerM: 0.02}
}

// Cost returns the USD estimate for a Usage. Context tokens are downstream prompt
// cost (paid by every system) and are NOT counted here.
func (p Pricing) Cost(u Usage) float64 {
	return (float64(u.LLMInputTokens)*p.LLMInputPerM+
		float64(u.LLMOutputTokens)*p.LLMOutputPerM+
		float64(u.EmbedTokens)*p.EmbedPerM) / 1_000_000
}

// EnforcementMode decides what happens when an operation would exceed a budget.
type EnforcementMode string

const (
	// EnforceReject fails the operation when it would exceed a budget.
	EnforceReject EnforcementMode = "reject"
	// EnforceDegrade runs a cheaper path instead (e.g. skip embedding, trim context).
	EnforceDegrade EnforcementMode = "degrade"
)

// Budget bounds the recurring cost of operations. A zero field means unbounded.
type Budget struct {
	// RecallContextTokens caps the tokens of context a recall may return.
	RecallContextTokens int `json:"recallContextTokens" yaml:"recallContextTokens"`
	// IngestMaxUSD caps the estimated memory-system cost of a single ingest.
	IngestMaxUSD float64 `json:"ingestMaxUSD" yaml:"ingestMaxUSD"`
	// Mode selects reject vs degrade. Empty defaults to degrade.
	Mode EnforcementMode `json:"mode" yaml:"mode"`
}

func (b Budget) mode() EnforcementMode {
	if b.Mode == EnforceReject {
		return EnforceReject
	}
	return EnforceDegrade
}

// MeterKey identifies a metering bucket.
type MeterKey struct {
	Tenant string
	Class  string
}

// Meter aggregates Usage per (tenant, memory class). It is safe for concurrent use.
type Meter struct {
	mu    sync.Mutex
	usage map[MeterKey]Usage
}

// NewMeter returns an empty meter.
func NewMeter() *Meter {
	return &Meter{usage: map[MeterKey]Usage{}}
}

// Record adds usage to the (tenant, class) bucket.
func (m *Meter) Record(tenant, class string, u Usage) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	k := MeterKey{Tenant: tenant, Class: class}
	cur := m.usage[k]
	cur.Add(u)
	m.usage[k] = cur
}

// Usage returns the accumulated usage for a bucket.
func (m *Meter) Usage(tenant, class string) Usage {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.usage[MeterKey{Tenant: tenant, Class: class}]
}

// Totals returns a copy of every bucket's usage.
func (m *Meter) Totals() map[MeterKey]Usage {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make(map[MeterKey]Usage, len(m.usage))
	for k, v := range m.usage {
		out[k] = v
	}
	return out
}

// CostFor prices a bucket. Re-pricing recomputes from recorded usage; ops are not
// re-run.
func (m *Meter) CostFor(tenant, class string, p Pricing) float64 {
	return p.Cost(m.Usage(tenant, class))
}
