#!/usr/bin/env python3
"""Benchmark entrypoint.

Examples
--------
  # dependency-free comparison (baseline + all cost estimators) + charts
  python3 run.py --adapters all --plot

  # only the cost estimators, just the preferences suite
  python3 run.py --adapters sim --suites preferences

  # include the real yazi adapter (needs a running server + yazictl on PATH)
  python3 run.py --adapters mock,yazi,sim-mem0,sim-zep --plot

  # re-render charts from a saved results file
  python3 run.py --from results/run.json --plot
"""

from __future__ import annotations

import argparse
import os
import sys

# make the benchmark/ dir importable regardless of cwd
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from adapters import registry
from diagrams import plot
from framework import report
from framework.metrics import Pricing
from framework.runner import RunConfig, run
from testcases.cases import load_suites


def main() -> int:
    ap = argparse.ArgumentParser(description="Continuous comparison of agent memory systems.")
    ap.add_argument("--adapters", default="all",
                    help="comma list of adapter names or selectors (all|sim|real). Default: all")
    ap.add_argument("--suites", default="all", help="comma list of suite names or 'all'")
    ap.add_argument("--pricing", default=None, help="path to a pricing JSON to override token prices")
    ap.add_argument("--out", default="results/run.json", help="where to write the JSON results")
    ap.add_argument("--plot", action="store_true", help="render SVG charts (and PNG if matplotlib present)")
    ap.add_argument("--png", action="store_true", help="also render PNG charts (requires matplotlib)")
    ap.add_argument("--from", dest="from_file", default=None, help="skip running; render report/charts from a results JSON")
    ap.add_argument("--quiet", action="store_true")
    args = ap.parse_args()

    here = os.path.dirname(os.path.abspath(__file__))
    out_path = args.out if os.path.isabs(args.out) else os.path.join(here, args.out)
    charts_dir = os.path.join(here, "results", "charts")
    pricing = Pricing.load(args.pricing)

    # Re-render mode: just charts/report from an existing file.
    if args.from_file:
        written = plot.render_from_file(args.from_file, charts_dir, want_png=args.png or args.plot)
        print(f"Charts written to {charts_dir}:")
        for p in written:
            print(f"  {os.path.relpath(p, here)}")
        return 0

    suite_names = [s.strip() for s in args.suites.split(",") if s.strip()]
    suites = load_suites(suite_names)
    if not suites:
        print(f"No suites matched {suite_names}. Available: see testcases/datasets/", file=sys.stderr)
        return 2

    names = registry.resolve([s.strip() for s in args.adapters.split(",") if s.strip()])
    adapters = []
    for n in names:
        try:
            adapters.append(registry.build(n))
        except KeyError as e:
            print(f"warning: {e}", file=sys.stderr)
    if not adapters:
        print("No valid adapters selected.", file=sys.stderr)
        return 2

    cfg = RunConfig(pricing=pricing, verbose=not args.quiet)
    results = run(adapters, suites, cfg)
    if not results:
        print("No adapters were available to run.", file=sys.stderr)
        return 1

    # write JSON
    os.makedirs(os.path.dirname(out_path), exist_ok=True)
    meta = {"suites": [s.name for s in suites], "adapters": [r.adapter for r in results]}
    with open(out_path, "w") as f:
        f.write(report.to_json(results, pricing, meta))
    # write CSV alongside
    csv_path = os.path.splitext(out_path)[0] + ".csv"
    with open(csv_path, "w") as f:
        f.write(report.to_csv(results, pricing))

    # markdown to stdout
    print("\n" + report.to_markdown(results, pricing) + "\n")
    print(f"Results: {os.path.relpath(out_path, here)}  (+ .csv)")

    if args.plot or args.png:
        written = plot.render(
            [r.to_dict(pricing) for r in results], charts_dir, want_png=args.png
        )
        print(f"Charts:  {os.path.relpath(charts_dir, here)}/  ({len(written)} files)")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
