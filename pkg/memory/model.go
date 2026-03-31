package memory

import "time"

type BasicKind string

const (
	BasicKindPreference BasicKind = "preference"
	BasicKindHistory    BasicKind = "history"
	BasicKindDecision   BasicKind = "decision"
	BasicKindContext    BasicKind = "context"
)

type BasicMemory struct {
	ID        string    `json:"id"`
	Kind      BasicKind `json:"kind"`
	Scope     string    `json:"scope"`
	Subject   string    `json:"subject"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AdvancedMemory struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	Template  string    `json:"template"`
	Evidence  []string  `json:"evidence"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CognitivePolicy struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Domain    string    `json:"domain"`
	Scenario  string    `json:"scenario"`
	Policy    string    `json:"policy"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type BasicFilter struct {
	Kind    BasicKind
	Scope   string
	Subject string
	Tag     string
	Limit   int
}

type AdvancedFilter struct {
	Template string
	Tag      string
	Limit    int
}

type PolicyFilter struct {
	Role     string
	Domain   string
	Scenario string
	Tag      string
	Limit    int
}
