## ADDED Requirements

### Requirement: Server-Owned Memory Service

The server SHALL own a transport-agnostic memory service that constructs the
memory store over its local `KVStore` and exposes one operation per memory
action (put, get, list, delete, recall) for each memory class (basic, advanced,
policy). Clients SHALL NOT construct memory keys or secondary indexes.

#### Scenario: A client performs a memory write without knowing key layout

- **WHEN** a client issues a memory put through any transport
- **THEN** the server builds the `/_memory/...` record key and all secondary
  index entries, and the client sends only the memory payload

#### Scenario: Recall completes in a single client round-trip

- **WHEN** a client issues a recall request
- **THEN** the server performs all index reads, record loads and ranking
  internally and returns the hits in one response

### Requirement: Server-Enforced Tenant Namespacing

The memory service SHALL resolve the tenant through the configured
`tenant.Service` and apply the resulting key prefix server-side for every memory
operation. A tenant supplied by the caller SHALL be resolved through the same
service rather than trusted directly as a key prefix.

#### Scenario: Two tenants are isolated by the server

- **WHEN** tenant `acme` writes a memory and tenant `globex` lists memories
- **THEN** `globex` receives no records belonging to `acme`, and the isolation
  holds regardless of which transport each client used

#### Scenario: Bypass mode preserves the original key layout

- **WHEN** no tenant is configured and no tenant is supplied by the caller
- **THEN** records are stored under the un-prefixed `/_memory/...` keys, byte
  identical to the layout produced before this change

### Requirement: Tenant Prefix Is Well-Formed

Derived tenant keys SHALL contain no empty path segment.

#### Scenario: Namespaced key has no duplicated separator

- **WHEN** tenant `acme` writes a basic memory with id `X`
- **THEN** the stored key is `tenant/acme/_memory/basic/X` and not
  `tenant/acme//_memory/basic/X`

### Requirement: Layering Is Preserved

The memory service SHALL depend only on the storage and tenant layers, and the
storage layer SHALL NOT import the memory-service, transport, or UI layers.

#### Scenario: Architecture conformance test passes

- **WHEN** the architecture conformance test runs
- **THEN** it reports no import from the storage engine into the memory-service,
  tenant, or user-interface layers
