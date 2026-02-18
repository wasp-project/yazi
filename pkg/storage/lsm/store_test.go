package lsm

import (
	"testing"

	"github.com/wasp-project/yazi/pkg/config"
)

func TestWALRecovery(t *testing.T) {
	dir := t.TempDir()
	conf := config.LSMConfig{Dir: dir, MemtableMaxEntries: 100}
	store, err := NewStore(conf)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	if err := store.Set("a", "1"); err != nil {
		t.Fatalf("set a: %v", err)
	}
	if err := store.Set("b", "2"); err != nil {
		t.Fatalf("set b: %v", err)
	}
	if err := store.Del("a"); err != nil {
		t.Fatalf("del a: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	store2, err := NewStore(conf)
	if err != nil {
		t.Fatalf("new store2: %v", err)
	}
	if _, err := store2.Get("a"); err == nil {
		t.Fatalf("expected a to be deleted")
	}
	if v, err := store2.Get("b"); err != nil || v != "2" {
		t.Fatalf("expected b=2, got %q err=%v", v, err)
	}
	_ = store2.Close()
}

func TestFlushToSSTable(t *testing.T) {
	dir := t.TempDir()
	conf := config.LSMConfig{Dir: dir, MemtableMaxEntries: 2}
	store, err := NewStore(conf)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	if err := store.Set("k1", "v1"); err != nil {
		t.Fatalf("set k1: %v", err)
	}
	if err := store.Set("k2", "v2"); err != nil {
		t.Fatalf("set k2: %v", err)
	}
	if err := store.Set("k3", "v3"); err != nil {
		t.Fatalf("set k3: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	store2, err := NewStore(conf)
	if err != nil {
		t.Fatalf("new store2: %v", err)
	}
	for k, v := range map[string]string{"k1": "v1", "k2": "v2", "k3": "v3"} {
		got, err := store2.Get(k)
		if err != nil || got != v {
			t.Fatalf("expected %s=%s, got %q err=%v", k, v, got, err)
		}
	}
	_ = store2.Close()
}
