# Yazi Memory Benchmark

A **continuous comparison** harness for agent memory systems, built to make
Yazi's [cost-aware](../STATE-OF-ART.md) thesis measurable: hold retrieval quality
roughly constant and compare the **recurring cost** of producing and retrieving
memory across systems (mem0, Zep, Letta, MemOS, mem9, … vs. Yazi).

It runs **with zero external dependencies** out of the box (Python 3.10+ stdlib
only): a cheap baseline plus per-system cost *estimators* produce a full report
and charts. Plug in real adapters (SDKs + keys) when you want ground truth.

```
$ python3 run.py --adapters all --plot
```

| Adapter | Acc. | Recall ms | Sys cost $ | $/1k ops | Ctx tok | Footprint |
| --- | --- | --- | --- | --- | --- | --- |
| mock | 100% | 0.1 | 0 | 0 | 354 | in-process |
| yazi | 100% | ~19 | 0 | 0 | 65 | single Go binary |
| sim-mem0 * | 100% | 60 | 0.000129 | 0.0052 | 354 | vector + graph DB |
| sim-letta * | 100% | 900 | 0.000723 | 0.0289 | 354 | server + Postgres |
| sim-memos * | 100% | 40 | 0.000129 | 0.0052 | **176** | graph + vector + KV-cache |

*(`*` = estimate; numbers above are illustrative. Accuracy is held constant by
design — see Methodology.)*

---

## What it measures

The cost model mirrors [STATE-OF-ART.md](../STATE-OF-ART.md) §1:

| Metric | Meaning | Cost point |
| --- | --- | --- |
| **accuracy** | fraction of queries whose top-k retrieval contained the ground truth | quality |
| **avg recall / ingest latency** | wall-clock per op | perf |
| **system cost ($)** | recurring spend the memory system itself incurs: extraction LLM + embeddings + recall-side LLM | C1+C2+C4-internal |
| **$/1k ops**, **$/correct** | normalized cost | — |
| **context tokens** | tokens of recalled context injected downstream (paid by *every* system) | C4-prompt |
| **footprint** | operational weight (single binary → DB+queue+coordinators) | C6 |

> **Key idea:** `system cost` is what scales **per message, forever**. A
> deterministic store (yazi/mock) spends $0 there; LLM-based systems do not. That
> gap — not accuracy on a toy suite — is the comparison.

## Methodology & honesty

- **Accuracy is deliberately held ~constant.** All dependency-free adapters share
  the same keyword retriever (`adapters/keyword_store.py`), so the report
  isolates the **cost axis**. Real adapters use their own engines and *will*
  differ on accuracy — run them for that.
- **`sim-*` adapters are estimators, not measurements.** They attach a documented
  per-system cost profile (`adapters/simulator.py`) — extraction multipliers,
  embedding on/off, recall-LLM, KV-cache reuse — grounded in each system's
  qualitative cost description. They are clearly flagged with `*` / `is_estimate`.
- **`yazi` and `mock` are real**, with truly zero token cost.
- **Real external adapters** (`adapters/external.py`) are import-guarded stubs;
  they report `available = False` until you wire up the SDK + backend + keys, so
  the harness never fabricates their numbers.

## Layout

```
benchmark/
  run.py                  CLI entrypoint
  framework/
    adapter.py            MemoryAdapter interface + data types
    metrics.py            cost model (Pricing, OpMetrics, AdapterResult)
    runner.py             drives adapters over suites, aggregates
    report.py             JSON / Markdown / CSV output
  adapters/
    keyword_store.py      shared retriever (keeps accuracy constant offline)
    mock_adapter.py       zero-cost baseline (yazi's class)
    yazi_adapter.py       REAL: drives a running yazi via yazictl
    simulator.py          sim-mem0/zep/letta/memos/mem9 cost estimators
    external.py           REAL mem0/zep/letta/memos/mem9/milvus (opt-in stubs)
    registry.py           name -> adapter resolution
  testcases/
    cases.py              suite loader
    datasets/*.json       test suites (preferences, multi_session, …)
  evaluation/
    evaluators.py         retrieval scoring (id / substring hit @k)
  diagrams/
    plot.py               dependency-free SVG charts (+ optional matplotlib PNG)
  results/                run outputs (gitignored)
```

## Usage

```bash
# everything that runs without setup (baseline + estimators) + charts
python3 run.py --adapters all --plot

# only estimators, one suite
python3 run.py --adapters sim --suites preferences

# include the REAL yazi adapter:
#   1) start a server:  (cd ../ && cp config/local.yml config/default.yml && go run ./cmd/yazi)
#   2) build the CLI:   (cd ../ && CGO_ENABLED=0 go build -o /tmp/yazictl ./cmd/cli)
#   3) point at it and run:
YAZICTL=/tmp/yazictl python3 run.py --adapters mock,yazi,sim-mem0 --plot

# re-price the whole comparison without re-running (token prices change)
python3 run.py --adapters all --pricing my_pricing.json

# re-render charts from a saved result
python3 run.py --from results/run.json --plot
```

Outputs: `results/run.json`, `results/run.csv`, and `results/charts/*.svg`
(`*.png` too with `--png` and matplotlib).

## Adding a real adapter

1. Implement `ingest()` and `recall()` in `adapters/external.py`, filling
   `OpMetrics` with **real** token counts from the SDK responses.
2. Make `_configured()` return True only when the SDK + backend + keys are present.
3. Run `python3 run.py --adapters mock,yazi,<name>`.

## Adding a test suite

Drop a JSON file in `testcases/datasets/` (see `preferences.json` for the shape:
`items` to store, `queries` with `expected_ids` / `expected_substring`). It is
picked up automatically.

## Continuous use (CI)

The dependency-free path is deterministic and fast — suitable for CI to track the
cost gap over time:

```bash
python3 benchmark/run.py --adapters all --out benchmark/results/run.json --plot
# publish results/run.json + results/charts/*.svg as build artifacts
```

Tune token prices in a pricing JSON and per-system multipliers in
`adapters/simulator.py` as the landscape changes; swap in real adapters for
periodic ground-truth runs.
