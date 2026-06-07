"""Adapter registry — resolve adapter names to instances.

Names:
  mock                       zero-cost keyword baseline (always available)
  yazi                       real adapter via yazictl (needs a running server)
  yazi-student /
  yazi-standard              real Yazi composition profiles via `memory recall`
                             (ground-truth metered cost; needs a running server)
  sim-mem0 / sim-zep /
  sim-letta / sim-memos /
  sim-mem9                   dependency-free cost ESTIMATORS
  mem0 / zep / letta /
  memos / mem9 / milvus      real adapters (opt-in, import-guarded)

Special selectors for --adapters:
  all            -> mock + yazi + every sim-*
  sim            -> every sim-*
  real           -> mock + yazi + every real external adapter
  yazi-profiles  -> yazi-student + yazi-standard
"""

from __future__ import annotations

from framework.adapter import MemoryAdapter

from adapters.mock_adapter import MockAdapter
from adapters.yazi_adapter import YaziAdapter
from adapters.yazi_profile_adapter import YaziProfileAdapter
from adapters.simulator import PROFILES, SimulatedAdapter
from adapters import external

_SIM_NAMES = [f"sim-{k}" for k in PROFILES]
_YAZI_PROFILES = ["student", "standard"]
_YAZI_PROFILE_NAMES = [f"yazi-{p}" for p in _YAZI_PROFILES]
_REAL_EXTERNAL = {
    "mem0": external.Mem0Adapter,
    "zep": external.ZepAdapter,
    "letta": external.LettaAdapter,
    "memos": external.MemOSAdapter,
    "mem9": external.Mem9Adapter,
    "milvus": external.MilvusAdapter,
}


def available_names() -> list[str]:
    return ["mock", "yazi", *_YAZI_PROFILE_NAMES, *_SIM_NAMES, *_REAL_EXTERNAL.keys()]


def build(name: str) -> MemoryAdapter:
    name = name.strip()
    if name == "mock":
        return MockAdapter()
    if name == "yazi":
        return YaziAdapter()
    if name in _YAZI_PROFILE_NAMES:
        return YaziProfileAdapter(name[len("yazi-"):])
    if name.startswith("sim-"):
        return SimulatedAdapter(name[len("sim-"):])
    if name in _REAL_EXTERNAL:
        return _REAL_EXTERNAL[name]()
    raise KeyError(f"unknown adapter '{name}'. Available: {available_names()}")


def expand(selector: str) -> list[str]:
    """Expand a selector token into concrete adapter names."""
    if selector == "all":
        return ["mock", "yazi", *_SIM_NAMES]
    if selector == "sim":
        return list(_SIM_NAMES)
    if selector == "real":
        return ["mock", "yazi", *_REAL_EXTERNAL.keys()]
    if selector == "yazi-profiles":
        return list(_YAZI_PROFILE_NAMES)
    return [selector]


def resolve(selectors: list[str]) -> list[str]:
    names: list[str] = []
    for sel in selectors:
        for n in expand(sel):
            if n not in names:
                names.append(n)
    return names
