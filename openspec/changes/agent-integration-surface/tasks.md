## 1. Configuration & Startup

- [ ] 1.1 Add `--config` flag and `YAZI_CONFIG` env support to `cmd/yazi`, resolving in the order flag → env → `./config/default.yml`; log the resolved path
- [ ] 1.2 Split `config.Load` into "load optional default" and "load required path": a missing/unparseable file at an explicitly requested path returns an error; a missing file at the default path logs INFO and keeps built-in defaults
- [ ] 1.3 Make `cmd/yazi` exit non-zero on fatal startup errors instead of returning silently from `Run()`
- [ ] 1.4 Change `.gitignore` from `config/**` to `config/default.yml`; `git add` `config/local.yml` and `config/cloud.yml` as tracked examples
- [ ] 1.5 Test: zero-config start succeeds; `--config /nonexistent.yml` exits non-zero; malformed YAML exits non-zero

## 2. Server-Side Memory Service

- [ ] 2.1 Create `pkg/memory/service` with a `Memory` service holding a `storage.KVStore` and a `tenant.Service`, constructing a tenant-scoped `*memory.Store` per request
- [ ] 2.2 Implement service methods for all three classes: `PutBasic/GetBasic/ListBasic/DeleteBasic` and the advanced/policy equivalents, each taking an explicit tenant argument
- [ ] 2.3 Implement `Recall(tenant, query, profileCfg)` on the service, reusing `memory.RecallWithProfile`, returning hits + `provider.Usage` + estimated cost
- [ ] 2.4 Fix the tenant prefix join so namespaced keys are `tenant/<id>/_memory/...` with no doubled separator; add a regression test
- [ ] 2.5 Wire the service into `pkg/server/server.go`, replacing the unused `s.tenant` placeholder with a live construction over the selected store
- [ ] 2.6 Test: service-produced keys are byte-identical to the pre-change client-side layout (bypass) and correctly namespaced (tenant); two tenants cannot see each other's records
- [ ] 2.7 Extend `pkg/arch` conformance test to assert the storage layer does not import the new memory-service or transport packages

## 3. HTTP Memory API

- [ ] 3.1 Add an `http:` config block (address, port, enabled) defaulting to `127.0.0.1:3457`, enabled
- [ ] 3.2 Create `pkg/server/http` with a stdlib `net/http` mux: `POST/GET/DELETE /v1/memory/{basic|advanced|policy}[/{id}]`, `POST /v1/memory/{class}:list`, `POST /v1/memory:recall`, `GET /healthz`
- [ ] 3.3 Resolve the tenant from the `X-Yazi-Tenant` header through the tenant service, falling back to the configured tenant
- [ ] 3.4 Implement the error contract: 400 malformed, 404 missing id, 422 budget violation, 503 unavailable provider — all with a JSON `{"error": "..."}` body
- [ ] 3.5 Start the HTTP listener alongside the existing protocol listener in `Server.Run`, and shut it down cleanly
- [ ] 3.6 Test with `httptest`: write→read→list→delete round-trip; recall returns hits + usage + `costUSD`; `lite` recall reports zero cost; tenant header isolates records; each error code path

## 4. MCP Server

- [ ] 4.1 Create `pkg/mcp` implementing JSON-RPC 2.0 framing over stdin/stdout, with all logging redirected to stderr
- [ ] 4.2 Implement `initialize` and `tools/list` returning exactly four tools: `remember`, `recall`, `list_memories`, `forget`, with compact descriptions and JSON input schemas
- [ ] 4.3 Implement `tools/call` mapping each tool onto `service.Memory`; `recall` returns memory texts plus the usage/cost block
- [ ] 4.4 Add a `yazi mcp` subcommand that serves MCP over stdio against the configured storage (no network port)
- [ ] 4.5 Test: initialize handshake; tool list is exactly four entries; a remember→recall→forget sequence over the JSON-RPC transport; stdout contains no non-protocol bytes
- [ ] 4.6 Document the MCP client configuration snippet in `QUICKSTART.md` and `LOCAL-DEPLOYMENT.md`

## 5. CLI Cleanup

- [ ] 5.1 Convert every `cmd/cli` command body from `Run` + `panic` to `RunE` returning errors; set `SilenceUsage` and `SilenceErrors` and print `Error: <msg>` to stderr with exit code 1
- [ ] 5.2 Route CLI memory subcommands through the server-side service (via gRPC) instead of `newMemoryStore`; delete `newMemoryStore` and `memoryClientKV`
- [ ] 5.3 Ensure successful commands write only their payload to stdout so the benchmark adapter and shell helpers keep working
- [ ] 5.4 Test: unreachable server, malformed JSON, unknown id, and `--profile pro` each produce a one-line error and a non-zero exit

## 6. Cleanup & Release

- [ ] 6.1 Delete `pkg/storage/memory_store.go` and `pkg/storage/memory_store_test.go` (duplicate, unreferenced memory model)
- [ ] 6.2 Fix the `Makefile` `build` target: correct the `LOCLABIN` typo, set `CGO_ENABLED=0`, build all three binaries from their packages
- [ ] 6.3 Add a multi-stage `Dockerfile` producing a static single-binary image with a default loopback HTTP + gRPC configuration
- [ ] 6.4 Add a GitHub Actions release workflow building static binaries for linux/darwin × amd64/arm64 and attaching them to a tag
- [ ] 6.5 Update `README.md` and `QUICKSTART.md`: replace the inline config heredoc with `cp config/local.yml config/default.yml`, document the HTTP API and MCP setup, and refresh the Known-limits list
- [ ] 6.6 Run `CGO_ENABLED=0 go build ./...` and `CGO_ENABLED=0 go test ./...`; confirm the full suite passes
- [ ] 6.7 Verify the benchmark still runs end-to-end: start the server, run `python3 run.py --adapters mock,yazi`, confirm the `yazi` adapter reports real numbers instead of SKIPPED

## 7. Engineering Blog

- [ ] 7.1 Write the milestone blog post for this change under `blog/` following the `engineering-blog` capability: the client-side-memory problem, why the service seam goes where it does, the HTTP/MCP surface, and measured before/after round-trip counts for a recall
