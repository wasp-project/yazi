## MODIFIED Requirements

### Requirement: Named Memory Profiles

The system SHALL support selecting a memory composition by a named profile in
configuration: `lite`, `standard`, `pro`, or `custom`. Each named profile
SHALL bind a concrete provider to every pipeline stage. The `lite` profile
SHALL use only deterministic, zero-token providers, and its retrieval SHALL be
inverted-index candidate selection with BM25 relevance combined with recency and
access-frequency weighting over the durable store. The `standard` profile SHALL
add write-time embedding into a persistent vector index, so its recall embeds only
the query. The `pro` profile SHALL additionally bind real LLM-backed extraction
and reranking providers and SHALL be selectable whenever those providers are
configured. Selecting a profile SHALL require no code change.

#### Scenario: Student profile is the cheap default

- **WHEN** `memory.profile` is `lite` or unset
- **THEN** the pipeline uses deterministic providers only and incurs zero
  LLM/embedding cost

#### Scenario: Cheap default still ranks well

- **WHEN** `memory.profile` is `lite` and a recall is issued against a corpus
  larger than any fixed scan window
- **THEN** candidates are selected from the inverted index and ranked by BM25
  with recency and frequency weighting, at zero LLM and embedding cost

#### Scenario: Standard profile pays embedding at write time

- **WHEN** `memory.profile` is `standard`
- **THEN** eligible memories are embedded on ingest into a persistent index, and
  a recall reports embedding usage for the query only

#### Scenario: Pro profile is selectable when configured

- **WHEN** `memory.profile` is `pro` and LLM providers are configured and
  reachable
- **THEN** the server starts and recall runs LLM reranking with its token usage
  metered

#### Scenario: Pro profile fails clearly when unconfigured

- **WHEN** `memory.profile` is `pro` and no LLM provider is configured
- **THEN** startup fails with an error naming the missing configuration
