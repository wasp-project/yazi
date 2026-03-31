package storage

import (
        "encoding/json"
        "errors"
        "fmt"
        "strings"
)

// SessionTurn represents a single turn in a session
type SessionTurn struct {
        Role    string `json:"role"`
        Content string `json:"content"`
}

// SessionSummary represents a summary of a session
type SessionSummary struct {
        Summary string `json:"summary"`
}

// LongTermMemory represents a long-term memory item
type LongTermMemory struct {
        Content string `json:"content"`
        Tags    []string `json:"tags,omitempty"`
}

// IssueSolutionMemory represents an issue and its solution
type IssueSolutionMemory struct {
        Issue    string `json:"issue"`
        Solution string `json:"solution"`
}

// MemoryStore is a wrapper around KVStore to support OpenClaw memory types
type MemoryStore struct {
        store KVStore
}

func NewMemoryStore(store KVStore) *MemoryStore {
        return &MemoryStore{
                store: store,
        }
}

// SaveSessionTurn saves a single turn in a session
func (m *MemoryStore) SaveSessionTurn(sid string, turn int, data SessionTurn) error {
        key := fmt.Sprintf("session:turn:%s:%d", sid, turn)
        val, err := json.Marshal(data)
        if err != nil {
                return err
        }
        return m.store.Set(key, string(val))
}

// GetSessionTurns retrieves all turns for a session, ordered by turn number
func (m *MemoryStore) GetSessionTurns(sid string) ([]SessionTurn, error) {
        keys, err := m.store.Keys()
        if err != nil {
                return nil, err
        }

        prefix := fmt.Sprintf("session:turn:%s:", sid)
        var turnKeys []string
        for _, k := range keys {
                if strings.HasPrefix(k, prefix) {
                        turnKeys = append(turnKeys, k)
                }
        }

        if len(turnKeys) == 0 {
                return []SessionTurn{}, nil
        }

        // TODO: MGet might not guarantee order matching the turn numbers naturally if keys are unordered.
        // For strict order, we should extract turn numbers and sort them. But for simplicity we just return.
        vals, err := m.store.MGet(turnKeys)
        if err != nil {
                return nil, err
        }

        var turns []SessionTurn
        for _, v := range vals {
                if v == "" {
                        continue
                }
                var turn SessionTurn
                if err := json.Unmarshal([]byte(v), &turn); err != nil {
                        continue // Skip invalid JSON
                }
                turns = append(turns, turn)
        }

        return turns, nil
}

// SaveSessionSummary saves the summary of a session
func (m *MemoryStore) SaveSessionSummary(sid string, summary SessionSummary) error {
        key := fmt.Sprintf("session:summary:%s", sid)
        val, err := json.Marshal(summary)
        if err != nil {
                return err
        }
        return m.store.Set(key, string(val))
}

// GetSessionSummary retrieves the summary of a session
func (m *MemoryStore) GetSessionSummary(sid string) (*SessionSummary, error) {
        key := fmt.Sprintf("session:summary:%s", sid)
        val, err := m.store.Get(key)
        if err != nil {
                return nil, err
        }
        if val == "" {
                return nil, errors.New("not found")
        }

        var summary SessionSummary
        if err := json.Unmarshal([]byte(val), &summary); err != nil {
                return nil, err
        }
        return &summary, nil
}

// SaveLongTermMemory saves a long-term memory item
func (m *MemoryStore) SaveLongTermMemory(id string, memory LongTermMemory) error {
        key := fmt.Sprintf("memory:longterm:%s", id)
        val, err := json.Marshal(memory)
        if err != nil {
                return err
        }
        return m.store.Set(key, string(val))
}

// ListLongTermMemories retrieves all long-term memory items
func (m *MemoryStore) ListLongTermMemories() ([]LongTermMemory, error) {
        keys, err := m.store.Keys()
        if err != nil {
                return nil, err
        }

        prefix := "memory:longterm:"
        var memKeys []string
        for _, k := range keys {
                if strings.HasPrefix(k, prefix) {
                        memKeys = append(memKeys, k)
                }
        }

        if len(memKeys) == 0 {
                return []LongTermMemory{}, nil
        }

        vals, err := m.store.MGet(memKeys)
        if err != nil {
                return nil, err
        }

        var memories []LongTermMemory
        for _, v := range vals {
                if v == "" {
                        continue
                }
                var memory LongTermMemory
                if err := json.Unmarshal([]byte(v), &memory); err != nil {
                        continue
                }
                memories = append(memories, memory)
        }

        return memories, nil
}

// SaveIssueSolutionMemory saves an issue-solution memory item
func (m *MemoryStore) SaveIssueSolutionMemory(id string, memory IssueSolutionMemory) error {
        key := fmt.Sprintf("memory:issue_solution:%s", id)
        val, err := json.Marshal(memory)
        if err != nil {
                return err
        }
        return m.store.Set(key, string(val))
}

// ListIssueSolutionMemories retrieves all issue-solution memory items
func (m *MemoryStore) ListIssueSolutionMemories() ([]IssueSolutionMemory, error) {
        keys, err := m.store.Keys()
        if err != nil {
                return nil, err
        }

        prefix := "memory:issue_solution:"
        var memKeys []string
        for _, k := range keys {
                if strings.HasPrefix(k, prefix) {
                        memKeys = append(memKeys, k)
                }
        }

        if len(memKeys) == 0 {
                return []IssueSolutionMemory{}, nil
        }

        vals, err := m.store.MGet(memKeys)
        if err != nil {
                return nil, err
        }

        var memories []IssueSolutionMemory
        for _, v := range vals {
                if v == "" {
                        continue
                }
                var memory IssueSolutionMemory
                if err := json.Unmarshal([]byte(v), &memory); err != nil {
                        continue
                }
                memories = append(memories, memory)
        }

        return memories, nil
}
