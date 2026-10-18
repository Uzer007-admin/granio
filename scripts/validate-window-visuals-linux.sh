#!/bin/sh
# Run from an exported public workspace. All caches and compositor state are
# confined to the supplied work directory; logs are retained for the caller.
set -eu
work=$(realpath "$1")
mkdir -p "$work/logs-window-visuals" "$work/cache" "$work/go-tmp" "$work/modules"
export GOCACHE="$work/cache" GOTMPDIR="$work/go-tmp" GOMODCACHE="$work/modules"
export GOPROXY="file://$work/proxy" GOSUMDB=off GOTOOLCHAIN=local GOEXPERIMENT=cgocheck2
export GOMAXPROCS=2 GOFLAGS=-p=1 FYNE_TEST_WINDOW_VISUALS=1
unset WAYLAND_DISPLAY
logs="$work/logs-window-visuals"
runtime=
weston_pid=
cleanup() {
    if [ -n "$weston_pid" ]; then kill "$weston_pid" 2>/dev/null || true; wait "$weston_pid" 2>/dev/null || true; fi
    case "$runtime" in "$work"/wv-runtime.*) rm -rf -- "$runtime" ;; esac
}
trap cleanup EXIT HUP INT TERM
run_check() {
    name=$1
    shift
    printf 'Checking %s\n' "$name"
    if "$@" > "$logs/$name.log" 2>&1; then
        printf 'PASS %s\n' "$name"
    else
        cat "$logs/$name.log"
        return 1
    fi
}
go version > "$logs/environment.log"
uname -a >> "$logs/environment.log"
weston --version >> "$logs/environment.log"
run_check x11 dbus-run-session -- xvfb-run -a go test -race -tags migrated_fynedo -run '^TestNativeWindowVisuals' -count=1 -v ./internal/driver/glfw
run_check gles2 dbus-run-session -- xvfb-run -a env MESA_GLES_VERSION_OVERRIDE=2.0 go test -race -tags migrated_fynedo,gles -run '^TestNativeWindowVisuals' -count=1 -v ./internal/driver/glfw
runtime=$(mktemp -d "$work/wv-runtime.XXXXXX")
chmod 700 "$runtime"
XDG_RUNTIME_DIR="$runtime" weston --backend=headless-backend.so --use-pixman --scale=2 --width=1280 --height=720 --idle-time=0 --no-config --socket=window-visuals --log="$logs/weston.log" > "$logs/weston-console.log" 2>&1 &
weston_pid=$!
i=0
while [ ! -S "$runtime/window-visuals" ]; do
    i=$((i+1))
    if ! kill -0 "$weston_pid" 2>/dev/null || [ "$i" -ge 100 ]; then cat "$logs/weston-console.log"; exit 1; fi
    sleep .05
done
run_check wayland dbus-run-session -- env -u DISPLAY XDG_RUNTIME_DIR="$runtime" WAYLAND_DISPLAY=window-visuals go test -race -tags migrated_fynedo,wayland -run '^TestNativeWindowVisuals' -count=1 -v ./internal/driver/glfw
run_check wayland-gles2 dbus-run-session -- env -u DISPLAY XDG_RUNTIME_DIR="$runtime" WAYLAND_DISPLAY=window-visuals MESA_GLES_VERSION_OVERRIDE=2.0 go test -race -tags migrated_fynedo,wayland,gles -run '^TestNativeWindowVisuals' -count=1 -v ./internal/driver/glfw
run_check root-race dbus-run-session -- xvfb-run -a env FYNE_TEST_WINDOW_VISUALS=0 go test -race -tags migrated_fynedo ./...
run_check gallery dbus-run-session -- xvfb-run -a go run -tags migrated_fynedo ./cmd/granio_gallery -smoke -backdrop acrylic
run_check wasm env GOOS=js GOARCH=wasm CGO_ENABLED=0 go build -tags migrated_fynedo -o "$work/visual-gallery.wasm" ./cmd/granio_gallery
