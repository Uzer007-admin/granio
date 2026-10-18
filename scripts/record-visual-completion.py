"""Preserve short visual acceptance evidence and verify source/reference identity."""

import argparse
import hashlib
import json
from pathlib import Path
import re
import shutil
import tarfile

ROOT = Path(__file__).resolve().parent.parent
WINDOWS_LOGS = (
    "windows-root-latest.log", "windows-functional-final.log", "windows-software-ci.log",
    "windows-vet.log", "windows-vet-core.log", "windows-vet-native.log",
    "animation-final.log", "animation-vet.log", "gallery-vet.log",
    "references-tests.log", "references-final.log", "references-repeat.log",
    "references-identity.json", "environment.json",
    "wasm-latest.log", "macos-software-latest.log",
)


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def package_count(path):
    log = path.read_text(encoding="utf-8", errors="replace")
    if re.search(r"^(?:FAIL|--- FAIL:)", log, re.M):
        raise ValueError(f"Failed final suite: {path}")
    packages = re.findall(r"^ok\s+(\S+)", log, re.M)
    if not packages:
        raise ValueError(f"Empty final suite: {path}")
    return len(packages)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--workdir", type=Path, required=True)
    parser.add_argument("--local-only", action="store_true", help="Preserve local results while remote evidence is inaccessible")
    args = parser.parse_args()
    work = args.workdir.resolve()
    if not work.is_relative_to(ROOT) or not work.name.startswith(".build-tz-completion-"):
        raise ValueError("workdir must be the task's temporary workspace directory")
    destination = ROOT / "docs/visual-completion-validation"
    destination.mkdir(exist_ok=True)
    (destination / "windows").mkdir(exist_ok=True)
    (destination / "linux").mkdir(exist_ok=True)
    for path in (work / "logs").iterdir():
        if path.name not in WINDOWS_LOGS and not path.name.startswith("gallery-windows-"):
            continue
        shutil.copy2(path, destination / "windows" / path.name)
    for path in ([] if args.local_only else (work / "linux-logs").iterdir()):
        if path.is_file():
            shutil.copy2(path, destination / "linux" / path.name)
    references = work / "references-final"
    repeat = work / "references-repeat"
    manifest = json.loads((references / "manifest.json").read_text())
    for item in manifest["references"]:
        name = item["File"]
        checksum = sha(references / name)
        if checksum != item["SHA256"] or sha(repeat / name) != checksum:
            raise ValueError(f"Non-deterministic reference: {name}")
    shutil.copytree(references, destination / "references", dirs_exist_ok=True)

    # The remote source manifest is generated from the actual exported checkout,
    # after all full/delta archives are applied. Compare each current code input.
    sources = {}
    remote_lines = [] if args.local_only else (work / "linux-source-sha256.txt").read_text().splitlines()
    for line in remote_lines:
        checksum, name = line.split(None, 1)
        relative = name.removeprefix("./")
        path = (ROOT / relative).resolve()
        if not path.is_relative_to(ROOT):
            raise ValueError(f"Source escaped repository: {relative}")
        if path.suffix not in (".go", ".c", ".h", ".m", ".mm") and path.name not in ("go.mod", "go.sum", "go.work", "go.work.sum"):
            continue
        if sha(path) != checksum:
            raise ValueError(f"Linux source differs from final workspace: {relative}")
        sources[relative] = checksum
    if args.local_only:
        names = set()
        for name in ("sources.tar.gz", "sources-delta.tar.gz"):
            with tarfile.open(work / "export" / name) as archive:
                names.update(member.name for member in archive.getmembers() if member.isfile())
        for relative in sorted(names):
            path = (ROOT / relative).resolve()
            if not path.is_relative_to(ROOT):
                raise ValueError(f"Source escaped repository: {relative}")
            if path.suffix in (".go", ".c", ".h", ".m", ".mm") or path.name in ("go.mod", "go.sum", "go.work", "go.work.sum"):
                sources[relative] = sha(path)
    if len(sources) < 500:
        raise ValueError("incomplete remote source identity inventory")
    (destination / "source-manifest.json").write_text(json.dumps(sources, indent=2) + "\n", encoding="utf-8")
    acceptance = {
        "date": "2026-10-14", "go": "1.27.1", "token_version": 1,
        "root_race_cgocheck2_packages": {
            "windows": package_count(work / "logs/windows-root-latest.log"),
            "linux": None if args.local_only else package_count(work / "linux-logs/root-race.log"),
        },
        "native_cycles": 100,
        "native_functional_windows_checkptr2": True,
        "live_text_scale": [1, 1.5, 2, 2.25, 1],
        "software_dpi": [1, 1.5, 2], "software_text_scale": [1, 1.5, 2, 2.25],
        "opaque_reference_images": len(manifest["references"]),
        "references_repeat_sha256_identical": True,
        "source_snapshot_files": len(sources),
        "matching_linux_source_files": None if args.local_only else len(sources),
        "linux_evidence_status": "earlier native/root run passed; final exporter delta and raw-log retrieval blocked by SSH reset" if args.local_only else "final source identity verified",
        "temporary_cleanup": {"local": "pending", "remote": "SSH unavailable; /tmp/fvc14-c83f cleanup pending" if args.local_only else "pending"},
        "wasm_build": True, "macos_arm64_software_build": True,
        "windows_native_vet_unsafeptr": "seven existing clipboard/IMM bridge diagnostics; other analyzers passed; raw log retained",
        "kde": {
            "f-test-osnovacustom": {"ip": "172.21.9.27", "expected_session": "X11", "tested": False},
            "f-autotest-altkde": {"ip": "172.21.9.21", "tested": False},
            "reason": "SSH no route/timeout; user confirmed inactive hosts",
        },
        "long_performance_runs": "excluded by user",
        "real_os_setting_changes_tested": False,
        "full_hardware_tz_accepted": False,
    }
    (destination / "acceptance.json").write_text(json.dumps(acceptance, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    print(json.dumps(acceptance, ensure_ascii=False))


if __name__ == "__main__":
    main()
