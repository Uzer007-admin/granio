"""Export current public project sources and their Go module cache for Linux checks."""

import argparse
import json
from pathlib import Path
import subprocess
import tarfile


ROOT = Path(__file__).resolve().parent.parent
MODULES = (".", "deps/dbus", "deps/gl", "deps/glfw", "deps/glow", "extras/apps")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output", required=True, type=Path)
    args = parser.parse_args()
    args.output.mkdir(parents=True, exist_ok=True)
    sources = set()
    for relative in MODULES:
        directory = ROOT / relative
        tracked = subprocess.check_output(["git", "-C", str(directory), "ls-files", "-z"])
        sources.update(directory / name.decode() for name in tracked.split(b"\0") if name)
	
        untracked = subprocess.check_output(["git", "-C", str(directory), "ls-files", "--others", "--exclude-standard", "-z"])
        for name in untracked.split(b"\0"):
            if name:
                path = directory / name.decode()
                if path.is_file() and not path.name.startswith(".") and (path.suffix in (".go", ".c", ".h", ".m", ".xml", ".json", ".mod", ".sum", ".md", ".txt") or path.name in ("go.mod", "go.sum")) and not any(part.startswith(".build") for part in path.relative_to(ROOT).parts):
                    sources.add(path)
    sources.update(ROOT / name for name in ("go.work", "go.work.sum", "internal/test/rounding_test.go"))
    # go.work may also reference the browser GL adapter. Preserve its source
    # when present so exported builds do not point at a missing local module.
    browser = ROOT / "deps/gl-js"
    if (browser / "go.mod").is_file():
        sources.update(path for path in browser.rglob("*") if path.is_file() and ".git" not in path.relative_to(browser).parts and not any(part.startswith(".build") for part in path.relative_to(browser).parts) and path.suffix in (".go", ".mod", ".sum", ".md", ".js"))
    sources.update(path for path in (ROOT / "scripts").rglob("*") if path.is_file() and path.suffix in (".py", ".ps1", ".sh", ".go", ".c"))
    sources.update((ROOT / "patches").glob("*.patch"))
    with tarfile.open(args.output / "sources.tar.gz", "w:gz") as archive:
        for path in sorted(sources):
            if path.is_file():
                archive.add(path, arcname=path.relative_to(ROOT).as_posix(), recursive=False)
    cache = Path(subprocess.check_output(["go", "env", "GOMODCACHE"], text=True).strip()) / "cache/download"
    text = subprocess.check_output(["go", "list", "-m", "-json", "all"], text=True)
    decoder = json.JSONDecoder()
    modules, files = [], set()
    while text.strip():
        module, end = decoder.raw_decode(text.lstrip())
        text = text.lstrip()[end:]
        modules.append(module)
        escaped = "".join("!" + c.lower() if c.isupper() else c for c in module["Path"])
        directory = cache / escaped / "@v"
        files.update(directory.glob("*.mod"))
        files.update(directory.glob("*.info"))
        if "Version" in module:
            files.update(directory.glob(module["Version"] + ".zip*"))
    with tarfile.open(args.output / "module-proxy.tar.gz", "w:gz") as archive:
        for path in sorted(files):
            archive.add(path, arcname=path.relative_to(cache).as_posix(), recursive=False)
    print(f"Exported {len(sources)} source entries, {len(modules)} modules and {len(files)} cache files to {args.output}")


if __name__ == "__main__":
    main()
