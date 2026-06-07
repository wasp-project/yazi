# cost-metering-and-budgets Specification

## Purpose
TBD - created by archiving change cost-aware-memory-composition. Update Purpose after archive.
## Requirements
### Requirement: Per-Operation Usage Accounting

The system SHALL aggregate the `Usage` returned by each provider into a per-operation
total for every ingest and recall, capturing LLM input/output tokens, embedding
tokens, latency, and the downstream context tokens a recall injects. The pipeline
SHALL make this aggregated usage available to callers and to the metering subsystem.

#### Scenario: Ingest usage is the sum of its stages

- **WHEN** an ingest runs through extractor, embedder, and index providers
- **THEN** the reported usage equals the sum of each stage's usage

#### Scenario: Deterministic recall reports zero token usage

- **WHEN** a recall runs with deterministic providers only
- **THEN** the aggregated usage reports zero LLM and embedding tokens (latency may
  be non-zero)

### Requirement: Cost Aggregation per Tenant and Memory Class

The system SHALL roll operation usage up into running totals keyed by tenant and
by memory class (basic / advanced / policy), and SHALL convert token/embedding
usage to a monetary estimate using a configurable price table. These totals SHALL
be retrievable for reporting.

#### Scenario: Costs attributed to the right tenant

- **WHEN** tenant `acme` performs ingests and recalls
- **THEN** their token/embedding usage and estimated cost accrue under `acme` and
  not under any other tenant

#### Scenario: Re-pricing without re-running

- **WHEN** the configurable price table is changed
- **THEN** the monetary estimates recompute from the recorded usage without
  re-executing operations

### Requirement: Budget Enforcement

A profile SHALL be able to declare budgets — for example a maximum recall context
token count and a maximum ingest cost per unit of work. When an operation would
exceed a declared budget, the system SHALL enforce the budget by either rejecting
the operation with a clear error or degrading it to a cheaper path, according to
the configured enforcement mode. Absence of a budget SHALL mean unbounded.

#### Scenario: Recall context capped to budget

- **WHEN** a recall would inject more context tokens than the profile's
  `recall_context_tokens` budget
- **THEN** the system returns a cost-bounded subset within the budget rather than
  the full result set

#### Scenario: Over-budget ingest is rejected or degraded

- **WHEN** an ingest's estimated cost exceeds the profile's ingest budget
- **THEN** the system either rejects it with a clear error or runs a cheaper
  provider path, per the configured enforcement mode

#### Scenario: No budget means no limit

- **WHEN** a profile declares no budgets
- **THEN** operations run without budget-based rejection or degradation

