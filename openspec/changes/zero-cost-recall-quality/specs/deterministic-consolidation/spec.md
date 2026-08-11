## ADDED Requirements

### Requirement: Exact-Duplicate Collapse

When an incoming memory is an exact duplicate of a live memory — identical
normalized content, kind, scope and subject — the system SHALL update the
existing record rather than create a second one, and SHALL return the existing
id. Deduplication SHALL be exact and SHALL NOT use similarity heuristics or an
LLM.

#### Scenario: Writing the same fact twice yields one record

- **WHEN** the same memory payload is written twice
- **THEN** one record exists, its update timestamp reflects the second write, and
  both writes return the same id

#### Scenario: Near-duplicates are kept separate

- **WHEN** two memories differ in wording but describe a similar thing
- **THEN** both are stored as distinct records

### Requirement: Preference Supersession With Lineage

When a new `preference` memory is written for a scope and subject already covered
by a live `preference`, the system SHALL mark the previous record as superseded by
the new one instead of deleting it. Only the `preference` kind SHALL be
auto-superseded.

#### Scenario: New preference supersedes the old one

- **WHEN** a preference for the same scope and subject as an existing preference
  is written
- **THEN** the previous record is marked superseded by the new record's id and
  remains stored

#### Scenario: Superseded records are excluded from recall by default

- **WHEN** a recall matches both a superseded record and the record that
  superseded it
- **THEN** only the current record is returned unless superseded records are
  explicitly requested

#### Scenario: Superseded history remains retrievable

- **WHEN** a caller explicitly requests superseded records
- **THEN** the previous preference is returned with the id of the record that
  superseded it

#### Scenario: Other kinds are never auto-superseded

- **WHEN** two `history`, `context` or `decision` memories share a scope and
  subject
- **THEN** both remain live and neither is marked superseded

### Requirement: Consolidation Is Auditable

Every duplicate collapse and every supersession SHALL be logged with both record
ids so the transformation can be audited.

#### Scenario: Supersession is logged

- **WHEN** a preference supersedes another
- **THEN** a log entry records the superseded id and the superseding id
