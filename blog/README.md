# Yazi Engineering Blog

Technical posts written **as each milestone ships** — while the measurements, the
dead ends, and the actual diffs are still at hand.

This is not release notes and not documentation. Documentation tells you what the
system is ([ARCHITECTURE.md](../ARCHITECTURE.md)); design docs
([openspec/changes](../openspec/changes)) tell you what we intend to build. These
posts tell you what was wrong, what we chose, how it was built, and what the
numbers said afterwards — including when they said nothing improved.

## Posts

| # | Title | Date | Change | In one line |
| --- | --- | --- | --- | --- |
| [000](./000-from-kv-store-to-agent-memory.md) | From KV Store to Agent Memory | 2026-08-10 | *(baseline)* | How a 3,000-line KV server became a cost-aware agent memory engine — and the five things it still only claims to do. |

## Conventions

**Language.** Posts are written in English, matching the rest of the repository's
documentation. Chinese translations are welcome as `NNN-slug-zh.md` siblings and
are never required for a post to ship.

**Structure.** Every milestone post has four sections, in order:

1. **Design** — the concrete problem in the code as it stood, naming the affected
   files. Not an abstract motivation.
2. **Principle** — the mechanism chosen, why it works, and at least one
   alternative rejected with the reason.
3. **Implementation** — what was actually built: interfaces, key layouts, file
   formats, seams. Points at real code.
4. **Practice** — reproducible commands and measured results, including what did
   **not** improve.

Start from [TEMPLATE.md](./TEMPLATE.md).

**Numbers.** Every quantitative claim either states the command, corpus size and
machine class that produced it, or is explicitly labelled an estimate. This is the
same standard [`benchmark/README.md`](../benchmark/README.md) applies to adapter
results: estimates are marked, and nothing is fabricated. Results that got worse
or stayed flat are reported alongside the ones that improved.

**A post ships with its milestone.** Each roadmap change carries writing its post
as its final task, and the change is not complete until the post exists. This
deliberately costs time at ship. For a project whose thesis is contrarian, the
explanation is part of the product.

**Posts are point-in-time records.** A published post is never rewritten to match
later code. When subsequent work supersedes it, a dated note is appended pointing
at the newer post. The living description of the system stays in
[ARCHITECTURE.md](../ARCHITECTURE.md).

## The roadmap these posts will cover

| Milestone | Change | Post |
| --- | --- | --- |
| M0 · Make it integrable | `agent-integration-surface` | pending |
| M1 · Make the tiering real | `tiered-memory-storage` | pending |
| M2 · Make the free tier good | `zero-cost-recall-quality` | pending |
| M3 · Make the paid tier honest | `metered-paid-capabilities` | pending |
