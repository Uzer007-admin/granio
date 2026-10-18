"""Record completed material pipeline validation; reject missing/failing evidence."""

import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil
import statistics

ROOT = Path(__file__).resolve().parent.parent
DEST = ROOT / "docs/graphics-assessment/material-pipeline"


def digest(path):
    return hashlib.sha256(path.read_bytes().replace(b"\r\n", b"\n")).hexdigest()


def benchmarks(path):
    groups = {}
    for line in path.read_text(encoding="utf-8-sig").splitlines():
        match = re.search(r"(Benchmark\w+)-\d+\s+\d+\s+([\d.]+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op", line)
        if match:
            if int(match[3]) or int(match[4]):
                raise ValueError(f"allocating benchmark: {path}: {line}")
            groups.setdefault(match[1], []).append(float(match[2]))
    if not groups or any(len(samples) != 3 for samples in groups.values()):
        raise ValueError(f"expected three benchmark samples: {path}")
    return {name: {"samples_ns": values, "median_ns": statistics.median(values), "bytes": 0, "allocations": 0} for name, values in groups.items()}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--work", required=True, type=Path)
    args = parser.parse_args()
    work = args.work.resolve()
    windows = (
        "unit-windows.txt", "final-unit-windows.txt", "vet-windows.txt", "fyne-windows.txt",
        "dbus-windows.txt", "generation-windows.txt", "gpu-windows.txt",
        "glfw-3.3-windows.txt", "glfw-3.4-windows.txt", "glfw-3.5-windows.txt",
        "bench-material-windows.txt", "bench-delta-windows.txt", "bench-surface-windows.txt",
    )
    linux = (
        "unit.txt", "vet.txt", "dbus.txt", "appearance-fyne.txt", "fyne.txt",
        "glfw-3.3.txt", "glfw-3.4.txt", "glfw-3.5.txt", "wasm-build.txt",
        "material-x11.txt", "material-ext21.txt", "material-gles20.txt",
        "material-wayland.txt", "material-wayland-gles20.txt",
        "bench-material.txt", "bench-delta.txt", "bench-surface.txt", "environment.txt",
    )
    logs = {}
    for platform, directory, names in (("windows", work, windows), ("linux", work / "linux-logs", linux)):
        for name in names:
            path = directory / name
            content = path.read_text(encoding="utf-8-sig")
            if re.search(r"(^FAIL\b|--- FAIL:|WARNING: DATA RACE|panic:)", content, re.M):
                raise ValueError(f"failed validation: {path}")
            logs[f"{platform}/{name}"] = path
    gpu = {}
    for name, path in logs.items():
        if "gpu-windows" not in name and "/material-" not in name:
            continue
        report = json.loads(next(line for line in path.read_text(encoding="utf-8-sig").splitlines() if line.startswith("[{")))
        if len(report) != 2:
            raise ValueError(f"expected two contexts: {path}")
        for item in report:
            for key in ("GPUSnapshot", "NoiseStable", "LinearAndEncoded", "StateRestored", "AlphaContract", "ResizeFallback", "GPUDownsample", "SurfaceTracked", "AdaptiveRenderer"):
                if not item[key]:
                    raise ValueError(f"failed {key}: {path}")
            if item["MaximumPixelError"] > 3 or item["CaptureMaximumPixelError"] > 3 or item["BatchedQueryTransitions"] != 1:
                raise ValueError(f"invalid GPU result: {path}")
            if any(item[key] for key in ("ShaderSourceAllocs", "RenderTargetAllocs", "CaptureAllocs", "BatchedQueryAllocs", "AdaptiveCacheAllocs")):
                raise ValueError(f"allocating GPU path: {path}")
        gpu[name] = report
    if len(gpu) != 6:
        raise ValueError("expected Windows and five Linux configurations")
    replay = json.loads((work / "patch-replay.json").read_text(encoding="utf-8"))
    if len(replay) != 5 or not all(item["identical"] for item in replay):
        raise ValueError("patch replay incomplete")
    bench = {name: benchmarks(path) for name, path in logs.items() if "/bench-" in name}
    sources = set((ROOT / "deps/gl/material").glob("*.go"))
    sources.update((ROOT / "deps/dbus/portal").glob("*.go"))
    sources.update(ROOT / name for name in ("deps/glow/tmpl/package.tmpl", "deps/glow/dispatch_test.go", "scripts/material-runtime-probe/main.go"))
    for version in ("3.3", "3.4", "3.5"):
        directory = ROOT / "deps/glfw" / ("v" + version) / "glfw"
        sources.update(directory / name for name in ("surface.go", "surface_tracker.go", "surface_tracker_test.go", "surface_integration_test.go"))
    for profile in ("v2.1/fyne", "v2.1/material", "v2.0/gles2"):
        sources.update(ROOT / "deps/gl" / profile / name for name in ("package.go", "symbol_manifest.json"))
    hashes = {path.relative_to(ROOT).as_posix(): digest(path) for path in sorted(sources)}
    for name, path in logs.items():
        target = DEST / name
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copyfile(path, target)
    shutil.copyfile(work / "patch-replay.json", DEST / "patch-replay.json")
    data = {
        "date": "2026-10-12", "go": "1.27.1", "race": True, "cgocheck2": True, "gpu_checkptr": 2,
        "gpu": gpu, "benchmarks": bench, "patch_replay": replay,
        "source_sha256_lf": hashes, "log_sha256_lf": {name: digest(path) for name, path in logs.items()},
        "limits": ["No Fyne canvas/popup material graph or WindowVisuals state machine integrated", "Four-tap factors 4/8 approximate, not exact area filters", "Pool budget covers owned RGBA8 texels; external source/program/window/driver storage needs compositor accounting", "No native material visual acceptance, physical mixed-DPI, macOS/BSD/ARM/mobile/WebGL runtime or 5x60s present/input performance acceptance", "Windows D-Bus checks are daemon-independent; real IPC normal/no-TCP tested on Linux"],
    }
    output = ROOT / "docs/graphics-assessment/material-pipeline.json"
    output.write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
    (work / "source-manifest.json").write_text(json.dumps(hashes, indent=2) + "\n", encoding="utf-8")
    print(f"Recorded {len(logs)} logs, {len(hashes)} sources and {sum(map(len, gpu.values()))} native contexts")
    print(json.dumps(bench, indent=2))


if __name__ == "__main__":
    main()
