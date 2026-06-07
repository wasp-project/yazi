"""Diagram tools — render benchmark results to charts.

Dependency-free by default: emits standalone SVG files using only the stdlib, so
charts render anywhere (browser, README, CI artifact). If matplotlib is
installed, `--png` additionally writes PNGs.

Charts produced:
  - cost_bar.svg       memory-system $ cost per adapter (log-friendly)
  - latency_bar.svg    average recall latency per adapter
  - accuracy_cost.svg  accuracy (x) vs cost (y) — the cost/quality frontier
"""

from __future__ import annotations

import html
import json
import os

_PALETTE = ["#2563eb", "#16a34a", "#dc2626", "#d97706", "#7c3aed", "#0891b2", "#db2777", "#65a30d"]


def _esc(s: str) -> str:
    return html.escape(str(s))


def _svg_header(w: int, h: int) -> str:
    return (
        f'<svg xmlns="http://www.w3.org/2000/svg" width="{w}" height="{h}" '
        f'viewBox="0 0 {w} {h}" font-family="system-ui,Arial,sans-serif">'
        f'<rect width="{w}" height="{h}" fill="white"/>'
    )


def _bar_chart(title: str, labels: list[str], values: list[float], unit: str) -> str:
    w, h = 720, 380
    pad_l, pad_r, pad_t, pad_b = 60, 20, 50, 90
    plot_w = w - pad_l - pad_r
    plot_h = h - pad_t - pad_b
    vmax = max(values) if values and max(values) > 0 else 1.0
    n = len(labels)
    gap = 14
    bw = (plot_w - gap * (n + 1)) / max(n, 1)

    parts = [_svg_header(w, h)]
    parts.append(f'<text x="{w/2}" y="26" text-anchor="middle" font-size="16" font-weight="600">{_esc(title)}</text>')
    # axis
    parts.append(f'<line x1="{pad_l}" y1="{pad_t+plot_h}" x2="{pad_l+plot_w}" y2="{pad_t+plot_h}" stroke="#999"/>')
    parts.append(f'<line x1="{pad_l}" y1="{pad_t}" x2="{pad_l}" y2="{pad_t+plot_h}" stroke="#999"/>')
    parts.append(f'<text x="{pad_l-8}" y="{pad_t+8}" text-anchor="end" font-size="11" fill="#555">{_esc(_num(vmax))}{_esc(unit)}</text>')
    parts.append(f'<text x="{pad_l-8}" y="{pad_t+plot_h}" text-anchor="end" font-size="11" fill="#555">0</text>')

    for i, (lab, val) in enumerate(zip(labels, values)):
        x = pad_l + gap + i * (bw + gap)
        bh = (val / vmax) * plot_h if vmax else 0
        y = pad_t + plot_h - bh
        color = _PALETTE[i % len(_PALETTE)]
        parts.append(f'<rect x="{x:.1f}" y="{y:.1f}" width="{bw:.1f}" height="{bh:.1f}" fill="{color}" rx="3"/>')
        parts.append(f'<text x="{x+bw/2:.1f}" y="{y-5:.1f}" text-anchor="middle" font-size="10" fill="#333">{_esc(_num(val))}</text>')
        parts.append(
            f'<text x="{x+bw/2:.1f}" y="{pad_t+plot_h+16:.1f}" text-anchor="end" font-size="10" '
            f'transform="rotate(-35 {x+bw/2:.1f} {pad_t+plot_h+16:.1f})" fill="#333">{_esc(lab)}</text>'
        )
    parts.append("</svg>")
    return "".join(parts)


