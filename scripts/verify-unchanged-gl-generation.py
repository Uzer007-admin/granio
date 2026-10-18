"""Verify unchanged regeneration preserves all generated/native input timestamps."""

import hashlib
import json
from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parent.parent


def snapshot(directory):
    return {path.relative_to(directory).as_posix(): (hashlib.sha256(path.read_bytes()).hexdigest(), path.stat().st_mtime_ns)
            for path in directory.rglob("*") if path.is_file()}


def main():
    results = []
    for profile, restriction, flags in (("fyne", "fyne-gl21.json", ["-remext=.*"]), ("material", "material-gl21.json", [])):
        directory = ROOT / "deps/gl/v2.1" / profile
        before = snapshot(directory)
        subprocess.run(["go", "run", "../glow/", "generate", "-dispatch", "-api=gl", "-version=2.1",
                        "-out=./v2.1/" + profile, "-restrict=./profiles/" + restriction,
                        "-xml=../glow/xml", "-tmpl=../glow/tmpl", *flags], cwd=ROOT / "deps/gl", check=True)
        after = snapshot(directory)
        if before != after:
            raise ValueError(f"Regeneration changed content or mtime: {profile}: {[name for name in set(before) | set(after) if before.get(name) != after.get(name)]}")
        results.append({"profile": profile, "files": len(before), "content_identical": True, "mtime_identical": True})
        print(f"PASS {profile}: {len(before)} files retained content and timestamps")
    (ROOT / "docs/graphics-assessment/unchanged-generation.json").write_text(json.dumps(results, indent=2) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
