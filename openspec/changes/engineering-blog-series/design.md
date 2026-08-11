## Context

The repository already carries an unusual amount of written reasoning:
`ARCHITECTURE.md` (layers, data flow, trade-offs), `STATE-OF-ART.md` (a
six-system comparison with an explicit cost model and a "Honest Assessment"
section), `COMPOSITION.md`, `MEMORY.md`, `LOCAL-DEPLOYMENT.md`, and OpenSpec
`design.md` files per change. The gap is not *documentation*; it is
**narrative with evidence**.

The distinction matters:

| Artifact | Written | Audience | Tense |
| --- | --- | --- | --- |
| `design.md` | before the work | the implementer | "we will" |
| `ARCHITECTURE.md` | after, continuously | an integrator | "it is" |
| a blog post | at ship time | someone deciding whether to care | "we tried, it broke, here are the numbers" |

Nothing currently records the third. And the third is the one that carries a
contrarian thesis: "cost is an unserved axis" is a claim that needs demonstrated
work behind it, not a table.

Two facts constrain the design. First, the project's credibility rests on its
honesty conventions — `benchmark/README.md` flags estimated numbers with `*` and
refuses to fabricate adapter results; `STATE-OF-ART.md` states plainly what Yazi
cannot do. A blog that oversells would cost more than it earns. Second, posts
written months after the fact lose exactly what makes them valuable: the failed
approach, the surprising measurement, the specific line of code that was wrong.

## Goals / Non-Goals

**Goals:**
- One post per shipped milestone, written while the work is fresh.
- Every post carries reproducible measurements, including negative results.
- A structure consistent enough that a reader knows where to find the mechanism
  and where to find the numbers.
- Zero build or dependency impact — plain Markdown in the repository.

**Non-Goals:**
- A website, static-site generator, CI publishing pipeline, or comment system.
  Markdown in `blog/` renders on any forge; a generator can be added later
  without rewriting posts.
- A posting cadence. Posts are event-driven (a milestone ships), not scheduled;
  a calendar would force filler.
- Translation policy. Posts are written in one primary language; translations
  are welcome as sibling files and are not required.
- Marketing content, release announcements, or roadmap teasers. This series is
  about work that is already merged.

## Decisions

### Decision 1: `blog/NNN-slug.md`, numbered and immutable in ordering

Posts are `blog/000-from-kv-store-to-agent-memory.md`,
`blog/001-...`, and so on. `blog/README.md` is the index: number, title, date,
the change it corresponds to, and a one-line hook.

- **Why numbers**: they encode reading order, which matters because the series is
  cumulative — the tiering post assumes the reader knows why tiering was absent.
- **Why not dates in filenames**: dates go in front matter; a renumbered or
  delayed post should not require renaming links.

### Decision 2: Four required sections — design, principle, implementation, practice

Every post carries, in order:

1. **Design** — the problem in the code as it stood, with file and line
   references, and the constraints that ruled options out. Concrete, not abstract.
2. **Principle** — the mechanism chosen and why it works, at the level of "sparse
   block index decouples resident memory from key count", including the
   alternatives rejected and the reason.
3. **Implementation** — what was actually built: the interfaces, the key layouts,
   the file formats, the seams. Points at the real code.
4. **Practice** — reproducible commands, measured before/after numbers, and what
   the reader should do differently. Includes what did *not* improve.

- **Why this shape**: it maps onto the four questions a technical reader asks in
  order — what was wrong, what is the idea, how is it built, does it work. It
  also maps cleanly onto material the OpenSpec change already contains, so a post
  is a rewrite of known facts for a different audience rather than new research.
- **Why "practice" is mandatory and last**: a post about a cost-aware system that
  reports no numbers would undercut the project's central claim. Putting it last
  forces the measurement to exist before the post can be finished.

### Decision 3: Numbers must be reproducible or labelled as estimates

Every figure states the command that produced it, the corpus size, and the
machine class. Vendor-reported or modelled numbers are labelled inline, matching
the `*` convention in `benchmark/README.md`. Negative and null results are
reported rather than omitted.

- **Why**: this is the project's existing standard applied to its public writing.
  A benchmark harness that refuses to fabricate adapter results, paired with a
  blog that quotes unreproducible wins, would be incoherent.

### Decision 4: The post is a task inside its milestone change

Each of `agent-integration-surface`, `tiered-memory-storage`,
`zero-cost-recall-quality` and `metered-paid-capabilities` ends with a blog task.
The change is not complete until it is checked off.

- **Why**: the alternative — a backlog of "write the post" items — reliably
  produces zero posts. Coupling it to the change makes the cost visible and the
  omission obvious.
- **Trade-off acknowledged**: this adds real time to shipping each milestone. It
  is accepted deliberately; for a project competing on a thesis rather than on
  features, the explanation is part of the product.

### Decision 5: Post 000 is a retrospective, written from verified state

The first post covers the path already travelled — KV server → memory model →
four-layer architecture → cost-aware composition — and states the current honest
position, including the gaps `QUICKSTART.md` documents. Its numbers come from
running the current build, not from the docs.

- **Why start with a retrospective**: the series needs a baseline. A reader
  arriving at post 001 ("we made storage tiered") needs to know what existed
  before, and the answer — "the tiering was documented but not implemented" — is
  itself the most interesting thing about the starting point.
- **Why it must be honest about the gaps**: `STATE-OF-ART.md` already sets that
  tone. A first post that claimed a finished tiered system would be contradicted
  by the second post.

## Risks / Trade-offs

- **[Posts become stale as the code moves]** → Posts are dated and describe a
  point in time; they are never edited to match later code. Corrections are
  appended as notes with dates. The living description stays in `ARCHITECTURE.md`.
- **[Writing slows shipping]** → Bounded by scope: a post is a rewrite of the
  change's own design and measurements, not new work. If a milestone slips
  because of its post, that is a signal the measurements were never taken.
- **[Documentation drifts across five files]** → `blog/README.md` is the only
  index; posts link to `ARCHITECTURE.md` for current state rather than restating
  it.
- **[Temptation to oversell]** → The mandatory "what did not improve" content in
  the practice section, and the estimate-labelling rule, exist specifically to
  make an overselling post fail review.

## Open Questions

- Primary language: the project's existing docs are in English while the
  maintainer works in Chinese. Leaning English-primary with optional
  `-zh.md` siblings, since the audience being persuaded is the international
  open-source one — to be confirmed by the maintainer.
- Should posts be published anywhere beyond the repository (a forge Pages site, a
  developer platform)? Out of scope here; the Markdown is portable to any of
  them.
