## ADDED Requirements

### Requirement: Three Storage Tiers

The system SHALL classify each segment into one of three tiers: hot (blocks
resident in the memory cache), warm (segment file present on local disk with its
sparse index resident), and cold (segment present only in object storage, with
only its manifest entry resident).

#### Scenario: Tier of every segment is reported

- **WHEN** the operator inspects storage status
- **THEN** each segment is reported with its tier, size, key range, last access
  time and access count

### Requirement: Classification Uses Recency And Frequency

Tier classification SHALL consider both how recently and how frequently a
segment has been read, not age alone. A segment that is old but frequently read
SHALL NOT be demoted ahead of a segment that is recent but never read.

#### Scenario: Frequently read old segment stays warm

- **WHEN** an old segment is read on most recall requests and a newer segment has
  never been read
- **THEN** the demotion pass demotes the newer unread segment before the older
  frequently read one

#### Scenario: Untouched segments are demoted

- **WHEN** a segment has not been read within the configured window and local
  capacity is under pressure
- **THEN** the segment is uploaded if not already present in object storage and
  its local file is removed

### Requirement: Promotion On Access With A Bounded Local Cache

Reading a key whose segment is cold SHALL make that segment locally available for
subsequent reads, and the local segment cache SHALL be bounded by a configured
size, evicting by the same recency-and-frequency score.

#### Scenario: Cold segment is promoted after access

- **WHEN** a cold segment is read
- **THEN** the read succeeds and the segment becomes warm for subsequent reads

#### Scenario: Local cache respects its size limit

- **WHEN** promotions would exceed the configured local cache size
- **THEN** the least valuable warm segments are evicted so the limit is not
  exceeded, and evicted segments remain readable from object storage

### Requirement: Tiering Never Loses Data

Demotion SHALL delete a local segment file only after the corresponding object is
confirmed durable in the backend.

#### Scenario: Failed upload aborts demotion

- **WHEN** uploading a segment to object storage fails during demotion
- **THEN** the local file is retained, the segment stays warm, and the failure is
  logged

#### Scenario: Every key survives a full tiering cycle

- **WHEN** a corpus is written, fully demoted to cold, and then read back
- **THEN** every key returns its original value

### Requirement: Tiering Is Observable

The system SHALL expose per-tier counters — segment count, bytes, read hits and
misses, promotions and demotions — so tier behavior can be measured rather than
assumed.

#### Scenario: Counters distinguish warm and cold reads

- **WHEN** recall requests are served partly from warm and partly from cold
  segments
- **THEN** the reported counters attribute each read to its tier
