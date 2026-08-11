## ADDED Requirements

### Requirement: Versioned HTTP Memory Endpoints

The server SHALL expose an HTTP/JSON API under `/v1/memory` providing create,
read, filtered list and delete for the basic, advanced and policy memory
classes, using the same payload schemas and filter fields as the existing CLI.
Request and response bodies SHALL be JSON.

#### Scenario: Write and read a basic memory over HTTP

- **WHEN** a client POSTs a basic memory payload to `/v1/memory/basic`
- **THEN** the server responds with the generated memory id, and a subsequent
  GET of that id returns the stored record

#### Scenario: Filtered list over HTTP

- **WHEN** a client requests a list of basic memories with a `tag` filter
- **THEN** the server returns only records carrying that tag, served from the
  secondary index rather than a full scan

#### Scenario: Delete removes record and index entries

- **WHEN** a client DELETEs an existing memory id
- **THEN** the record and all of its secondary index entries are removed, and a
  subsequent list with a matching filter does not return it

### Requirement: Recall Endpoint Reports Cost

The recall endpoint SHALL return the ranked hits together with the aggregated
resource `Usage` and the estimated cost in USD for that request, in the same
shape the CLI reports today.

#### Scenario: Recall response includes a usage block

- **WHEN** a client POSTs a recall query
- **THEN** the response contains the hits, the profile used, the `usage` object
  (LLM input/output tokens, embed tokens, context tokens, latency) and
  `costUSD`

#### Scenario: Default profile costs nothing

- **WHEN** a recall is served with the `lite` profile
- **THEN** the reported `costUSD` is zero and no embedding or LLM tokens are
  recorded

### Requirement: Tenant Selection Over HTTP

The API SHALL accept a tenant identifier per request via a request header and
resolve it through the tenant service. When the header is absent, the server's
configured tenant SHALL apply.

#### Scenario: Per-request tenant header

- **WHEN** a request carries a tenant header naming `acme`
- **THEN** the operation is scoped to `acme` and cannot read or write another
  tenant's records

### Requirement: HTTP Error Semantics

The API SHALL return conventional status codes with a JSON error body: 400 for a
malformed payload or filter, 404 for a missing memory id, 422 for a request that
violates a configured budget, and 503 for a selected provider whose requirements
are unmet. The error body SHALL contain a human-readable message.

#### Scenario: Unknown id returns 404

- **WHEN** a client GETs a memory id that does not exist
- **THEN** the server responds 404 with a JSON error message and does not return
  a partial record

#### Scenario: Unavailable provider returns 503

- **WHEN** a recall requests a profile whose providers are unavailable
- **THEN** the server responds 503 naming the unavailable provider and its
  unmet requirements

### Requirement: Safe Default Binding

The HTTP API SHALL bind to the loopback interface by default and SHALL require
explicit configuration to listen on a non-loopback address.

#### Scenario: Default listener is loopback only

- **WHEN** the server starts with no HTTP address configured
- **THEN** the HTTP API accepts connections on `127.0.0.1` and is not reachable
  from another host
