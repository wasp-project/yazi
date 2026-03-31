package storage

import (
        "testing"
)

func TestMemoryStore(t *testing.T) {
        engine := NewKVStore(100, "")
        store := NewMemoryStore(engine)

        err := store.SaveSessionTurn("sid1", 1, SessionTurn{Role: "user", Content: "hello"})
        if err != nil {
                t.Fatal(err)
        }

        turns, err := store.GetSessionTurns("sid1")
        if err != nil {
                t.Fatal(err)
        }
        if len(turns) != 1 || turns[0].Content != "hello" {
                t.Fatalf("unexpected turns: %+v", turns)
        }

        err = store.SaveSessionSummary("sid1", SessionSummary{Summary: "A greeting session"})
        if err != nil {
                t.Fatal(err)
        }

        summary, err := store.GetSessionSummary("sid1")
        if err != nil {
                t.Fatal(err)
        }
        if summary.Summary != "A greeting session" {
                t.Fatalf("unexpected summary: %+v", summary)
        }

        err = store.SaveLongTermMemory("mem1", LongTermMemory{Content: "User likes golang", Tags: []string{"preference"}})
        if err != nil {
                t.Fatal(err)
        }

        memories, err := store.ListLongTermMemories()
        if err != nil {
                t.Fatal(err)
        }
        if len(memories) != 1 || memories[0].Content != "User likes golang" {
                t.Fatalf("unexpected memories: %+v", memories)
        }

        err = store.SaveIssueSolutionMemory("issue1", IssueSolutionMemory{Issue: "panic in yazi", Solution: "check nil pointer"})
        if err != nil {
                t.Fatal(err)
        }

        issues, err := store.ListIssueSolutionMemories()
        if err != nil {
                t.Fatal(err)
        }
        if len(issues) != 1 || issues[0].Solution != "check nil pointer" {
                t.Fatalf("unexpected issues: %+v", issues)
        }
}