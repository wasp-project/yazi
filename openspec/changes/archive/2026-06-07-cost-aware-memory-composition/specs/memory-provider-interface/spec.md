## ADDED Requirements

### Requirement: Composable Memory Pipeline

The memory engine SHALL process writes and reads through an ordered pipeline of
provider stages rather than a fixed code path. The write pipeline SHALL be
`Extractor → Distiller → Embedder → Index`; the read pipeline SHALL be
`Retriever → Reranker → ContextBudgeter`, with an optional `KVCacheReuse`
capability. Every stage SHALL have a default provider, and the default for every
stage SHALL be the existing deterministic, zero-token behavior so that an
unconfigured pipeline is functionally identical to the current memory store.

#### Scenario: Default pipeline reproduces current behavior

- **WHEN** no providers are configured for any stage
- **THEN** writes store structured records and reads return key/index results
  exactly as the current memory store does, spending zero LLM/embedding tokens

#### Scenario: A stage is swapped without touching others

- **WHEN** the `Embedder` stage is bound to a real provider
- **THEN** the write pipeline produces embeddings at that stage while every other
  stage keeps its existing provider, with no code change to those stages

### Requirement: Provider Interfaces

The system SHALL define a Go interface for each pipeline capability: `Extractor`,
`Distiller`, `Embedder`, `Index`, `Retriever`, `Reranker`, and `KVCacheReuse`.
Each operation SHALL return a `Usage` value describing the resources it consumed
(LLM input/output tokens, embedding tokens, latency). Each provider SHALL expose
metadata describing its cost characteristics, latency class, and external
requirements. Adding a new provider SHALL require only implementing the relevant
interface and registering it.

#### Scenario: Operations report usage

- **WHEN** any provider performs an ingest or recall operation
- **THEN** it returns a `Usage` record that the pipeline can aggregate, even when
  the usage is zero (deterministic providers report zero tokens)

#### Scenario: Provider declares its requirements

- **WHEN** a provider requires an external service (e.g. an embedder or vector store)
- **THEN** its metadata lists those requirements so the system can validate them
  before the provider is activated

### Requirement: Deterministic Default Providers

The system SHALL provide deterministic default implementations that wrap the
existing memory behavior: a no-op `Extractor`/`Distiller`/`Embedder`/`Reranker`,
a key+secondary-index `Index` and `Retriever`, and a top-k `ContextBudgeter`.
These defaults SHALL report zero token/embedding usage.

#### Scenario: Default providers cost nothing

- **WHEN** the deterministic default providers run a full write and read
- **THEN** the aggregated `Usage` reports zero LLM and zero embedding tokens

### Requirement: Interface-Only Stubs for Wrapped Systems

The system SHALL allow capabilities that wrap external systems (LLM extraction,
vector search, graph/temporal retrieval, reranking) to exist as registered
providers whose interface is defined even when no implementation is wired up. An
unavailable provider SHALL be detectable so the system can refuse to select it
rather than failing mid-operation.

#### Scenario: Selecting an unavailable provider fails fast

- **WHEN** a profile selects a provider whose external requirements are not met
- **THEN** the system reports the missing requirement at configuration time and
  does not start serving with a broken pipeline
