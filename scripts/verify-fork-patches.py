"""Replay exported fork patches against clean pinned upstream archives."""

import argparse
import hashlib
import json
from pathlib import Path
import runpy
import subprocess
import tarfile
import tempfile


ROOT = Path(__file__).resolve().parent.parent
FORKS = runpy.run_path(str(ROOT / "scripts/export-fork-patches.py"))["FORKS"]


def git(directory, *arguments):
    return subprocess.check_output(["git", "-C", str(directory), *arguments], stderr=subprocess.PIPE)


def normalized(data):
    # Git on Windows normalizes text line endings; binary data stays exact.
    return data if b"\0" in data else data.replace(b"\r\n", b"\n")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--temp-root", type=Path)
    parser.add_argument("--output", type=Path, default=ROOT / "docs/graphics-assessment/fork-patch-replay.json")
    args = parser.parse_args()
    temp_root = (args.temp_root or Path(tempfile.gettempdir())).resolve()
    results = []
    with tempfile.TemporaryDirectory(prefix="fork-replay-", dir=temp_root) as temporary:
        directory = Path(temporary).resolve()
        if not directory.is_relative_to(temp_root):
            raise ValueError("Temporary directory escaped its root")
        for relative, filename in FORKS:
            source = ROOT / relative
            target = directory / source.name
            target.mkdir()
            archive = directory / (source.name + ".tar")
            with archive.open("wb") as out:
                subprocess.run(["git", "-C", str(source), "archive", "--format=tar", "HEAD"], stdout=out, check=True)
            with tarfile.open(archive) as tar:
                tar.extractall(target, filter="data")
            # Without its own repository, git apply can discover the parent
            # checkout and silently skip paths outside that checkout's prefix.
            git(target, "init", "--quiet")
            patch = ROOT / "patches" / filename
            git(target, "apply", "--check", str(patch))
            git(target, "apply", "--binary", str(patch))
            entries = git(source, "ls-files", "--cached", "--others", "--exclude-standard", "-z")
            expected = {name.decode("utf-8") for name in entries.split(b"\0") if name and (source / name.decode("utf-8")).is_file()}
            actual = {path.relative_to(target).as_posix() for path in target.rglob("*") if path.is_file() and ".git" not in path.relative_to(target).parts}
            if expected != actual:
                raise ValueError(f"{filename}: missing {sorted(expected-actual)}, extra {sorted(actual-expected)}")
            for name in sorted(expected):
                if normalized((source / name).read_bytes()) != normalized((target / name).read_bytes()):
                    raise ValueError(f"{filename}: content differs: {name}")
            # Reapplying must be recognized as already applied by apply-patches.py.
            git(target, "apply", "--reverse", "--check", str(patch))
            results.append({"patch": filename, "base": git(source, "rev-parse", "HEAD").decode().strip(), "sha256": hashlib.sha256(patch.read_bytes()).hexdigest(), "identical": True, "files": len(expected)})
            print(f"PASS {filename}: {len(expected)} matching files")
    args.output.write_text(json.dumps(results, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
