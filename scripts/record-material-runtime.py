"""Record verified native material reports, source hashes and validation limits."""

import hashlib
import json
from pathlib import Path
import re
import statistics

ROOT = Path(__file__).resolve().parent.parent
LOGS = ROOT / "docs/graphics-assessment/material-runtime"


def digest(path):
    return hashlib.sha256(path.read_bytes().replace(b"\r\n", b"\n")).hexdigest()


def bench(path):
    samples = []
    for line in path.read_text(encoding="utf-8-sig").splitlines():
        match = re.search(r"BenchmarkCachedMaterial-\d+\s+\d+\s+([\d.]+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op", line)
        if match:
            samples.append({"ns": float(match[1]), "bytes": int(match[2]), "allocations": int(match[3])})
    if len(samples) != 3 or any(s["bytes"] or s["allocations"] for s in samples):
        raise ValueError(f"invalid cache benchmark: {path}")
    return {"samples": samples, "median_ns": statistics.median(s["ns"] for s in samples)}


def main():
    gpu = {}
    paths = [LOGS / "windows/material-runtime-windows.txt"] + [LOGS / "linux" / name for name in (
        "material-x11.txt", "material-ext21.txt", "material-gles20.txt", "material-wayland.txt", "material-wayland-gles20.txt")]
    for path in paths:
        lines = path.read_text(encoding="utf-8-sig").splitlines()
        report = json.loads(next(line for line in lines if line.startswith("[{")))
        if len(report) != 2:
            raise ValueError(f"expected two independent contexts: {path}")
        for item in report:
            for key in ("GPUSnapshot", "NoiseStable", "LinearAndEncoded", "StateRestored", "AlphaContract", "ResizeFallback"):
                if not item[key]:
                    raise ValueError(f"failed {key}: {path}")
            if item["MaximumPixelError"] > 3 or item["CacheHits"] != 2001 or item["ShaderSourceAllocs"] or item["RenderTargetAllocs"]:
                raise ValueError(f"invalid GPU result: {path}")
        gpu[path.relative_to(LOGS).as_posix()] = report

    required = {
        "windows": ("unit-windows.txt", "vet-windows.txt", "fyne-windows.txt", "appearance-dispatch-windows.txt", "dbus-unit-windows.txt", "generation-windows.txt", "glfw-3.3-windows.txt", "glfw-3.4-windows.txt", "glfw-3.5-windows.txt"),
        "linux": ("unit.txt", "vet.txt", "dbus.txt", "appearance-fyne.txt", "fyne.txt", "glfw-3.3.txt", "glfw-3.4.txt", "glfw-3.5.txt", "wasm-build.txt", "material-final-unit.txt", "material-final-vet.txt"),
    }
    for platform, names in required.items():
        for name in names:
            text = (LOGS / platform / name).read_text(encoding="utf-8-sig")
            if re.search(r"(^FAIL\b|--- FAIL:|WARNING: DATA RACE|panic:)", text, re.M):
                raise ValueError(f"failed validation: {platform}/{name}")

    source_paths = []
    for relative in ("deps/gl/material", "deps/glow", "deps/dbus/portal"):
        source_paths += [p for p in (ROOT / relative).glob("*.go")]
    source_paths += [ROOT / "deps/glow/tmpl/package.tmpl"]
    for version in ("3.3", "3.4", "3.5"):
        directory = ROOT / "deps/glfw" / ("v" + version) / "glfw"
        source_paths += list(directory.glob("visuals*")) + [directory / "window.go"]
    for profile in ("v2.1/fyne", "v2.1/material", "v2.0/gles2"):
        directory = ROOT / "deps/gl" / profile
        source_paths += [directory / name for name in ("package.go", "symbol_manifest.json")]
        manifest = json.loads((directory / "symbol_manifest.json").read_text(encoding="utf-8"))
        if not manifest.get("context_dispatch"):
            raise ValueError(f"missing dispatch manifest: {profile}")
    source_paths += list((ROOT / "app").glob("system_appearance*.go"))
    source_paths += [ROOT / relative for relative in ("system_appearance.go", "app/settings.go", "app/settings_desktop.go", "app/app.go", "app/app_xdg.go", "internal/painter/gl/gl_core.go", "internal/painter/gl/gl_es.go", "internal/driver/glfw/context_dispatch_native_test.go", "scripts/material-runtime-probe/main.go")]
    data = {
        "date": "2026-10-12", "go": "1.27.1", "race": True, "cgocheck2": True, "gpu_checkptr": 2,
        "gpu": gpu, "windows_cache": bench(LOGS / "windows/cache-benchmark-windows.txt"), "linux_cache": bench(LOGS / "linux/cache-benchmark.txt"),
        "source_sha256_lf": {p.relative_to(ROOT).as_posix(): digest(p) for p in sorted(set(source_paths))},
        "log_sha256_lf": {p.relative_to(LOGS).as_posix(): digest(p) for p in sorted(LOGS.rglob("*.txt"))},
        "limits": ["DWM accepted requests are not visual acceptance; Windows 22621/22631, AMD/Intel/ARM64/RDP/VM remain untested", "Native Windows D-Bus daemon not available; Windows units and full Linux normal/no-TCP IPC validated", "WASM compile only; WebGL/mobile/macOS/BSD runtime not validated", "64 MiB is requested RGBA8 pool texel storage, excludes driver/program/window/other caches", "Fyne popup material graph and WindowVisuals state machine/observers remain product integration", "Legacy generated Ptr(uintptr)/PtrOffset vet warnings remain; handwritten layer vet passed"],
    }
    (ROOT / "docs/graphics-assessment/material-runtime.json").write_text(json.dumps(data, indent=2) + "\n", encoding="utf-8")
    print(f"Recorded {sum(map(len, gpu.values()))} native contexts, {len(source_paths)} source entries and {len(data['log_sha256_lf'])} logs")


if __name__ == "__main__":
    main()
