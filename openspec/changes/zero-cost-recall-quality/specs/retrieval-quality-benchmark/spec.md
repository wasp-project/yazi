## ADDED Requirements

### Requirement: Labelled Retrieval Quality Suite

The benchmark SHALL include a labelled evaluation suite mapping queries to the
memory records that are relevant to them, kept separate from the existing
functional suites.

#### Scenario: Suite defines ground truth

- **WHEN** the quality suite is loaded
- **THEN** each query carries the set of memory ids considered relevant

### Requirement: Quality Metrics Reported Per Profile

The benchmark SHALL report precision@k, recall@k and mean reciprocal rank for
each evaluated profile, alongside the existing cost metrics, so quality and cost
are visible together rather than one being held constant.

#### Scenario: Free and paid profiles are compared on both axes

- **WHEN** the benchmark runs the deterministic and the embedding-based profiles
  over the quality suite
- **THEN** the report shows each profile's precision@k, recall@k, MRR, recurring
  cost and context tokens

#### Scenario: Estimated adapters remain flagged

- **WHEN** the report includes cost estimators for external systems
- **THEN** those rows remain marked as estimates and are not presented as measured
  quality

### Requirement: Ranking Regressions Are Detectable

The suite SHALL be runnable as a regression check that compares current metrics
against recorded baseline values and reports any decline.

#### Scenario: A ranking change that lowers quality is reported

- **WHEN** a ranking change reduces MRR below the recorded baseline
- **THEN** the regression check reports the decline with the affected queries

#### Scenario: Baselines are explicit

- **WHEN** the quality suite is run
- **THEN** the baseline values used for comparison are recorded in the output
