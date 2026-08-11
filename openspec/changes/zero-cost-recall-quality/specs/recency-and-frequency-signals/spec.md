## ADDED Requirements

### Requirement: Recency Weighting With Per-Class Half-Lives

The final recall score SHALL combine textual relevance with a time-decay factor
derived from the memory's age, using a configurable half-life that MAY differ per
memory kind. Decay SHALL modulate relevance multiplicatively so that a memory
with no textual match is never surfaced by recency alone.

#### Scenario: Recent memory wins at equal relevance

- **WHEN** two memories have identical textual relevance and one is much older
- **THEN** the newer memory ranks higher

#### Scenario: Recency cannot manufacture a match

- **WHEN** a memory has zero textual relevance to the query
- **THEN** it is not returned regardless of how recently it was written

#### Scenario: Preferences decay more slowly than transient context

- **WHEN** a preference and a context memory of the same age and relevance are
  ranked with default settings
- **THEN** the preference retains more of its score than the context memory

### Requirement: Access-Frequency Weighting

The final score SHALL include a dampened contribution from how often a memory has
been returned by recall, so repeatedly useful memories rank higher. The
contribution SHALL be sub-linear in the access count.

#### Scenario: Frequently recalled memory ranks higher

- **WHEN** two memories are equal in relevance and age and one has been recalled
  many times
- **THEN** the frequently recalled memory ranks higher

#### Scenario: Frequency boost is dampened

- **WHEN** a memory's access count grows by an order of magnitude
- **THEN** its frequency contribution grows sub-linearly rather than
  proportionally

### Requirement: Scores Are Explainable

Each returned hit SHALL carry the components of its score — textual relevance,
recency factor and frequency factor — so ranking decisions can be inspected.

#### Scenario: Hit exposes its score components

- **WHEN** a recall returns hits
- **THEN** each hit reports its relevance, recency and frequency contributions
  alongside the final score

### Requirement: Ranking Parameters Are Configurable

Half-lives per memory kind, the frequency damping factor, and the BM25 parameters
SHALL be configurable, with documented defaults that apply when unset.

#### Scenario: Defaults apply when unconfigured

- **WHEN** no ranking parameters are configured
- **THEN** the documented default values are used and recall works without any
  ranking configuration
