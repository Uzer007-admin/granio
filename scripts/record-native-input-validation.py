"""Record complete native-input runs without retaining unrelated desktop speech."""

import argparse
import hashlib
import json
from pathlib import Path
import runpy

ROOT = Path(__file__).resolve().parent.parent
COMMON = runpy.run_path(str(ROOT / "scripts/record-display-validation.py"))


def passed_tests(path):
    tests = set()
    for line in path.read_text(encoding="utf-8-sig").splitlines():
        if not line.startswith("{"):
            continue
        event = json.loads(line)
        if event["Action"] == "pass" and event.get("Test"):
            tests.add(event["Test"])
    return tests


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--directory", type=Path, default=ROOT / "docs/native-input-validation")
    args = parser.parse_args()
    directory = args.directory
    names = ("windows-suite.jsonl", "windows-headless.jsonl", "linux-suite.jsonl")
    suites = {name: COMMON["summarize"](directory / name) for name in names}
    if any(s["failures"] or s["unfinished_packages"] for s in suites.values()):
        raise SystemExit("A full suite failed or did not finish")
    required = {
        "windows-suite.jsonl": {
            "TestWindowsAccessibilityOffscreenHierarchy", "TestWindowsAccessibilityProviderLifetime",
            "TestTSFTextStoreNativeContract", "TestTSFNativeEditSession", "TestNativeCanvasAndPopupMaterials",
            "TestClipboardMIMENativeTransfer", "TestExternalWindowsOLEDrop", "TestNativeDisplayResolutionMatrix",
        },
        "linux-suite.jsonl": {
            "TestATSPINativeOffscreenHierarchyAndCache", "TestATSPINativeClient",
            "TestNativeCanvasAndPopupMaterials", "TestNativeDisplayResolutionMatrix",
            "TestClipboardMIMENativeTransfer", "TestExternalX11DataDrop",
        },
    }
    for name, tests in required.items():
        missing = tests - passed_tests(directory / name)
        if missing:
            raise SystemExit(f"Native checks were skipped or absent in {name}: {sorted(missing)}")
    headless = passed_tests(directory / "windows-headless.jsonl")
    if not {"TestAccessibilityVirtualCollectionsDoNotRealizeOnRead", "TestAccessibilityVirtualClosedTree"} <= headless:
        raise SystemExit("Logical collection tests are absent from the headless suite")
    expected_logs = {
        "real-anthy.txt": "real IBus Anthy Japanese commit: PASS",
        "real-pinyin-orca.txt": "real Fcitx Pinyin commits, Orca field speech and external AT-SPI offscreen/password checks: PASS",
        "ibus-privacy-lifetime.txt": "--- PASS: TestIBusNativePasswordAndLifecycle",
        "material-gles2.txt": "--- PASS: TestNativeCanvasAndPopupMaterials",
        "material-wayland.txt": "--- PASS: TestATSPINativeOffscreenHierarchyAndCache",
        "wayland-drag.txt": "external Wayland unsupported MIME: rejected, PASS",
        "nvda-check.txt": "PASS real NVDA protected field and native UIA password redaction",
    }
    for name, marker in expected_logs.items():
        log = (directory / name).read_text(encoding="utf-8-sig")
        if marker not in log or "WARNING: DATA RACE" in log or "--- FAIL:" in log:
            raise SystemExit(f"Native acceptance is missing or failed: {name}")
    local = COMMON["source_hashes"]()
    windows = json.loads((directory / "windows-source-hashes.json").read_text())
    linux = json.loads((directory / "linux-source-hashes.json").read_text())
    if local != windows or local != linux:
        raise SystemExit("Current, Windows and Linux source fingerprints differ")
    summary = {
        "suites": suites,
        "required_native_checks": {name: sorted(tests) for name, tests in required.items()},
        "matching_source_inputs": len(local),
        "cleanup": {
            "linux": "owned temporary directory removed",
            "windows": "all owned temporary files and directory removed, including both NVDA proxy DLLs; verified 2026-10-13 after authorized Yandex Browser shutdown and Explorer shell restart",
            "windows_language_profiles": "restored and compared to the saved original",
        },
        "real_input_acceptance": {
            "linux_anthy": "commit 日本語, separate preedit, undo, cancellation; native race probe",
            "linux_pinyin": "commits 你好 and 中国; native race probe",
            "linux_ibus_privacy_lifetime": "password redacted and closed context released; native race test",
            "orca": "43.1: actual CJK input field speech; external libatspi offscreen/selection/password client",
            "nvda": "2026.2: actual Fyne/entry/protected speech, nonempty password redaction via external UIA Value/Text queries, silence synth; basic scenario only",
            "windows_microsoft_japanese": "FAILED activation (stage 4, HRESULT 0x80004005); not accepted",
            "wayland_drag": "independent wl_data_device source: Unicode, HTML, URI, 512 KiB binary; unsupported MIME rejected",
        },
    }
    (directory / "summary.json").write_text(json.dumps(summary, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    artifacts = {
        p.name: {"bytes": p.stat().st_size, "sha256": hashlib.sha256(p.read_bytes()).hexdigest()}
        for p in sorted(directory.iterdir()) if p.is_file() and p.name != "evidence.json"
    }
    (directory / "evidence.json").write_text(json.dumps({"source_hash_algorithm": "SHA-256 exact bytes", "artifacts": artifacts}, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"suites": suites, "matching_source_inputs": len(local)}, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
