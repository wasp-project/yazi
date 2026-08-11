## ADDED Requirements

### Requirement: Block-Indexed SSTable Format

Sealed SSTables SHALL store entries in sorted, fixed-size-target blocks with a
sparse index mapping each block's first key to its offset and length, plus a
footer recording format version, index location, entry count and key range.
Opening a table SHALL read only its footer and sparse index.

#### Scenario: Opening a table does not load its entries

- **WHEN** a store opens an SSTable containing many entries
- **THEN** only the footer and sparse index are read into memory, and no entry
  value is resident until a read requires its block

#### Scenario: Point read fetches exactly one block

- **WHEN** a key present in a sealed table is read and its block is not cached
- **THEN** the store binary-searches the sparse index and reads exactly one
  block from the underlying medium

#### Scenario: Absent key is rejected without a block read

- **WHEN** a key outside a table's recorded key range is read
- **THEN** the table is skipped without reading any block

### Requirement: Configurable Resident Memory Budget

The storage engine SHALL accept a resident-memory budget and SHALL keep the sum
of memtable, sparse indexes and cached blocks within that budget by evicting
cached blocks. Resident memory SHALL NOT grow proportionally to total stored
data.

#### Scenario: Memory stays bounded as data grows

- **WHEN** the stored dataset grows to many times the configured resident budget
- **THEN** the engine's resident memory remains within the configured budget and
  all keys remain readable

#### Scenario: Block cache evicts under pressure

- **WHEN** cached blocks would exceed the budget
- **THEN** the least valuable blocks are evicted and subsequent reads of those
  blocks fetch them again from the underlying medium

### Requirement: Format Version Migration

The engine SHALL detect SSTables written in the previous full-map format and
rewrite them into the block-indexed format on startup, without data loss. When
migration is declined by configuration, the engine SHALL refuse to start rather
than read or modify the older files.

#### Scenario: Legacy tables are migrated on first start

- **WHEN** a data directory containing previous-format tables is opened
- **THEN** the tables are rewritten in the new format, the migration is logged,
  and every previously stored key is still readable with its original value

#### Scenario: Migration can be refused

- **WHEN** migration is disabled and previous-format tables are present
- **THEN** the engine exits with an explanatory error and leaves the files
  untouched

### Requirement: Read Correctness Under Compaction And Eviction

For any sequence of writes, deletes, compactions and cache evictions, a read
SHALL return the value of the most recent write for that key, or report absence
if the most recent operation was a delete.

#### Scenario: Randomized operations match an in-memory oracle

- **WHEN** a randomized sequence of writes, deletes, compactions and evictions is
  applied to both the engine and an in-memory reference map
- **THEN** every subsequent read returns the same result from both
