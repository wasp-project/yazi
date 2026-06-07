# memory-profiles Specification

## Purpose
TBD - created by archiving change cost-aware-memory-composition. Update Purpose after archive.
## Requirements
### Requirement: Named Memory Profiles

The system SHALL support selecting a memory composition by a named profile in
configuration: `student`, `standard`, `pro`, or `custom`. Each named profile
SHALL bind a concrete provider to every pipeline stage. The `student` profile
SHALL use only deterministic, zero-token providers (key+index retrieval, local
storage). `standard` and `pro` SHALL add progressively more capable (and more
costly) providers. Selecting a profile SHALL require no code change.

#### Scenario: Student profile is the cheap default

- **WHEN** `memory.profile` is `student` or unset
- **THEN** the pipeline uses deterministic providers only and incurs zero
  LLM/embedding cost

#### Scenario: Higher profile adds capability

- **WHEN** `memory.profile` is `standard`
- **THEN** the pipeline additionally runs an embedder and a vector index/retriever
  as defined by that profile

### Requirement: Custom Composition

When `memory.profile` is `custom`, the system SHALL let the operator bind each
pipeline stage independently (e.g. choose the extractor, embedder, index,
retriever, reranker). Unspecified stages in a custom profile SHALL fall back to
the deterministic default for that stage.

#### Scenario: Custom profile overrides selected stages only

- **WHEN** a custom profile sets only the embedder and index
- **THEN** those two stages use the chosen providers and all other stages use
  their deterministic defaults

### Requirement: Default Resolves to Student

When no `memory` configuration is present, the system SHALL resolve to the
`student` profile so existing deployments are unaffected.

#### Scenario: Absent config is backward compatible

- **WHEN** a server starts with no `memory` block configured
- **THEN** it behaves identically to the current deterministic memory store

### Requirement: Per-Tenant Profile Selection

The system SHALL allow a profile to be associated with a tenant so different
tenants can run different compositions within the same deployment, composing with
the existing tenant key-namespacing.

#### Scenario: Two tenants run different profiles

- **WHEN** tenant `acme` is on `student` and tenant `globex` is on `standard`
- **THEN** each tenant's memory operations run its own pipeline and cost profile,
  and their data remains isolated by tenant prefix

