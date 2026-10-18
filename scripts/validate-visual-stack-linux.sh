#!/bin/sh
# Run from the exported workspace on f-test-bench. Logs remain in the supplied
# validation directory; the isolated compositor runtime is always removed.
set -eu
work=$(realpath "$1")
mkdir -p "$work/logs" "$work/cache" "$work/tmp"
export GOCACHE="$work/cache" GOTMPDIR="$work/tmp" GOTOOLCHAIN=local
export GOPROXY="file://$work/proxy" GOSUMDB=off GOEXPERIMENT=cgocheck2
unset WAYLAND_DISPLAY
runtime=
weston_pid=
cleanup() {
    if [ -n "$weston_pid" ]; then
        kill "$weston_pid" 2>/dev/null || true
        wait "$weston_pid" 2>/dev/null || true
    fi
    case "$runtime" in
        "$work"/wayland-runtime.*) rm -rf -- "$runtime" ;;
    esac
}
trap cleanup EXIT HUP INT TERM

go version > "$work/logs/environment.txt"
uname -a >> "$work/logs/environment.txt"
weston --version >> "$work/logs/environment.txt"
go test -race ./deps/glow ./deps/gl/v2.1/gl ./deps/gl/v2.1/fyne ./deps/gl/v2.0/gles2 > "$work/logs/generated-tests.txt" 2>&1
printf '%s\n' 'Generated packages: PASS'
dbus-run-session -- sh -c 'set -e; export DBUS_SYSTEM_BUS_ADDRESS=$DBUS_SESSION_BUS_ADDRESS; go -C deps/dbus test -race ./...; go -C deps/dbus test -race -tags godbus_no_tcp ./...' > "$work/logs/dbus-tests.txt" 2>&1
dbus-run-session -- go -C deps/dbus test -race -v -run 'TestConnectToDifferentUserNamespace' . > "$work/logs/namespace-gate.txt" 2>&1
printf '%s\n' 'D-Bus daemon suites and portal: PASS'
for version in 3.3 3.4 3.5; do
    xvfb-run -a go test -race -tags glfw_integration "./deps/glfw/v$version/glfw" > "$work/logs/glfw-$version.txt" 2>&1
done
xvfb-run -a go test -tags glfw_integration -run '^$' -bench 'Benchmark(Legacy)?SurfaceMetrics$' -benchmem -count=3 ./deps/glfw/v3.5/glfw > "$work/logs/surface-benchmark.txt" 2>&1
printf '%s\n' 'X11 surfaces and benchmarks: PASS'
python3 scripts/test-native-bridges.py --sanitize > "$work/logs/c-bridges.txt" 2>&1
python3 scripts/test-joystick-guid.py --sanitize > "$work/logs/c-guid.txt" 2>&1
xvfb-run -a go run -tags graphics_probe ./scripts/compact-gl-probe/main.go > "$work/logs/compact-gl.txt" 2>&1
xvfb-run -a go run -tags egl ./scripts/egl-probe/main.go > "$work/logs/egl.txt" 2>&1
xvfb-run -a env MESA_GLES_VERSION_OVERRIDE=2.0 go run -tags egl ./scripts/egl-probe/main.go > "$work/logs/gles20.txt" 2>&1
xvfb-run -a go run -tags graphics_probe ./scripts/surface-probe/main.go -backend x11 > "$work/logs/surface-x11.txt" 2>&1
xvfb-run -a go test -race -tags migrated_fynedo ./app ./internal/painter/gl ./internal/driver/glfw > "$work/logs/fyne-integration.txt" 2>&1
printf '%s\n' 'C sanitizers, GL/GLES/EGL and Fyne integration: PASS'

runtime=$(mktemp -d "$work/wayland-runtime.XXXXXX")
chmod 700 "$runtime"
XDG_RUNTIME_DIR="$runtime" weston --backend=headless-backend.so --use-pixman --scale=2 --width=1280 --height=720 --idle-time=0 --no-config --socket=visual-stack --log="$work/logs/weston.txt" > "$work/logs/weston-console.txt" 2>&1 &
weston_pid=$!
i=0
while [ ! -S "$runtime/visual-stack" ]; do
    if ! kill -0 "$weston_pid" 2>/dev/null || [ "$i" -ge 100 ]; then
        printf '%s\n' 'Weston failed to start' >&2
        exit 1
    fi
    sleep 0.1
    i=$((i + 1))
done
env -u DISPLAY XDG_RUNTIME_DIR="$runtime" WAYLAND_DISPLAY=visual-stack go run -tags graphics_probe,wayland ./scripts/surface-probe/main.go -backend wayland > "$work/logs/surface-wayland.txt" 2>&1
printf '%s\n' 'Wayland/EGL surfaces on Weston at output scale 2: PASS'
