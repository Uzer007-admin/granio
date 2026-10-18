"""Export reviewable local fork changes against the pinned upstream commits."""

from pathlib import Path
import subprocess

ROOT = Path(__file__).resolve().parent.parent
FORKS = (("deps/gl", "gl-fork.patch"), ("deps/glfw", "glfw-fork.patch"), ("deps/glow", "glow-fork.patch"), ("deps/dbus", "dbus-fork.patch"), ("extras/apps", "apps-fork.patch"))


def main():
    for relative, filename in FORKS:
        module = ROOT / relative
        git = ["git", "-c", "core.safecrlf=false", "-C", str(module)]
        patch = subprocess.check_output(git + ["diff", "--binary", "HEAD"])
        entries = subprocess.check_output(git + ["ls-files", "--others", "--exclude-standard", "-z"])
        names = {value.decode("utf-8") for value in entries.split(b"\0") if value}
        if relative == "deps/glfw":
            # New native glue/header inputs can match upstream ignore rules.
            # Reproduction must include them even when Git omits them from status.
            directory = module / "v3.5/glfw"
            tracked = set(subprocess.check_output(git + ["ls-files", "-z"]).decode("utf-8").split("\0"))
            for path in directory.rglob("*"):
                if path.is_file() and path.suffix in (".c", ".h", ".m", ".mm", ".xml"):
                    name = path.relative_to(module).as_posix()
                    if name not in tracked:
                        names.add(name)
        for name in sorted(names):
            if (module / name).is_file():
                result = subprocess.run(git + ["diff", "--no-index", "--binary", "--", "/dev/null", name], capture_output=True)
                if result.returncode not in (0, 1):
                    raise RuntimeError(result.stderr.decode())
                patch += result.stdout
        (ROOT / "patches" / filename).write_bytes(patch)
        print(f"{filename}: {len(patch)} bytes")


if __name__ == "__main__":
    main()
