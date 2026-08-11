## ADDED Requirements

### Requirement: Durable Cost Counters

Cost counters SHALL be persisted in the store, namespaced per tenant, and SHALL
survive process restarts, so cumulative spend is answerable rather than
per-process.

#### Scenario: Totals accumulate across restarts

- **WHEN** metered operations run, the server restarts, and more operations run
- **THEN** the reported totals include operations from before the restart

#### Scenario: Counters are tenant-scoped

- **WHEN** two tenants perform metered operations
- **THEN** each tenant's totals reflect only its own operations

### Requirement: Cost Report Breakdown

The system SHALL expose a cost report broken down by tenant, memory class and
provider, reporting token usage, estimated token cost, and at-rest storage cost as
separate figures over a requested period.

#### Scenario: Report separates token and storage cost

- **WHEN** a cost report is requested
- **THEN** recurring token cost and at-rest storage cost appear as distinct line
  items rather than a single blended total

#### Scenario: Report attributes cost to providers

- **WHEN** a deployment runs an embedder and an LLM reranker
- **THEN** the report shows each provider's share of the token cost

#### Scenario: Free profile reports zero token cost

- **WHEN** a deployment runs only the deterministic profile
- **THEN** the report shows zero token cost and may show non-zero storage cost

### Requirement: Cost Report Is Reachable From Every Surface

The cost report SHALL be available through the memory service, the HTTP API and a
CLI command.

#### Scenario: CLI prints the cost report

- **WHEN** the operator runs the cost command
- **THEN** the report is printed for the requested tenant and period

#### Scenario: API returns the cost report

- **WHEN** a client requests the cost report over HTTP
- **THEN** the report is returned as JSON with the same breakdown as the CLI

### Requirement: Projection Accuracy Is Measurable

For operations whose cost was projected before execution, the system SHALL record
the projected and the actual usage so systematic projection bias is visible.

#### Scenario: Projection error is reported

- **WHEN** metered operations with pre-call projections have run
- **THEN** the report includes the aggregate difference between projected and
  actual usage
