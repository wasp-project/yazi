package memory

import (
	"testing"

	"github.com/wasp-project/yazi/pkg/policy"
	"github.com/wasp-project/yazi/pkg/storage"
)

func TestBasicMemoryIndexes(t *testing.T) {
	kv := storage.NewKVStore(1024, policy.KeyPolicy(""))
	store := NewStore(kv)
	id, err := store.PutBasic(BasicMemory{
		Kind:    BasicKindPreference,
		Scope:   "user",
		Subject: "ui",
		Content: "dark-mode",
		Tags:    []string{"theme", "ui"},
	})
	if err != nil {
		t.Fatalf("put basic: %v", err)
	}
	got, err := store.GetBasic(id)
	if err != nil {
		t.Fatalf("get basic: %v", err)
	}
	if got.Content != "dark-mode" {
		t.Fatalf("unexpected content: %s", got.Content)
	}
	list, err := store.ListBasic(BasicFilter{Tag: "ui"})
	if err != nil {
		t.Fatalf("list basic: %v", err)
	}
	if len(list) != 1 || list[0].ID != id {
		t.Fatalf("unexpected list size: %d", len(list))
	}
	if err := store.DeleteBasic(id); err != nil {
		t.Fatalf("delete basic: %v", err)
	}
	list, err = store.ListBasic(BasicFilter{Tag: "ui"})
	if err != nil {
		t.Fatalf("list basic after delete: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty list after delete")
	}
}

func TestAdvancedMemoryIndexes(t *testing.T) {
	kv := storage.NewKVStore(1024, policy.KeyPolicy(""))
	store := NewStore(kv)
	id, err := store.PutAdvanced(AdvancedMemory{
		Title:    "planning",
		Summary:  "step-by-step",
		Template: "plan",
		Tags:     []string{"workflow"},
	})
	if err != nil {
		t.Fatalf("put advanced: %v", err)
	}
	list, err := store.ListAdvanced(AdvancedFilter{Template: "plan"})
	if err != nil {
		t.Fatalf("list advanced: %v", err)
	}
	if len(list) != 1 || list[0].ID != id {
		t.Fatalf("unexpected list size: %d", len(list))
	}
	if err := store.DeleteAdvanced(id); err != nil {
		t.Fatalf("delete advanced: %v", err)
	}
}

func TestPolicyIndexes(t *testing.T) {
	kv := storage.NewKVStore(1024, policy.KeyPolicy(""))
	store := NewStore(kv)
	id, err := store.PutPolicy(CognitivePolicy{
		Role:     "doctor",
		Domain:   "medical",
		Scenario: "triage",
		Policy:   "clarify symptoms before advice",
		Tags:     []string{"safety"},
	})
	if err != nil {
		t.Fatalf("put policy: %v", err)
	}
	list, err := store.ListPolicy(PolicyFilter{Role: "doctor"})
	if err != nil {
		t.Fatalf("list policy: %v", err)
	}
	if len(list) != 1 || list[0].ID != id {
		t.Fatalf("unexpected list size: %d", len(list))
	}
	if err := store.DeletePolicy(id); err != nil {
		t.Fatalf("delete policy: %v", err)
	}
}
