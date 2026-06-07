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

func TestRunProfilesFrontier(t *testing.T) {
	items := []Item{
		{ID: "p1", Text: "The user prefers concise answer-first responses", Tags: []string{"style"}},
		{ID: "p2", Text: "The user prefers dark mode in the UI", Tags: []string{"ui"}},
		{ID: "p3", Text: "The user wants code examples in Go", Tags: []string{"code"}},
	}
	queries := []Query{
		{Text: "what UI theme does the user want", Tag: "ui", TopK: 3},
		{Text: "how should answers be formatted", Tag: "style", TopK: 3},
	}

	reports := RunProfiles([]string{"student", "standard", "pro"}, items, queries, DefaultPricing())

	// pro requires unavailable LLM providers -> skipped; student + standard run.
	byName := map[string]ProfileReport{}
	for _, r := range reports {
		byName[r.Profile] = r
	}
	student, okS := byName["student"]
	standard, okStd := byName["standard"]
	if !okS || !okStd {
		t.Fatalf("expected student and standard reports, got %v", reports)
	}
	if _, hasPro := byName["pro"]; hasPro {
		t.Fatalf("pro should be skipped (unavailable providers)")
	}

	// The cost-aware frontier: student is free, standard costs (embeddings).
	if student.CostUSD != 0 {
		t.Fatalf("student profile should be $0, got %v", student.CostUSD)
	}
	if standard.CostUSD <= 0 {
		t.Fatalf("standard profile should incur embedding cost, got %v", standard.CostUSD)
	}
	if student.Queries != 2 || standard.Queries != 2 {
		t.Fatalf("both profiles should run all queries")
	}
}
