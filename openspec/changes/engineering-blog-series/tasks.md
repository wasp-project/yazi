## 1. Blog Scaffolding

- [x] 1.1 Create `blog/` with `README.md` as the index: a table of number, title, date, corresponding change, and one-line summary
- [x] 1.2 Create `blog/TEMPLATE.md` with front matter (title, date, change, author) and the four required sections — design, principle, implementation, practice — each with a short prompt describing what belongs there
- [x] 1.3 Record the honesty rules in the template: every measured figure names its command, corpus size and machine class; estimates are labelled; results that did not improve are reported
- [x] 1.4 Confirm the primary language and whether translated siblings are wanted; record the answer in `blog/README.md` — English primary, optional `NNN-slug-zh.md` siblings

## 2. Post 000 — Baseline Retrospective

- [x] 2.1 Gather the baseline measurements from the current build: build and test times, memory write/read/recall latency, `lite` and `standard` recall usage and cost, resident memory at a known corpus size, and the on-disk layout after a restart
- [x] 2.2 Write the design section: the path from a lightweight KV server to a structured memory model, and what problem each step solved
- [x] 2.3 Write the principle section: why cost was chosen as the competitive axis instead of capability, summarizing the six-system comparison and the six cost points
- [x] 2.4 Write the implementation section: the four layers, the `/_memory/...` key and index layout, the tenant prefix seam, the LSM/S3 backends, and the provider pipeline — with file references
- [x] 2.5 Write the practice section: the verified quick-start command sequence with real output, plus the honest gap list (documented-but-unimplemented tiering, client-side memory model, no HTTP/MCP surface, per-recall re-embedding, naive ranking)
- [x] 2.6 Add post 000 to the index and link the blog from `README.md`

## 3. Wire The Series Into The Roadmap

- [x] 3.1 Verify each of the four milestone changes ends with a blog task, and that each task names the specific content that post must carry
- [x] 3.2 Add a line to the project contribution notes stating that a milestone change is not complete until its post exists — recorded as the `context:` block in `openspec/config.yaml`, which is shown to every agent creating change artifacts
- [x] 3.3 Add the point-in-time rule to `blog/README.md`: published posts are not rewritten; corrections are appended as dated notes and the current state lives in `ARCHITECTURE.md`
