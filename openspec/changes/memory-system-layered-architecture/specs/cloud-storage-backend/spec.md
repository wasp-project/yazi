## ADDED Requirements

### Requirement: Cloud Persistence Backend Behind PersistentStorage

The system SHALL support cloud-vendor persistence backends that implement the existing `storage.PersistentStorage` interface (`Write([]byte) (int, error)` and `Read([]byte) (int, error)`). A cloud backend SHALL be selectable through the storage-class configuration and SHALL be usable by the storage `Manager` in the same way as the local-disk backend, so the memory and storage-engine layers require no changes to use it.

#### Scenario: Cloud backend selected by configuration

- **WHEN** the server starts with `storage: s3`
- **THEN** the storage factory constructs the S3 backend as the active `PersistentStorage` and the memory store persists through it

#### Scenario: Encoded snapshot round-trips through the backend

- **WHEN** the storage layer writes an encoded snapshot to the cloud backend and later reads it back on startup
- **THEN** the decoded contents match what was written

### Requirement: AWS S3 Object Storage Implementation

The system SHALL provide an AWS S3 implementation of `storage.PersistentStorage` that writes the encoded store snapshot to a configured S3 bucket and object key, and reads it back on load. The implementation SHALL accept configuration for bucket, region, credentials, and an optional custom endpoint (to support S3-compatible services).

#### Scenario: Write persists snapshot to S3

- **WHEN** the S3 backend `Write` is called with an encoded snapshot
- **THEN** an object is created or replaced at the configured bucket and key with that content

#### Scenario: Read loads snapshot from S3 on startup

- **WHEN** the server starts in cloud mode and the configured object exists
- **THEN** the S3 backend `Read` returns the object's bytes so the store can decode them

#### Scenario: Custom endpoint for S3-compatible storage

- **WHEN** an endpoint override is configured
- **THEN** the S3 client targets that endpoint instead of the default AWS endpoint

### Requirement: Cloud Backend Configuration

The configuration SHALL include an `s3:` block (bucket, region, key/prefix, optional endpoint, and credential source) that is read when `storage: s3` is selected. Missing or invalid required S3 configuration SHALL cause startup to fail with a clear error rather than silently falling back to local storage.

#### Scenario: Missing bucket fails fast

- **WHEN** `storage: s3` is selected but no bucket is configured
- **THEN** server startup fails with an error identifying the missing S3 bucket configuration

### Requirement: Extension Point for Vector Storage

The cloud-storage abstraction SHALL leave a documented extension point so a future AWS vector-storage backend can be added by implementing the same persistence interface (or a clearly defined sibling interface) without altering the memory data model or the tenant layer.

#### Scenario: Future vector backend added without core changes

- **WHEN** a developer adds an AWS vector-storage backend
- **THEN** they implement the cloud-storage persistence contract and register it via the storage-class configuration, leaving the memory and tenant layers unchanged
