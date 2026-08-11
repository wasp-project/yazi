## Why

Yazi's roadmap is a sequence of opinionated engineering decisions — why storage
tiering beats a bigger cache, why BM25 beats embeddings for the free tier, why
embedding follows value rather than volume, why cost is metered before it is
spent. Those decisions currently live in `openspec/changes/*/design.md`, which is
the right home for *deciding* and the wrong home for *explaining*: design docs
are written before the work, in spec language, for the person implementing it.

For an open-source project whose thesis is contrarian ("compete on cost, not
capability"), the explanation **is** the distribution strategy. mem0, Zep and
MemOS all built their audience on published reasoning. Yazi has strong written
material already (`STATE-OF-ART.md` is a genuine comparison, not marketing), but
nothing that shows the work: what the code looked like before, what broke, what
was measured, what the numbers were afterwards.

This change establishes a technical blog written **as each milestone ships**,
not reconstructed later — while the measurements, the dead ends and the actual
diffs are still at hand. It also writes the first post now, covering how the
project reached its current state, so the series starts from a known baseline
rather than mid-stream.

## What Changes

- Add a `blog/` directory with a fixed post structure and a tracked index, plus a
  template that every post follows.
- Define the required content of a milestone post: **design** (the problem and
  the constraints), **principle** (the mechanism and why it works), **implementation**
  (what the code actually does, with file references), and **practice**
  (reproducible commands and measured numbers, including what did not improve).
- Require that every roadmap change ships its post as part of the change, not
  afterwards — the post is the last task in each of the four milestone changes.
- Require measured numbers to be reproducible: every figure names the command
  that produced it and the machine class it ran on; estimates are labelled as
  estimates, matching the honesty convention `benchmark/README.md` already sets.
- Write **post 000** now: the retrospective from KV server → memory model →
  layered architecture → cost-aware composition, with the verified current-state
  measurements and the honest list of what is not yet implemented.
- Add the blog to the docs index in `README.md`.

## Capabilities

### New Capabilities
- `engineering-blog`: the convention for milestone technical posts — location,
  structure, required sections, honesty rules for numbers, and the requirement
  that a post ships with its milestone.

### Modified Capabilities
<!-- None. This change adds a documentation convention and does not alter any
     existing system requirement. -->

## Impact

- **New files**: `blog/README.md` (index), `blog/TEMPLATE.md`,
  `blog/000-from-kv-store-to-agent-memory.md`.
- **Modified files**: `README.md` (link the blog), and the `tasks.md` of the four
  milestone changes, each of which already carries a final blog task.
- **Code**: none. No build, dependency or runtime impact.
- **Process**: a milestone is not complete until its post is written, which
  deliberately adds a small cost to shipping in exchange for the reasoning being
  captured while it is still accurate.
