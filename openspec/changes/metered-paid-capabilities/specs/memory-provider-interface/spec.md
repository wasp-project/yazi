## ADDED Requirements

### Requirement: Pre-Call Cost Projection

Every provider whose operation can incur token cost SHALL be able to project the
`Usage` of an operation **without** performing it, so the pipeline can evaluate
budgets before spending. Projections SHALL be conservative, rounding up rather
than down.

#### Scenario: Projection requires no external call

- **WHEN** a costly provider is asked to project the usage of an operation
- **THEN** it returns an estimated `Usage` without contacting any external service

#### Scenario: Pipeline evaluates budgets before dispatch

- **WHEN** an operation's projected cost exceeds a declared budget
- **THEN** the pipeline rejects or degrades the operation before the costly
  provider is called

#### Scenario: Deterministic providers project zero

- **WHEN** a deterministic provider is asked to project usage
- **THEN** it reports zero LLM and embedding tokens

### Requirement: Persistent Incremental Index Semantics

An `Index` provider SHALL be able to declare that it persists its state and
updates incrementally. The pipeline SHALL NOT re-ingest stored items to serve a
recall against a persistent index.

#### Scenario: Recall does not re-ingest a persistent index

- **WHEN** a recall runs against a provider that declares persistent state
- **THEN** the pipeline queries the existing index and performs no ingest of
  stored items

#### Scenario: Non-persistent index is still supported

- **WHEN** a provider does not declare persistence
- **THEN** the pipeline may populate it before use and reports the resulting usage

## MODIFIED Requirements

### Requirement: Interface-Only Stubs for Wrapped Systems

The system SHALL allow capabilities that wrap external systems (LLM extraction,
vector search, graph/temporal retrieval, reranking) to exist as registered
providers whose interface is defined even when no implementation is wired up. An
unavailable provider SHALL be detectable so the system can refuse to select it
rather than failing mid-operation. A provider that has a real implementation SHALL
determine its availability from actual configuration and reachability rather than
reporting a fixed value, and SHALL name the missing settings when unavailable.

#### Scenario: Selecting an unavailable provider fails fast

- **WHEN** a profile selects a provider whose external requirements are not met
- **THEN** the system reports the missing requirement at configuration time and
  does not start serving with a broken pipeline

#### Scenario: Implemented provider derives availability from configuration

- **WHEN** an implemented provider's endpoint is configured and reachable
- **THEN** it reports itself available, and when the endpoint is absent or
  unreachable it reports itself unavailable and names the missing settings
