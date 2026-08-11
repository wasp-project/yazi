## ADDED Requirements

### Requirement: Embedding Eligibility Rules

The system SHALL embed only memories that satisfy configured eligibility rules,
rather than every ingested memory. Rules SHALL support selecting by memory class,
kind, tag, minimum content length, and a minimum access count. The default SHALL
embed distilled advanced-class memories only.

#### Scenario: Ineligible memory is not embedded

- **WHEN** a memory that matches no eligibility rule is ingested on a profile
  with an embedder
- **THEN** no embedding usage is reported for it and it is stored with
  deterministic indexing only

#### Scenario: Default embeds distilled memories only

- **WHEN** no eligibility rules are configured and both a basic and an advanced
  memory are ingested
- **THEN** only the advanced memory is embedded

#### Scenario: Rules are configurable

- **WHEN** an operator configures eligibility by tag
- **THEN** memories carrying that tag are embedded and others are not

### Requirement: Memories Can Earn Embedding Through Use

A memory that reaches the configured access-count threshold SHALL become eligible
for embedding, and its embedding SHALL be performed once and reported as ingest
cost.

#### Scenario: Frequently recalled memory becomes embedded

- **WHEN** a memory not initially eligible is recalled enough times to cross the
  configured threshold
- **THEN** it is embedded once, the cost is reported, and subsequent recalls do
  not re-embed it

### Requirement: Coverage Is Visible, Not Silent

The system SHALL report what fraction of stored memories are embedded, so the
coverage limits of semantic recall are explicit.

#### Scenario: Coverage is reported

- **WHEN** an operator requests memory statistics
- **THEN** the report includes the number and fraction of memories that have
  vectors

#### Scenario: Unembedded memories remain recallable

- **WHEN** a memory has no vector
- **THEN** it is still returned by the deterministic retrieval path when it
  matches