def _scatter(title: str, labels: list[str], xs: list[float], ys: list[float], xlab: str, ylab: str) -> str:
    w, h = 720, 420
    pad_l, pad_r, pad_t, pad_b = 70, 30, 50, 60
    plot_w, plot_h = w - pad_l - pad_r, h - pad_t - pad_b
    xmax = max(xs) if xs else 1.0
    ymax = max(ys) if ys and max(ys) > 0 else 1.0
    parts = [_svg_header(w, h)]
    parts.append(f'<text x="{w/2}" y="26" text-anchor="middle" font-size="16" font-weight="600">{_esc(title)}</text>')
    parts.append(f'<line x1="{pad_l}" y1="{pad_t+plot_h}" x2="{pad_l+plot_w}" y2="{pad_t+plot_h}" stroke="#999"/>')
    parts.append(f'<line x1="{pad_l}" y1="{pad_t}" x2="{pad_l}" y2="{pad_t+plot_h}" stroke="#999"/>')
    parts.append(f'<text x="{pad_l+plot_w/2}" y="{h-18}" text-anchor="middle" font-size="12" fill="#555">{_esc(xlab)}</text>')
    parts.append(f'<text x="18" y="{pad_t+plot_h/2}" text-anchor="middle" font-size="12" fill="#555" transform="rotate(-90 18 {pad_t+plot_h/2})">{_esc(ylab)}</text>')
    for i, (lab, x, y) in enumerate(zip(labels, xs, ys)):
        px = pad_l + (x / xmax) * plot_w if xmax else pad_l
        py = pad_t + plot_h - (y / ymax) * plot_h if ymax else pad_t + plot_h
        color = _PALETTE[i % len(_PALETTE)]
        parts.append(f'<circle cx="{px:.1f}" cy="{py:.1f}" r="6" fill="{color}" fill-opacity="0.85"/>')
        parts.append(f'<text x="{px+9:.1f}" y="{py+4:.1f}" font-size="11" fill="#333">{_esc(lab)}</text>')
    # "cheaper + better" hint (bottom-right is ideal: high accuracy, low cost)
    parts.append(f'<text x="{pad_l+plot_w}" y="{pad_t+plot_h-6}" text-anchor="end" font-size="11" fill="#16a34a">← cheaper · more accurate ↗ is worse-cost</text>')
    parts.append("</svg>")
    return "".join(parts)


def _num(v: float) -> str:
    if v == 0:
        return "0"
    if abs(v) < 0.001:
        return f"{v:.2e}"
    if abs(v) < 1:
        return f"{v:.4f}".rstrip("0").rstrip(".")
    return f"{v:.1f}"


def render(results: list[dict], outdir: str, want_png: bool = False) -> list[str]:
    os.makedirs(outdir, exist_ok=True)
    labels = [r["adapter"] + ("*" if r.get("is_estimate") else "") for r in results]
    costs = [r.get("system_cost_usd", 0.0) for r in results]
    lat = [r.get("avg_recall_latency_ms", 0.0) for r in results]
    acc = [r.get("accuracy", 0.0) * 100 for r in results]

    written = []
    charts = {
        "cost_bar.svg": _bar_chart("Memory-system cost per suite (USD)", labels, costs, "$"),
        "latency_bar.svg": _bar_chart("Average recall latency (ms)", labels, lat, "ms"),
        "accuracy_cost.svg": _scatter("Accuracy vs. cost frontier", labels, costs, acc, "memory-system cost (USD)", "accuracy (%)"),
    }
    for fname, svg in charts.items():
        path = os.path.join(outdir, fname)
        with open(path, "w") as f:
            f.write(svg)
        written.append(path)

    if want_png:
        written += _render_png(labels, costs, lat, acc, outdir)
    return written


def _render_png(labels, costs, lat, acc, outdir) -> list[str]:
    try:
        import matplotlib

        matplotlib.use("Agg")
        import matplotlib.pyplot as plt
    except Exception:
        print("  (matplotlib not installed; skipping PNG — SVGs were written)")
        return []
    written = []
    for data, title, fname, ylab in [
        (costs, "Memory-system cost per suite (USD)", "cost_bar.png", "USD"),
        (lat, "Average recall latency (ms)", "latency_bar.png", "ms"),
    ]:
        fig, ax = plt.subplots(figsize=(8, 4))
        ax.bar(labels, data, color=_PALETTE[: len(labels)])
        ax.set_title(title)
        ax.set_ylabel(ylab)
        plt.xticks(rotation=35, ha="right")
        fig.tight_layout()
        p = os.path.join(outdir, fname)
        fig.savefig(p, dpi=120)
        plt.close(fig)
        written.append(p)
    fig, ax = plt.subplots(figsize=(8, 4.5))
    ax.scatter(costs, acc, c=_PALETTE[: len(labels)], s=60)
    for lab, x, y in zip(labels, costs, acc):
        ax.annotate(lab, (x, y), textcoords="offset points", xytext=(6, 3), fontsize=9)
    ax.set_xlabel("memory-system cost (USD)")
    ax.set_ylabel("accuracy (%)")
    ax.set_title("Accuracy vs. cost frontier")
    fig.tight_layout()
    p = os.path.join(outdir, "accuracy_cost.png")
    fig.savefig(p, dpi=120)
    plt.close(fig)
    written.append(p)
    return written


def render_from_file(results_json: str, outdir: str, want_png: bool = False) -> list[str]:
    with open(results_json) as f:
        payload = json.load(f)
    return render(payload["results"], outdir, want_png)
