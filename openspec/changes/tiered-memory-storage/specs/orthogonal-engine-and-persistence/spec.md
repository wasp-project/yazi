## ADDED Requirements

### Requirement: Engine And Durability Backend Are Independent

The storage engine selection and the durable backend selection SHALL be
independent configuration axes that compose. Selecting a disk-based engine SHALL
NOT disable cloud persistence, and selecting cloud persistence SHALL NOT force an
in-memory engine.

#### Scenario: Disk engine with cloud backend

- **WHEN** the server is configured with the log-structured engine and the
  object-storage backend
- **THEN** the server starts, writes are durable locally, and sealed segments are
  additionally persisted to object storage

#### Scenario: Disk engine with local backend

- **WHEN** the server is configured with the log-structured engine and the local
  backend
- **THEN** the server behaves as before this change, with segments persisted to
  the local data directory

#### Scenario: In-memory engine with cloud backend

- **WHEN** the server is configured with the in-memory engine and the
  object-storage backend
- **THEN** whole-store snapshot persistence to object storage continues to work
  as before this change

### Requirement: Existing Configurations Remain Valid

Configuration files written before this change SHALL continue to start the
server with equivalent behavior, without edits.

#### Scenario: Previous local-mode config still works

- **WHEN** the server starts with the previously published local-mode
  configuration
- **THEN** it starts successfully and its data remains readable

#### Scenario: Previous cloud-mode config still works

- **WHEN** the server starts with the previously published cloud-mode
  configuration
- **THEN** it starts successfully and continues to persist to the configured
  bucket

### Requirement: Fail Fast On Incomplete Backend Configuration

When a durable backend is selected but its required settings are missing, the
server SHALL fail at startup with an error naming the missing settings, and
SHALL NOT silently fall back to another backend.

#### Scenario: Missing object-storage settings abort startup

- **WHEN** the object-storage backend is selected without a bucket or credentials
- **THEN** the server exits with an error naming what is missing and does not
  start serving from local storage instead
