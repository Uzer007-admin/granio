"""Record samples and validate the exact library sources tested on Linux."""

import argparse
import hashlib
import json
from pathlib import Path
import re
import statistics

ROOT = Path(__file__).resolve().parent.parent


def samples(path):
    result = {}
    for match in re.finditer(r"^(Benchmark\w+)-\d+\s+\d+\s+([\d.]+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op$", path.read_text(), re.M):
        name, ns, size, allocations = match.groups()
        result.setdefault(name, []).append({"ns": float(ns), "bytes": int(size), "allocations": int(allocations)})
    if not result or any(len(runs) != 3 for runs in result.values()):
        raise ValueError(f"Expected three samples: {path}")
    return {name: {"samples": runs, "median_ns": statistics.median(run["ns"] for run in runs)} for name, runs in result.items()}


def probe(path):
    return next(json.loads(line) for line in path.read_text().splitlines() if line.startswith("{"))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--workdir", required=True, type=Path)
    work = parser.parse_args().workdir
    sources = {}
    for line in (work / "linux-source-sha256.txt").read_text().splitlines():
        checksum, name = line.split(None, 1)
        path = (ROOT / name).resolve()
        if not path.is_relative_to(ROOT) or hashlib.sha256(path.read_bytes()).hexdigest() != checksum:
            raise ValueError(f"Source did not match Linux validation: {name}")
        sources[name] = checksum
    profiles = {}
    for path in sorted((ROOT / "deps/gl").rglob("symbol_manifest.json")):
        manifest = json.loads(path.read_text())
        functions, enums = manifest["functions"], manifest["enums"]
        names = [function["name"] for function in functions]
        if manifest["schema"] != 2 or names != sorted(set(names)) or enums != sorted(set(enums)):
            raise ValueError(f"Invalid schema or symbol ordering: {path}")
        for function in functions:
            if "return" not in function or "parameters" not in function:
                raise ValueError(f"Missing signature metadata: {path}")
        code = (path.parent / "package.go").read_text()
        if names != re.findall(r'case "([^"]+)":', code.split("func IsFunctionAvailable", 1)[1]):
            raise ValueError(f"Manifest and procedure table differ: {path}")
        profiles[path.parent.relative_to(ROOT).as_posix()] = {"functions": len(functions), "required": sum(function["required"] for function in functions), "enums": len(enums), "sha256": hashlib.sha256(path.read_bytes()).hexdigest()}
    linux = work / "linux-logs"
    result = {"date": "2026-10-12", "go": "1.27.1", "experiment": "cgocheck2",
              "windows_geometry": samples(work / "geometry-benchmark-windows.txt"),
              "linux_geometry": samples(linux / "geometry-benchmark.txt"),
              "windows_signal": samples(work / "signal-benchmark-windows.txt"),
              "linux_signal": samples(linux / "signal-benchmark.txt"),
              "windows_material": probe(work / "material-windows.txt"),
              "linux_material_x11": probe(linux / "material-x11.txt"),
              "linux_material_wayland": probe(linux / "material-wayland.txt"),
              "linux_visible_wayland_surfaces": probe(linux / "surface-wayland.txt"),
              "verified_sources": sources, "profiles": profiles,
              "limits": ["Three sequential samples per variant; no frame latency/FPS inference.", "Signal benchmark measures a pre-existing Signal delivered to one full queue, excluding decoding and IPC.", "Hidden material windows have not entered an output; their scale need not equal the compositor's output scale.", "Pure capability package WASM compile is not browser rendering validation.", "Native Windows/macOS backdrop materials and Fyne's full material graph are not implemented by these primitives."]}
    output = ROOT / "docs/graphics-assessment/library-development.json"
    output.write_text(json.dumps(result, indent=2) + "\n", encoding="utf-8")
    print(f"Verified {len(sources)} exact Linux sources and {len(profiles)} schema-2 profiles")


if __name__ == "__main__":
    main()
