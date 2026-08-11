## ADDED Requirements

### Requirement: Segments Are Immutable Objects

Sealed segments SHALL be written to the durable backend as individually named,
immutable objects. An existing segment object SHALL never be rewritten in place.

#### Scenario: New data uploads only new segments

- **WHEN** new memories are written and a segment is sealed
- **THEN** only the newly sealed segment is uploaded, and previously uploaded
  segment objects are not re-transmitted

#### Scenario: Upload volume is proportional to new data

- **WHEN** a store holding a large corpus seals one small segment
- **THEN** the bytes transmitted to the durable backend are proportional to that
  segment, not to the total corpus size

### Requirement: Manifest Is The Commit Point

A manifest object SHALL record the set of live segments with their key ranges,
sizes and tier. Publishing the manifest SHALL be the point at which newly
uploaded segments become visible. Startup SHALL reconstruct state from the
manifest.

#### Scenario: Crash before manifest publish loses nothing committed

- **WHEN** the process fails after uploading a segment but before publishing the
  manifest
- **THEN** on restart the store reflects the previously published manifest, all
  acknowledged writes are recoverable from the write-ahead log, and the orphaned
  object is not treated as live data

#### Scenario: Orphaned objects are reclaimed

- **WHEN** a garbage-collection pass runs and finds objects not referenced by the
  manifest and older than the retention grace period
- **THEN** those objects are deleted and the deletion is logged

### Requirement: On-Demand Ranged Reads From Object Storage

Reading a key from a segment that resides only in object storage SHALL fetch a
bounded byte range covering the required block, not the whole object or the whole
dataset.

#### Scenario: Cold point read is a ranged fetch

- **WHEN** a key is read from a segment present only in object storage
- **THEN** the backend performs a ranged read of that segment's index and the
  single block containing the key

#### Scenario: Cold read result is correct

- **WHEN** a key written before its segment was demoted to object storage is read
- **THEN** the value returned equals the value originally written

### Requirement: Object Storage Backend Supports Segment Operations

The object-storage backend SHALL support putting a named object, getting an
object, getting a byte range of an object, listing objects under a prefix, and
deleting an object, using the existing dependency-free request signing.

#### Scenario: Backend round-trips a segment

- **WHEN** a segment is put and then read back by name and by byte range against
  an S3-compatible endpoint
- **THEN** the full object and the ranged read both return the expected bytes

#### Scenario: Missing object is reported distinctly

- **WHEN** a segment named in a request does not exist
- **THEN** the backend reports a not-found condition distinguishable from a
  transport error
