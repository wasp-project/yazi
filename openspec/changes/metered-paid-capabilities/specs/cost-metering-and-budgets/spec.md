## ADDED Requirements

### Requirement: Per-Tenant Spend Ceiling

The system SHALL support a per-tenant maximum spend over a rolling period. When an
operation's projected cost would exceed the remaining allowance, the system SHALL
refuse the operation with a clear error, regardless of the configured
reject/degrade mode. A ceiling SHALL never be enforced by silently reducing
quality.

#### Scenario: Operation beyond the ceiling is refused

- **WHEN** a tenant has consumed its spend ceiling for the current period and
  issues an operation with non-zero projected cost
- **THEN** the operation is refused with an error naming the ceiling and the
  remaining allowance

#### Scenario: Ceiling refuses rather than degrades

- **WHEN** a tenant is over its ceiling and the enforcement mode is degrade
- **THEN** the operation is still refused rather than served on a cheaper path

#### Scenario: Free operations are unaffected by the ceiling

- **WHEN** a tenant is over its ceiling and issues a deterministic zero-cost
  operation
- **THEN** the operation succeeds

#### Scenario: Allowance resets with the period

- **WHEN** a new rolling period begins
- **THEN** the tenant's remaining allowance is restored

## MODIFIED Requirements

### Requirement: Budget Enforcement

A profile SHALL be able to declare budgets — for example a maximum recall context
token count and a maximum ingest cost per unit of work. Budget evaluation SHALL
use each costly provider's pre-call cost projection so the decision precedes the
spend. When an operation would exceed a declared budget, the system SHALL enforce
the budget by either rejecting the operation with a clear error or degrading it to
a cheaper path, according to the configured enforcement mode. Absence of a budget
SHALL mean unbounded. After execution the system SHALL record actual usage
alongside the projection so projection accuracy is measurable.

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

#### Scenario: Enforcement happens before the costly call

- **WHEN** an operation is rejected for exceeding its budget
- **THEN** no LLM or embedding request was issued for that operation

#### Scenario: Actual usage is reconciled against the projection

- **WHEN** a projected operation completes
- **THEN** both the projected and the actual usage are recorded
