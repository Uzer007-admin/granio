#!/bin/sh
# Run from an exported workspace with sources/proxy unpacked in the supplied
# directory. Only this script's own compositor runtime is removed automatically.
set -eu
work=$(realpath "$1")
mkdir -p "$work/logs" "$work/cache" "$work/tmp" "$work/modules"
export GOCACHE="$work/cache" GOTMPDIR="$work/tmp" GOMODCACHE="$work/modules"
export GOPROXY="file://$work/proxy" GOSUMDB=off GOTOOLCHAIN=local GOEXPERIMENT=cgocheck2
unset WAYLAND_DISPLAY
runtime=
weston_pid=
cleanup() {
    if [ -n "$weston_pid" ]; then kill "$weston_pid" 2>/dev/null || true; wait "$weston_pid" 2>/dev/null || true; fi
    case "$runtime" in "$work"/wayland-runtime.*) rm -rf -- "$runtime" ;; esac
}
trap cleanup EXIT HUP INT TERM
go version > "$work/logs/environment.txt"
uname -a >> "$work/logs/environment.txt"
weston --version >> "$work/logs/environment.txt"
go test -race ./deps/glow ./deps/gl/capabilities/... ./deps/gl/v2.1/material ./deps/gl/v2.1/fyne ./deps/gl/v2.1/gl ./deps/gl/v2.0/gles2 > "$work/logs/unit.txt" 2>&1
printf '%s\n' 'GL/Glow capability and material-profile tests: PASS'
dbus-run-session -- sh -c 'set -e; export DBUS_SYSTEM_BUS_ADDRESS=$DBUS_SESSION_BUS_ADDRESS; go -C deps/dbus test -race ./...; go -C deps/dbus test -race -tags godbus_no_tcp ./...; go -C deps/dbus test -race -count=5 -run TestAppearanceWatcher ./portal' > "$work/logs/dbus.txt" 2>&1
printf '%s\n' 'D-Bus suites and repeated portal watcher tests: PASS'
dbus-run-session -- go -C deps/dbus test -race -v -run TestConnectToDifferentUserNamespace . > "$work/logs/namespace.txt" 2>&1
for version in 3.3 3.4 3.5; do xvfb-run -a go test -race -tags glfw_integration "./deps/glfw/v$version/glfw" > "$work/logs/glfw-$version.txt" 2>&1; done
xvfb-run -a go test -tags glfw_integration -run '^$' -bench 'Benchmark(Legacy)?SurfaceGeometry$' -benchmem -count=3 ./deps/glfw/v3.5/glfw > "$work/logs/geometry-benchmark.txt" 2>&1
go -C deps/dbus test -run '^$' -bench BenchmarkBoundedSignalDelivery -benchmem -count=3 . > "$work/logs/signal-benchmark.txt" 2>&1
python3 scripts/test-native-bridges.py --sanitize > "$work/logs/bridges.txt" 2>&1
python3 scripts/test-joystick-guid.py --sanitize > "$work/logs/guid.txt" 2>&1
printf '%s\n' 'GLFW geometry, signal delivery and C sanitizer tests: PASS'
xvfb-run -a go run -tags graphics_probe ./scripts/material-gl-probe/main.go -backend x11 > "$work/logs/material-x11.txt" 2>&1
xvfb-run -a env MESA_GLES_VERSION_OVERRIDE=2.0 go run -tags egl ./scripts/egl-probe/main.go > "$work/logs/gles20.txt" 2>&1
xvfb-run -a go test -race -tags migrated_fynedo ./app ./internal/painter/gl ./internal/driver/glfw > "$work/logs/fyne.txt" 2>&1
GOOS=js GOARCH=wasm CGO_ENABLED=0 go test -c -o "$work/capabilities.wasm" ./deps/gl/capabilities > "$work/logs/wasm-build.txt" 2>&1
printf '%s\n' 'Desktop/GLES texture-FBO alpha, Fyne and pure capability WASM build: PASS'
runtime=$(mktemp -d "$work/wayland-runtime.XXXXXX")
chmod 700 "$runtime"
XDG_RUNTIME_DIR="$runtime" weston --backend=headless-backend.so --use-pixman --scale=2 --width=1280 --height=720 --idle-time=0 --no-config --socket=library-next --log="$work/logs/weston.txt" > "$work/logs/weston-console.txt" 2>&1 &
weston_pid=$!
i=0
while [ ! -S "$runtime/library-next" ]; do
    if ! kill -0 "$weston_pid" 2>/dev/null || [ "$i" -ge 100 ]; then printf '%s\n' 'Weston failed to start' >&2; exit 1; fi
    sleep 0.1
    i=$((i+1))
done
env -u DISPLAY XDG_RUNTIME_DIR="$runtime" WAYLAND_DISPLAY=library-next go run -tags graphics_probe,wayland ./scripts/material-gl-probe/main.go -backend wayland > "$work/logs/material-wayland.txt" 2>&1
env -u DISPLAY XDG_RUNTIME_DIR="$runtime" WAYLAND_DISPLAY=library-next go run -tags graphics_probe,wayland ./scripts/surface-probe/main.go -backend wayland > "$work/logs/surface-wayland.txt" 2>&1
printf '%s\n' 'Wayland/EGL material FBO and surface metrics: PASS'
