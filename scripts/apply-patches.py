"""Replay local submodule changes without resetting an existing checkout."""

from pathlib import Path
import subprocess


ROOT = Path(__file__).resolve().parent.parent
PATCHES = (
    ("extras/apps", "apps-fork.patch"),
    ("deps/gl", "gl-fork.patch"),
    ("deps/glfw", "glfw-fork.patch"),
    ("deps/glow", "glow-fork.patch"),
    ("deps/dbus", "dbus-fork.patch"),
)


def main():
    for directory, filename in PATCHES:
        module = ROOT / directory / "go.mod"
        if module.is_file() and module.read_text(encoding="utf-8").startswith("module github.com/Uzer007-admin/"):
            print(f"Исходники Granio уже закреплены в подмодуле: {directory}")
            continue
        patch = ROOT / "patches" / filename
        if not patch.is_file():
            raise SystemExit(f"Missing patch: {patch}")
        command = ["git", "-C", str(ROOT / directory), "apply"]
        if subprocess.run(command + ["--reverse", "--check", str(patch)], capture_output=True).returncode == 0:
            print(f"Already applied: {filename}")
            continue
        subprocess.run(command + ["--check", str(patch)], check=True)
        subprocess.run(command + [str(patch)], check=True)
        print(f"Applied: {filename}")


if __name__ == "__main__":
    main()
