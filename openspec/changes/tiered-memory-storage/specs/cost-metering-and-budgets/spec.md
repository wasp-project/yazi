## ADDED Requirements

### Requirement: Storage Footprint Accounting

The system SHALL account for stored bytes and object/segment counts per storage
tier (hot, warm, cold) and per tenant, as a point-in-time footprint rather than a
per-operation flow. The footprint SHALL be retrievable for reporting.

#### Scenario: Footprint is reported per tier

- **WHEN** an operator requests the storage footprint
- **THEN** the report gives bytes and segment counts for the hot, warm and cold
  tiers

#### Scenario: Footprint is attributed per tenant

- **WHEN** two tenants hold memories in the same deployment
- **THEN** each tenant's stored bytes are reported separately

### Requirement: Storage Cost Estimation

The configurable price table SHALL include per-tier storage prices expressed per
gigabyte-month and a per-request price for object-storage operations. The system
SHALL convert the recorded footprint and request counts into a monthly cost
estimate, and SHALL recompute estimates when prices change without re-running any
operation.

#### Scenario: Monthly storage cost is estimated

- **WHEN** a footprint and a price table are available
- **THEN** the system reports an estimated monthly storage cost per tier and a
  total

#### Scenario: Object-storage requests are priced

- **WHEN** cold reads issue ranged object-storage requests
- **THEN** the request count is recorded and priced into the cost estimate

#### Scenario: Storage cost is separate from token cost

- **WHEN** a deployment runs the deterministic profile
- **THEN** the reported token cost is zero while the storage cost may be non-zero,
  and the two are reported as distinct figures

## MODIFIED Requirements

### Requirement: Cost Aggregation per Tenant and Memory Class

The system SHALL roll operation usage up into running totals keyed by tenant and
by memory class (basic / advanced / policy), and SHALL convert token/embedding
usage to a monetary estimate using a configurable price table. The system SHALL
additionally report, per tenant, the stored-byte footprint per storage tier and
its estimated monthly cost, keeping recurring token cost and at-rest storage cost
as separate figures rather than a single blended total. These totals SHALL be
retrievable for reporting.

#### Scenario: Costs attributed to the right tenant

- **WHEN** tenant `acme` performs ingests and recalls
- **THEN** their token/embedding usage and estimated cost accrue under `acme` and
  not under any other tenant

#### Scenario: Re-pricing without re-running

- **WHEN** the configurable price table is changed
- **THEN** the monetary estimates recompute from the recorded usage without
  re-executing operations

#### Scenario: Token cost and storage cost are reported separately

- **WHEN** a tenant's cost report is produced
- **THEN** it shows recurring token/embedding cost and at-rest storage cost as
  distinct line items, and does not sum stored bytes into per-operation usage
  totals
