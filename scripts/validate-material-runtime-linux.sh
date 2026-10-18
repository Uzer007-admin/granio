#!/bin/sh
# Exported-workspace checks; each writable cache/runtime belongs to this run.
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
if [ "${2-}" != --resume-renderers ] && [ "${2-}" != --gpu-only ]; then
go test -race ./deps/glow ./deps/gl/material ./deps/gl/capabilities/... ./deps/gl/v2.1/material ./deps/gl/v2.1/fyne ./deps/gl/v2.0/gles2 > "$work/logs/unit.txt" 2>&1
go vet ./deps/glow ./deps/gl/material ./deps/gl/capabilities/... ./deps/dbus/portal > "$work/logs/vet.txt" 2>&1
printf '%s\n' 'Material/dispatch/portal units and handwritten vet: PASS'
dbus-run-session -- sh -c 'set -e; export DBUS_SYSTEM_BUS_ADDRESS=$DBUS_SESSION_BUS_ADDRESS; go -C deps/dbus test -race ./...; go -C deps/dbus test -race -tags godbus_no_tcp ./...; go -C deps/dbus test -race -count=10 -run TestAppearanceWatcher ./portal' > "$work/logs/dbus.txt" 2>&1
printf '%s\n' 'D-Bus normal/no-TCP suites, owner/retry/shutdown repeated: PASS'
dbus-run-session -- xvfb-run -a go test -race -count=5 -tags migrated_fynedo,system_appearance_integration -run TestNativeXDGAppearanceProviderLifecycle ./app > "$work/logs/appearance-fyne.txt" 2>&1
for version in 3.3 3.4 3.5; do xvfb-run -a go test -race -tags glfw_integration "./deps/glfw/v$version/glfw" > "$work/logs/glfw-$version.txt" 2>&1; done
printf '%s\n' 'App-owned portal lifecycle and GLFW platform fallback/lifecycle: PASS'
xvfb-run -a go run -race -gcflags=all=-d=checkptr=2 -tags graphics_probe ./scripts/material-runtime-probe/main.go -backend x11 > "$work/logs/material-x11.txt" 2>&1
xvfb-run -a env MESA_GLES_VERSION_OVERRIDE=2.0 go run -race -gcflags=all=-d=checkptr=2 -tags graphics_probe,egl ./scripts/material-runtime-probe/main.go -backend x11 -gles2 > "$work/logs/material-gles20.txt" 2>&1
xvfb-run -a env MESA_GL_VERSION_OVERRIDE=2.1 MESA_EXTENSION_OVERRIDE=-GL_ARB_framebuffer_object go run -race -gcflags=all=-d=checkptr=2 -tags graphics_probe ./scripts/material-runtime-probe/main.go -backend x11 > "$work/logs/material-ext21.txt" 2>&1
go test -run '^$' -bench BenchmarkCachedMaterial -benchmem -count=3 ./deps/gl/material > "$work/logs/cache-benchmark.txt" 2>&1
fi
if [ "${2-}" = --gpu-only ]; then
    go test -race ./deps/gl/material > "$work/logs/material-final-unit.txt" 2>&1
    go vet ./deps/gl/material > "$work/logs/material-final-vet.txt" 2>&1
    xvfb-run -a go run -race -gcflags=all=-d=checkptr=2 -tags graphics_probe ./scripts/material-runtime-probe/main.go -backend x11 > "$work/logs/material-x11.txt" 2>&1
    xvfb-run -a env MESA_GLES_VERSION_OVERRIDE=2.0 go run -race -gcflags=all=-d=checkptr=2 -tags graphics_probe,egl ./scripts/material-runtime-probe/main.go -backend x11 -gles2 > "$work/logs/material-gles20.txt" 2>&1
    xvfb-run -a env MESA_GL_VERSION_OVERRIDE=2.1 MESA_EXTENSION_OVERRIDE=-GL_ARB_framebuffer_object go run -race -gcflags=all=-d=checkptr=2 -tags graphics_probe ./scripts/material-runtime-probe/main.go -backend x11 > "$work/logs/material-ext21.txt" 2>&1
fi
if [ "${2-}" != --gpu-only ]; then
xvfb-run -a go test -race -tags migrated_fynedo ./app ./internal/painter/gl ./internal/driver/glfw > "$work/logs/fyne.txt" 2>&1
GOOS=js GOARCH=wasm CGO_ENABLED=0 go test -c -o "$work/material.wasm" ./deps/gl/material > "$work/logs/wasm-build.txt" 2>&1
printf '%s\n' 'Two-context desktop/GLES GPU blur, cache, alpha and Fyne: PASS'
fi
runtime=$(mktemp -d "$work/wayland-runtime.XXXXXX")
chmod 700 "$runtime"
XDG_RUNTIME_DIR="$runtime" weston --backend=headless-backend.so --use-pixman --scale=2 --width=1280 --height=720 --idle-time=0 --no-config --socket=material-runtime --log="$work/logs/weston.txt" > "$work/logs/weston-console.txt" 2>&1 &
weston_pid=$!
i=0
while [ ! -S "$runtime/material-runtime" ]; do
    if ! kill -0 "$weston_pid" 2>/dev/null || [ "$i" -ge 100 ]; then printf '%s\n' 'Weston failed to start' >&2; exit 1; fi
    sleep 0.1
    i=$((i+1))
done
env -u DISPLAY XDG_RUNTIME_DIR="$runtime" WAYLAND_DISPLAY=material-runtime go run -race -gcflags=all=-d=checkptr=2 -tags graphics_probe,wayland ./scripts/material-runtime-probe/main.go -backend wayland > "$work/logs/material-wayland.txt" 2>&1
env -u DISPLAY XDG_RUNTIME_DIR="$runtime" WAYLAND_DISPLAY=material-runtime MESA_GLES_VERSION_OVERRIDE=2.0 go run -race -gcflags=all=-d=checkptr=2 -tags graphics_probe,egl,wayland ./scripts/material-runtime-probe/main.go -backend wayland -gles2 > "$work/logs/material-wayland-gles20.txt" 2>&1
printf '%s\n' 'Wayland/EGL two-context GL/GLES blur: PASS'
