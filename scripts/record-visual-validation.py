"""Record native surface benchmarks and verify the Linux validation sources."""

import argparse
import hashlib
import json
from pathlib import Path
import re
import statistics

ROOT = Path(__file__).resolve().parent.parent


def benchmarks(path):
    samples = {}
    for match in re.finditer(r"^(Benchmark(?:Legacy)?SurfaceMetrics)-\d+\s+\d+\s+([\d.]+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op$", path.read_text().strip(), re.M):
        name, ns, size, allocations = match.groups()
        samples.setdefault(name, []).append({"ns": float(ns), "bytes": int(size), "allocs": int(allocations)})
    if any(len(samples.get(name, [])) != 3 for name in ("BenchmarkLegacySurfaceMetrics", "BenchmarkSurfaceMetrics")):
        raise ValueError(f"Expected three runs of both benchmarks: {path}")
    return {name: {"samples": runs, "median_ns": statistics.median(run["ns"] for run in runs)} for name, runs in samples.items()}


def probe(path):
    return next(json.loads(line) for line in path.read_text().splitlines() if line.startswith("{"))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--workdir", required=True, type=Path)
    args = parser.parse_args()
    work = args.workdir
    linux = work / "linux-logs"
    sources = {}
    for line in (work / "linux-source-sha256.txt").read_text().splitlines():
        checksum, name = line.split(None, 1)
        path = (ROOT / name).resolve()
        if not path.is_relative_to(ROOT):
            raise ValueError(f"Source escaped workspace: {name}")
        actual = hashlib.sha256(path.read_bytes()).hexdigest()
        if actual != checksum:
            raise ValueError(f"Linux source differs from workspace: {name}")
        sources[name] = checksum
    profiles = {}
    for path in sorted((ROOT / "deps/gl").rglob("symbol_manifest.json")):
        manifest = json.loads(path.read_text())
        functions, enums = manifest["functions"], manifest["enums"]
        names = [function["name"] for function in functions]
        if names != sorted(set(names)) or enums != sorted(set(enums)):
            raise ValueError(f"Unsorted or duplicate manifest symbols: {path}")
        code = (path.parent / "package.go").read_text()
        available = re.findall(r'case "([^"]+)":', code.split("func IsFunctionAvailable", 1)[1])
        if names != available:
            raise ValueError(f"Manifest/availability table mismatch: {path}")
        profiles[path.parent.relative_to(ROOT).as_posix()] = {"functions": len(functions), "enums": len(enums), "sha256": hashlib.sha256(path.read_bytes()).hexdigest()}
    result = {"date": "2026-10-12", "go": "1.27.1", "experiment": "cgocheck2",
              "windows": {"os_build": "10.0.26300", "surface_benchmarks": benchmarks(work / "surface-benchmark-windows.txt"), "probe": probe(work / "surface-windows.txt")},
              "linux": {"compositor": "Weston 10.0.1 headless/pixman, scale=2", "surface_benchmarks": benchmarks(linux / "surface-benchmark.txt"), "x11": probe(linux / "surface-x11.txt"), "wayland": probe(linux / "surface-wayland.txt")},
              "verified_linux_sources": sources, "generated_profiles": profiles,
              "limits": ["Three sequential benchmark runs per variant; no confidence intervals or frame performance claims.", "X11/Xvfb latency did not improve in these samples; both platforms removed three Go allocations.", "GL Init probes report repeated averages, not an equivalent benchmark series.", "Wayland scale=2 is integer headless output scaling, not fractional-DPI hardware validation.", "Portal service is a test export on a real bus daemon, not a desktop portal integration."]}
    output = ROOT / "docs/graphics-assessment/visual-validation.json"
    output.write_text(json.dumps(result, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    print(f"Recorded {len(sources)} matching Linux sources and {len(profiles)} generated profile manifests")


if __name__ == "__main__":
    main()
