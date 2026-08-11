## ADDED Requirements

### Requirement: Inverted Index Over Memory Text

The system SHALL maintain an inverted index mapping terms to the memories
containing them, together with the corpus statistics needed for length-normalized
scoring. The index SHALL be updated when a memory is written, updated or deleted,
and SHALL be stored in the same key space as existing secondary indexes so it
inherits tenant namespacing and durability.

#### Scenario: Writing a memory indexes its terms

- **WHEN** a memory is written
- **THEN** each of its distinct indexable terms has a postings entry referring to
  that memory

#### Scenario: Deleting a memory removes it from the index

- **WHEN** a memory is deleted
- **THEN** no postings entry refers to it and it cannot be returned by recall

#### Scenario: Index is tenant-scoped

- **WHEN** two tenants write memories containing the same term
- **THEN** each tenant's recall consults only its own postings and cannot observe
  the other tenant's documents or corpus statistics

### Requirement: Query-Driven Candidate Selection

Recall SHALL select candidates by looking up the query's terms in the inverted
index, and SHALL NOT rely on scanning a fixed-size prefix of the corpus before
scoring.

#### Scenario: A relevant memory beyond the old scan window is found

- **WHEN** the corpus contains far more memories than the previous scan limit and
  the only relevant memory was written last
- **THEN** recall returns that memory

#### Scenario: Non-matching memories are never scored

- **WHEN** a query term appears in only a few memories
- **THEN** only those memories are scored

### Requirement: BM25 Relevance Scoring With Field Weights

Textual relevance SHALL be computed with BM25 using configurable `k1` and `b`
parameters, over fields weighted so that a memory's subject and tags contribute
more than its body text. Term frequency saturation and document length
normalization SHALL both apply.

#### Scenario: Rare terms outweigh common ones

- **WHEN** a query contains both a term occurring in nearly every memory and a
  term occurring in one memory
- **THEN** the memory containing the rare term ranks above memories matching only
  the common term

#### Scenario: Stopword-only overlap does not produce a hit

- **WHEN** a query and a memory share only stopwords
- **THEN** that memory is not returned

#### Scenario: Subject match outranks equal body match

- **WHEN** one memory matches a query term in its subject and another matches the
  same term only in its body, all else equal
- **THEN** the subject match ranks higher

#### Scenario: Long memories do not dominate by length

- **WHEN** a long memory repeats a query term many times and a short memory
  contains it once with no other distinguishing signal
- **THEN** the long memory's score is saturated and length-normalized rather than
  scaling linearly with term count

### Requirement: Explicit Tokenization And Stopwords

Tokenization SHALL be Unicode-aware and case-insensitive, and a configurable
stopword list SHALL be applied. When a query consists entirely of stopwords, the
system SHALL fall back to the unfiltered query terms rather than returning no
results.

#### Scenario: Stopword list is configurable

- **WHEN** an operator supplies a custom stopword list
- **THEN** indexing and querying both use that list

#### Scenario: All-stopword query still returns results

- **WHEN** a query contains only stopwords
- **THEN** the system scores against the unfiltered terms and returns its best
  matches rather than an empty result

### Requirement: Recall Remains Free Of Token Cost

Ranking SHALL introduce no LLM or embedding calls. A recall served by the
deterministic profile SHALL report zero LLM and embedding tokens.

#### Scenario: Deterministic recall reports zero token cost

- **WHEN** a recall runs on the deterministic profile with the new ranker
- **THEN** the reported usage shows zero LLM and embedding tokens and the
  estimated token cost is zero

### Requirement: Index Rebuild For Existing Data

The system SHALL rebuild the inverted index from stored records when the index is
absent or of an older version. The rebuild SHALL be resumable and SHALL keep the
store readable while it runs.

#### Scenario: Existing corpus becomes searchable after upgrade

- **WHEN** a store written before this change is opened
- **THEN** the index is rebuilt from the existing records and all of them become
  reachable by term-driven recall

#### Scenario: Interrupted rebuild resumes

- **WHEN** a rebuild is interrupted and the process restarts
- **THEN** the rebuild continues from its recorded progress rather than starting
  over
