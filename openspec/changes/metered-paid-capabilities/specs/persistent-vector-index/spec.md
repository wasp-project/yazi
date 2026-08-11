## ADDED Requirements

### Requirement: Embedding Is Paid At Write Time

Vectors SHALL be computed when a memory is ingested and persisted alongside it.
Recall SHALL NOT re-embed stored memories; it SHALL embed only the query.

#### Scenario: Recall embeds only the query

- **WHEN** a semantic recall is issued against a store of many embedded memories
- **THEN** the reported embedding usage corresponds to the query text alone and
  does not grow with the number of stored memories

#### Scenario: Repeated recalls do not repeat embedding cost

- **WHEN** the same query is issued twice against an unchanged corpus
- **THEN** each recall reports only query-embedding usage, and no stored memory is
  re-embedded

#### Scenario: Ingest reports the embedding it paid for

- **WHEN** a memory eligible for embedding is ingested
- **THEN** the ingest operation reports the embedding usage for that memory

### Requirement: Durable Vector Index

Stored vectors SHALL survive process restarts, SHALL be namespaced per tenant,
and SHALL be removed when their memory is deleted.

#### Scenario: Vectors survive a restart

- **WHEN** memories are embedded, the server is restarted, and a semantic recall
  is issued
- **THEN** the recall returns results without re-embedding any stored memory

#### Scenario: Deleting a memory removes its vector

- **WHEN** an embedded memory is deleted
- **THEN** its vector is removed and it can no longer be returned by semantic
  recall

#### Scenario: Vectors are tenant-isolated

- **WHEN** two tenants embed memories
- **THEN** neither tenant's semantic recall can return the other's memories

### Requirement: Vectors Record Their Embedder Identity

Each stored vector SHALL record the embedder model identifier and vector
dimension. The system SHALL refuse to compare vectors produced by different
embedders or dimensions, and SHALL report the cost of a re-embedding pass before
running one.

#### Scenario: Embedder change is detected

- **WHEN** the configured embedder differs from the one that produced the stored
  vectors
- **THEN** semantic recall reports the mismatch instead of returning results
  computed across incompatible vectors

#### Scenario: Re-embedding is projected before it runs

- **WHEN** an operator requests re-embedding after an embedder change
- **THEN** the system reports the projected token cost and proceeds only on
  confirmation

### Requirement: Bounded Candidate Set For Vector Scoring

When a text query is present, semantic recall SHALL bound the set of vectors it
scores by pre-filtering with the deterministic index, and SHALL report which
retrieval path produced each hit.

#### Scenario: Hybrid retrieval bounds the scan

- **WHEN** a semantic recall carries a text query
- **THEN** candidates are pre-filtered by the deterministic index before vector
  scoring

#### Scenario: Hits identify their retrieval path

- **WHEN** a recall returns hits from both the deterministic and the vector path
- **THEN** each hit indicates which path produced it
