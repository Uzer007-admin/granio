"""Summarize completed display/input test runs and fingerprint public sources."""

import argparse
import hashlib
import json
import os
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SOURCE_SUFFIXES = {".go", ".c", ".h", ".m", ".mm", ".s", ".S", ".glsl", ".xml"}
SOURCE_NAMES = {"go.mod", "go.sum", "go.work", "go.work.sum"}


def checksum(path):
    with path.open("rb") as stream:
        return hashlib.file_digest(stream, "sha256").hexdigest()


def source_hashes():
    result = {}
    for directory, children, files in os.walk(ROOT):
        children[:] = [name for name in children if not name.startswith(".") and name not in {"docs", "bin", "patches", "node_modules"}
                       and not (Path(directory).name == "testdata" and name == "failed")]
        for name in files:
            path = Path(directory) / name
            if path.suffix in SOURCE_SUFFIXES or name in SOURCE_NAMES:
                result[path.relative_to(ROOT).as_posix()] = checksum(path)
    return dict(sorted(result.items()))


def summarize(path):
    result = {"packages_pass": 0, "tests_pass": 0, "failures": 0, "skips": [], "external_stdout": []}
    pending = set()
    with path.open(encoding="utf-8-sig") as stream:
        for line in stream:
            if not line.strip():
                continue
            if line.startswith("SpiRegistry daemon is running"):
                result["external_stdout"].append(line.rstrip())
                continue
            event = json.loads(line)
            action, package, test = event["Action"], event.get("Package"), event.get("Test")
            if action == "start":
                pending.add(package)
            if action in {"pass", "fail", "skip"} and not test:
                pending.discard(package)
            if action == "pass":
                result["tests_pass" if test else "packages_pass"] += 1
            if action == "fail":
                result["failures"] += 1
            if action == "skip" and test:
                result["skips"].append(test)
    result["unfinished_packages"] = sorted(pending)
    return result


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--directory", type=Path, default=ROOT / "docs/display-validation")
    parser.add_argument("--write-source-hashes", type=Path)
    parser.add_argument("--remote-source-hashes", type=Path)
    parser.add_argument("--allow-source-drift", action="store_true", help="Record subsequent shared-tree changes without treating them as validated")
    args = parser.parse_args()
    sources = source_hashes()
    if args.write_source_hashes:
        args.write_source_hashes.write_text(json.dumps(sources, indent=2), encoding="utf-8")
        print(f"Fingerprinted {len(sources)} public source inputs")
        return
    names = ("windows-native-final.json", "windows-headless-final.json", "linux-native-final.json")
    results = {name: summarize(args.directory / name) for name in names}
    driver_repeat = args.directory / "windows-native-driver-final.json"
    if driver_repeat.is_file():
        results[driver_repeat.name] = summarize(driver_repeat)
    (args.directory / "suite-summary.json").write_text(json.dumps(results, indent=2), encoding="utf-8")
    artifacts = {path.name: {"bytes": path.stat().st_size, "sha256": checksum(path)}
                 for path in sorted(args.directory.iterdir()) if path.is_file() and path.name != "evidence.json"}
    evidence = {"source_hash_algorithm": "SHA-256 (exact bytes)", "source_inputs": sources, "artifacts": artifacts}
    if args.remote_source_hashes:
        remote = json.loads(args.remote_source_hashes.read_text(encoding="utf-8"))
        evidence["validated_linux_source_inputs"] = remote
        missing = sorted(set(sources) - set(remote))
        changed = sorted(name for name in sources.keys() & remote.keys() if sources[name] != remote[name])
        evidence["linux_source_comparison"] = {"matching": len(sources) - len(missing) - len(changed), "missing": missing, "changed": changed}
    (args.directory / "evidence.json").write_text(json.dumps(evidence, indent=2), encoding="utf-8")
    print(json.dumps(results, indent=2))
    if any(value["failures"] or value["unfinished_packages"] for value in results.values()):
        raise SystemExit("A full suite failed or is incomplete")
    if not args.allow_source_drift and (evidence.get("linux_source_comparison", {}).get("missing") or evidence.get("linux_source_comparison", {}).get("changed")):
        raise SystemExit("Linux and local source inputs differ; inspect evidence.json")


if __name__ == "__main__":
    main()
