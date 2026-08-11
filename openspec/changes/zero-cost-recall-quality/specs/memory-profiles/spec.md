## MODIFIED Requirements

### Requirement: Named Memory Profiles

The system SHALL support selecting a memory composition by a named profile in
configuration: `lite`, `standard`, `pro`, or `custom`. Each named profile
SHALL bind a concrete provider to every pipeline stage. The `lite` profile
SHALL use only deterministic, zero-token providers, and its retrieval SHALL be
inverted-index candidate selection with BM25 relevance combined with recency and
access-frequency weighting over the durable store — not a scan of a fixed-size
prefix of the corpus. `standard` and `pro` SHALL add progressively more capable
(and more costly) providers. Selecting a profile SHALL require no code change.

#### Scenario: Student profile is the cheap default

- **WHEN** `memory.profile` is `lite` or unset
- **THEN** the pipeline uses deterministic providers only and incurs zero
  LLM/embedding cost

#### Scenario: Cheap default still ranks well

- **WHEN** `memory.profile` is `lite` and a recall is issued against a corpus
  larger than any fixed scan window
- **THEN** candidates are selected from the inverted index and ranked by BM25
  with recency and frequency weighting, at zero LLM and embedding cost

#### Scenario: Higher profile adds capability

- **WHEN** `memory.profile` is `standard`
- **THEN** the pipeline additionally runs an embedder and a vector index/retriever
  as defined by that profile
