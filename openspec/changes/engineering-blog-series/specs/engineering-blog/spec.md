## ADDED Requirements

### Requirement: Blog Location And Index

The repository SHALL contain a `blog/` directory holding numbered Markdown posts
named `NNN-slug.md`, and an index listing every post with its number, title,
date, the change it documents, and a one-line summary. The index SHALL be updated
whenever a post is added.

#### Scenario: A new post is discoverable

- **WHEN** a post is added to the blog directory
- **THEN** the index lists it with its number, title, date, corresponding change
  and summary

#### Scenario: Reading order is explicit

- **WHEN** a reader opens the index
- **THEN** posts are ordered by their number, which reflects the order in which
  the work shipped

### Requirement: Required Post Structure

Every milestone post SHALL contain, in order, a design section, a principle
section, an implementation section and a practice section. The design section
SHALL describe the concrete problem in the code as it stood, with references to
the affected files. The principle section SHALL explain the chosen mechanism and
the alternatives rejected. The implementation section SHALL describe what was
built and point at the real code. The practice section SHALL give reproducible
commands and measured results.

#### Scenario: Post follows the required structure

- **WHEN** a milestone post is written
- **THEN** it contains the design, principle, implementation and practice
  sections in that order

#### Scenario: Design section is concrete

- **WHEN** a post describes the problem it solved
- **THEN** it names the affected files or components rather than describing the
  problem only in the abstract

#### Scenario: Alternatives are recorded

- **WHEN** a post explains a chosen mechanism
- **THEN** it states at least one alternative that was considered and why it was
  rejected

### Requirement: Measurements Are Reproducible Or Labelled

Every quantitative claim in a post SHALL either state the command, corpus size
and machine class that produced it, or be explicitly labelled as an estimate or a
vendor-reported figure. Posts SHALL report results that did not improve alongside
those that did.

#### Scenario: A measured figure names its command

- **WHEN** a post reports a performance or cost number obtained by running the
  system
- **THEN** it states the command and conditions under which it was obtained

#### Scenario: An estimate is labelled

- **WHEN** a post cites a modelled or vendor-reported figure
- **THEN** the figure is marked as an estimate and its source is named

#### Scenario: Negative results are included

- **WHEN** a milestone produced a change that did not improve a measured metric
- **THEN** the post reports it rather than omitting it

### Requirement: A Post Ships With Its Milestone

Each roadmap change SHALL include writing its blog post as a task, and the change
SHALL NOT be considered complete until the post exists.

#### Scenario: Milestone completion requires its post

- **WHEN** every implementation task of a milestone change is complete but its
  blog post has not been written
- **THEN** the change is not complete

#### Scenario: Post corresponds to a change

- **WHEN** a milestone post is published
- **THEN** the index records which change it documents

### Requirement: Posts Are Point-In-Time Records

A published post SHALL NOT be rewritten to reflect later changes to the code.
Corrections and superseding information SHALL be added as dated notes, and the
current description of the system SHALL remain in the architecture documentation.

#### Scenario: Later change does not rewrite an earlier post

- **WHEN** subsequent work supersedes what a post described
- **THEN** the post retains its original content and carries a dated note
  pointing at the newer post or the current architecture document

### Requirement: Series Baseline Post

The series SHALL begin with a retrospective post establishing the project's state
before the roadmap work, describing how the system reached that state and stating
honestly what was not yet implemented.

#### Scenario: Baseline post states the starting position

- **WHEN** a reader opens the first post
- **THEN** it describes the path from the original key-value server to the
  current memory system and lists the capabilities that were documented but not
  yet implemented

#### Scenario: Baseline numbers come from the running system

- **WHEN** the baseline post reports the system's behavior
- **THEN** the figures were obtained by running the current build, not copied
  from prior documentation
