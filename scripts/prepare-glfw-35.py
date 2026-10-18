"""Build the v3.5 module from a verified GLFW 3.5.1 checkout.

Existing targets are never replaced. Wayland output may be reused only when
its input XML is identical; otherwise wayland-scanner is required.
"""

import argparse
import hashlib
from pathlib import Path
import shutil
import subprocess

ROOT = Path(__file__).resolve().parent.parent
REVISION = "d9d6f0f1f967807ffade6598ea9a631ebaf37a56"


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("upstream", type=Path)
    args = parser.parse_args()
    upstream = args.upstream.resolve()
    commit = subprocess.check_output(["git", "-C", str(upstream), "rev-parse", "HEAD"], text=True).strip()
    if commit != REVISION:
        raise SystemExit(f"Expected GLFW 3.5.1 {REVISION}, got {commit}")
    source = ROOT / "deps/glfw/v3.4/glfw"
    target = ROOT / "deps/glfw/v3.5/glfw"
    if target.exists():
        raise SystemExit(f"Target already exists: {target}")
    target.mkdir(parents=True)
    for path in source.iterdir():
        if path.is_file() and path.suffix in (".go", ".c", ".h", ".md", ".mod", ".sum"):
            text = path.read_text(encoding="utf-8").replace("github.com/Uzer007-admin/glfw/v3.4", "github.com/Uzer007-admin/glfw/v3.5")
            text = text.replace("glfw/src/cocoa_time.c", "glfw/src/macos_time.c")
            (target / path.name).write_text(text, encoding="utf-8", newline="\n")
    tree = target / "glfw"
    for name in ("src", "deps", "include"):
        shutil.copytree(upstream / name, tree / name)
    shutil.copyfile(upstream / "LICENSE.md", tree / "LICENSE.md")
    for path in (tree / "src").iterdir():
        if path.name == "CMakeLists.txt" or path.suffix == ".in":
            path.unlink()
    for protocol in sorted((tree / "deps/wayland").glob("*.xml")):
        old = source / "glfw/deps/wayland" / protocol.name
        for mode, suffix in (("client-header", "-client-protocol.h"), ("private-code", "-client-protocol-code.h")):
            output = tree / "include" / (protocol.stem + suffix)
            previous = source / "glfw/include" / output.name
            if old.exists() and old.read_bytes() == protocol.read_bytes() and previous.exists():
                shutil.copyfile(previous, output)
            else:
                subprocess.run(["wayland-scanner", mode, str(protocol), str(output)], check=True)
    dummy = '//go:build required\n\n// Package dummy keeps native source files in vendored modules.\npackage dummy\n'
    for name in ("src", "include", "deps"):
        for directory in [tree / name, *(p for p in (tree / name).rglob("*") if p.is_dir())]:
            (directory / "dummy.go").write_text(dummy, encoding="utf-8")
    children = sorted(p.name for p in (tree / "deps").iterdir() if p.is_dir())
    imports = ''.join(f'\t_ "github.com/Uzer007-admin/glfw/v3.5/glfw/glfw/deps/{name}"\n' for name in children)
    (tree / "deps/dummy.go").write_text(dummy + '\nimport (\n' + imports + ')\n', encoding="utf-8")
    (target / "GLFW_C_REVISION.txt").write_text(REVISION + "\n", encoding="utf-8")
    patch = (ROOT / "patches/glfw-frame-wait.patch").read_text(encoding="utf-8").replace("v3.4/", "v3.5/")
    # The cgo capability definition was already copied with the Go wrappers.
    sections = patch.split("diff --git ")
    patch = ''.join("diff --git " + s for s in sections[1:] if not s.startswith("a/v3.5/glfw/c_glfw_windows.go"))
    subprocess.run(["git", "-C", str(ROOT / "deps/glfw"), "apply", "--check", "-"], input=patch, text=True, check=True)
    subprocess.run(["git", "-C", str(ROOT / "deps/glfw"), "apply", "-"], input=patch, text=True, check=True)
    digest = hashlib.sha256()
    for path in sorted(p for p in tree.rglob("*") if p.is_file()):
        digest.update(path.relative_to(tree).as_posix().encode() + b'\0' + path.read_bytes())
    (target / "glfw_tree_rebuild.go").write_text('package glfw\n\n// Native source content hash invalidates cgo caches.\nconst upstreamTreeSHA = "' + digest.hexdigest() + '"\n', encoding="utf-8")
    print(f"Prepared {target} from {REVISION}")


if __name__ == "__main__":
    main()
