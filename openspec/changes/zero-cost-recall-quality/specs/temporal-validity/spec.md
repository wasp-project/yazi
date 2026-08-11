## ADDED Requirements

### Requirement: Validity Interval On Memory Records

Every memory class SHALL support optional `validFrom` and `validUntil`
timestamps. Absent bounds SHALL mean the memory is always valid. These fields
SHALL be optional so existing stored records remain valid without modification.

#### Scenario: Record without bounds is always valid

- **WHEN** a memory is stored with no validity bounds
- **THEN** it is considered valid at every point in time

#### Scenario: Bounds are round-tripped

- **WHEN** a memory is written with validity bounds and read back
- **THEN** the stored bounds are returned unchanged

### Requirement: Point-In-Time Filtering

Recall and list operations SHALL accept an optional point-in-time and SHALL
return only memories valid at that instant. Without a point-in-time, operations
SHALL return memories valid now.

#### Scenario: Historical query excludes later facts

- **WHEN** a caller asks for memories as of a past instant
- **THEN** memories whose validity begins after that instant are excluded

#### Scenario: Expired memory is excluded from current recall

- **WHEN** a memory's validity has ended and a recall is issued without a
  point-in-time
- **THEN** that memory is not returned

#### Scenario: Point-in-time filtering costs no tokens

- **WHEN** a point-in-time recall runs on the deterministic profile
- **THEN** it is served by comparison against stored timestamps and reports zero
  LLM and embedding tokens

### Requirement: Temporal Filtering Composes With Ranking And Supersession

Validity filtering SHALL apply before ranking, and SHALL compose with supersession
so that a superseded-and-expired record is excluded once, not double-counted in
scoring.

#### Scenario: Only valid candidates are ranked

- **WHEN** a recall with a point-in-time matches both valid and invalid memories
- **THEN** only the valid ones are scored and returned
