---
title: <Post title — say what changed, not what it enables>
date: <YYYY-MM-DD>
change: <openspec change name, e.g. tiered-memory-storage>
author: <name>
---

# <Post title>

<One paragraph. What was broken, what you did, and the single number that shows
it worked. A reader who stops here should still have learned something true.>

---

## 1. Design — what was actually wrong

<The problem as it existed in the code, not in the abstract. Name files and
lines: `pkg/storage/lsm/sstable.go:76-125`. Quote the offending code if it is
short. State how you noticed — a measurement, a user report, a review.

Then the constraints that ruled options out. For this project they usually
include: single static binary, no new module dependency, the `lite` profile stays
at zero tokens, the on-disk key layout is fixed.

If the problem was documented-but-unimplemented behavior, say so plainly. That is
the more interesting failure and hiding it fools nobody who reads the code.>

## 2. Principle — the mechanism, and what we rejected

<The idea, at the level of "a sparse block index decouples resident memory from
key count". Enough theory that a reader could implement it differently and still
get it right.

Then at least one alternative, with the real reason it lost. "It was more
complex" is not a reason; "it makes resident memory unmeasurable, which defeats
the point of the change" is.

If the mechanism is standard (BM25, LSM compaction, SigV4), say so and cite it.
Novelty is not the goal.>

## 3. Implementation — what the code does now

<The seams: interfaces added, key layouts, file formats, where the decision
points live. Point at real code so a reader can follow.

Include the things that were harder than expected. Include the migration if there
was one — how old data is detected, converted, and what happens if conversion is
declined.>

## 4. Practice — the numbers, and how to get them yourself

<Machine class first, so every figure below is interpretable:>

> Measured on <CPU>, <RAM>, <OS>, <Go version>. Corpus: <N records, description>.
> Configuration: <engine, relevant settings>.

<Then the commands, verbatim and runnable:>

```bash
<command>
```

<Then a before/after table. Include the metrics that did not move, and say so.>

| Metric | Before | After |
| --- | --- | --- |
|  |  |  |

<Then: what a reader should do differently now — a config value to set, a
default that changed, a limitation to plan around.>

<Then, honestly: what this did not fix, and which post will.>

---

## Checklist before publishing

- [ ] Every number states its command, corpus size and machine class, or is
      labelled an estimate
- [ ] At least one rejected alternative, with a real reason
- [ ] At least one result that did not improve, or an explicit statement that
      everything measured improved
- [ ] File references point at code that exists on the commit being described
- [ ] Added to the table in `blog/README.md`
- [ ] Linked from the milestone's `tasks.md`
