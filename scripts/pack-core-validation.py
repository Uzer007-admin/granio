"""Подготовка исходников Granio и публичных модулей для автономных проверок Linux."""

import argparse
import json
from pathlib import Path
import subprocess
import tarfile

ROOT = Path(__file__).resolve().parent.parent
MODULES = (".",)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", required=True, type=Path)
    parser.add_argument("--sources-only", action="store_true")
    parser.add_argument("--delta", action="store_true", help="Pack source changes since the previous full archive")
    args = parser.parse_args()
    args.output.mkdir(parents=True, exist_ok=True)
    sources = set()
    submodules = subprocess.check_output(["git", "-C", str(ROOT), "config", "--file", ".gitmodules", "--get-regexp", "path"], text=True)
    modules = list(MODULES)+[line.split(maxsplit=1)[1] for line in submodules.splitlines()]
    for relative in modules:
        directory = ROOT / relative
        files = subprocess.check_output(["git", "-C", str(directory), "ls-files", "--cached", "--others", "--exclude-standard", "-z"])
        for name in files.split(b"\0"):
            if not name:
                continue
            path = directory / name.decode("utf-8")
            parts = path.relative_to(ROOT).parts
            if parts[0] in ("docs", "bin") or any(part.startswith(".build-") for part in parts):
                continue
            if path.is_file():
                sources.add(path)
    sources.update(ROOT / name for name in ("go.work", "go.work.sum"))
    # New module directories can contain ignored C/header sources. Include all
    # native inputs from configured workspace modules, not just Git's inventory.
    workspace = json.loads(subprocess.check_output(["go", "work", "edit", "-json"], text=True))
    for module in workspace["Use"]:
        directory = (ROOT / module["DiskPath"]).resolve()
        if directory == ROOT or not directory.is_relative_to(ROOT):
            continue
        sources.update(path for path in directory.rglob("*") if path.is_file() and path.suffix in (".c", ".h", ".m", ".go", ".mm"))
    target = args.output / "sources.tar.gz"
    if args.delta:
        with tarfile.open(target, "r:gz") as previous:
            baseline = {member.name: member for member in previous.getmembers() if member.isfile()}
        sources = {path for path in sources if (member := baseline.get(path.relative_to(ROOT).as_posix())) is None
                   or path.stat().st_size != member.size or abs(path.stat().st_mtime-member.mtime) > 0.00001}
        target = args.output / "sources-delta.tar.gz"
    with tarfile.open(target, "w:gz") as archive:
        for path in sorted(sources):
            if path.is_file():
                archive.add(path, arcname=path.relative_to(ROOT).as_posix(), recursive=False)

    if args.sources_only or args.delta:
        print(f"Packed {len(sources)} source files in {args.output}")
        return

    cache = Path(subprocess.check_output(["go", "env", "GOMODCACHE"], text=True).strip()) / "cache/download"
    output = subprocess.check_output(["go", "list", "-m", "-json", "all"], text=True)
    decoder = json.JSONDecoder()
    modules, files = [], set()
    while output.strip():
        module, end = decoder.raw_decode(output.lstrip())
        output = output.lstrip()[end:]
        modules.append(module)
        escaped = "".join("!" + c.lower() if c.isupper() else c for c in module["Path"])
        directory = cache / escaped / "@v"
        files.update(directory.glob("*.mod"))
        files.update(directory.glob("*.info"))
        if "Version" in module:
            for suffix in (".zip", ".ziphash"):
                candidate = directory / (module["Version"] + suffix)
                if candidate.is_file():
                    files.add(candidate)
    with tarfile.open(args.output / "module-proxy.tar.gz", "w:gz") as archive:
        for path in sorted(files):
            archive.add(path, arcname=path.relative_to(cache).as_posix(), recursive=False)
    (args.output / "module-graph.json").write_text(json.dumps(modules, indent=2), encoding="utf-8")
    print(f"Packed {len(sources)} files and {len(modules)} public modules in {args.output}")


if __name__ == "__main__":
    main()
