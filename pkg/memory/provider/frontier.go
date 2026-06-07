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

// Frontier helpers run the same workload through several profiles and report
// each profile's cost, so a caller can place profiles on the accuracy-vs-cost
// curve. The Go benchmark/ harness (or a CLI) can call RunProfiles to produce a
// continuous, in-process comparison without standing up external services.

// ProfileReport summarizes one profile's run over a workload.
type ProfileReport struct {
	Profile    string
	Ingests    int
	Queries    int
	Hits       int
	IngestUsage Usage
	RecallUsage Usage
	CostUSD    float64
	ContextTokens int
}

// RunProfiles ingests every item and runs every query through each named profile,
// returning a report per profile. Pricing converts usage to cost. A profile that
// fails to build (e.g. requires an unavailable provider) is skipped in the result.
func RunProfiles(profiles []string, items []Item, queries []Query, pricing Pricing) []ProfileReport {
	if pricing == (Pricing{}) {
		pricing = DefaultPricing()
	}
	out := make([]ProfileReport, 0, len(profiles))
	for _, name := range profiles {
		p, err := BuildPipeline(Config{Profile: name, Pricing: pricing}, nil, "")
		if err != nil {
			continue
		}
		rep := ProfileReport{Profile: name}
		for _, it := range items {
			u, err := p.Ingest(it)
			if err != nil {
				continue
			}
			rep.Ingests++
			rep.IngestUsage.Add(u)
		}
		for _, q := range queries {
			hits, u, err := p.Recall(q)
			if err != nil {
				continue
			}
			rep.Queries++
			rep.Hits += len(hits)
			rep.RecallUsage.Add(u)
		}
		var total Usage
		total.Add(rep.IngestUsage)
		total.Add(rep.RecallUsage)
		rep.CostUSD = pricing.Cost(total)
		rep.ContextTokens = rep.RecallUsage.ContextTokens
		out = append(out, rep)
	}
	return out
}
