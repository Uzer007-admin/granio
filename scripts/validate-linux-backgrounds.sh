#!/bin/sh
# Exported workspace checks. Every compositor, socket, cache and binary belongs
# to the supplied temporary directory. No desktop session or package is changed.
set -eu
work=$(realpath "$1")
mkdir -p "$work/cache" "$work/go-tmp" "$work/modules" "$work/logs-backgrounds"
export GOCACHE="$work/cache" GOTMPDIR="$work/go-tmp" GOMODCACHE="$work/modules"
export GOPROXY="file://$work/proxy" GOSUMDB=off GOTOOLCHAIN=local GOEXPERIMENT=cgocheck2
export GOMAXPROCS=2 GOFLAGS=-p=1 FYNE_TEST_WINDOW_VISUALS=1
logs="$work/logs-backgrounds"
runtime=
compositor_pid=
cleanup() {
    if [ -n "$compositor_pid" ]; then kill "$compositor_pid" 2>/dev/null || true; wait "$compositor_pid" 2>/dev/null || true; fi
    case "$runtime" in "$work"/background-runtime.*) rm -rf -- "$runtime" ;; esac
}
trap cleanup EXIT HUP INT TERM
check() {
    name=$1; shift
    printf 'Checking %s\n' "$name"
    if "$@" > "$logs/$name.log" 2>&1; then printf 'PASS %s\n' "$name"; else cat "$logs/$name.log"; return 1; fi
}
if [ "${2:-}" = x11-fixture ]; then
    runtime=$(mktemp -d "$work/background-runtime.XXXXXX")
    export FYNE_TEST_BACKGROUND_SOCKET="$runtime/control" FYNE_TEST_XCOMPMGR="$work/native/usr/bin/xcompmgr"
    python3 scripts/x11-background-test-fixture.py > "$logs/x11-fixture-${3:-gl}.log" 2>&1 &
    compositor_pid=$!
    i=0
    while [ ! -S "$FYNE_TEST_BACKGROUND_SOCKET" ]; do
        i=$((i+1)); if ! kill -0 "$compositor_pid" 2>/dev/null || [ "$i" -ge 100 ]; then cat "$logs/x11-fixture-${3:-gl}.log"; exit 1; fi
        sleep .05
    done
    export FYNE_EXPECT_TRANSPARENT=1 FYNE_EXPECT_SYSTEM_BLUR=1
    tags=migrated_fynedo,x11
    if [ "${3:-}" = gles ]; then tags="$tags,gles"; export MESA_GLES_VERSION_OVERRIDE=2.0; fi
    go test -race -gcflags=all=-d=checkptr=2 -tags "$tags" -run '^Test(NativeWindowVisuals|WindowVisualsLinuxPolicy)' -count=1 -v ./internal/driver/glfw
    go test -race -tags x11,glfw_integration -run '^TestBackgroundNativeLifecycle$' -count=1 -v github.com/Uzer007-admin/glfw/v3.5/glfw
    exit
fi
go version > "$logs/environment.log"
uname -a >> "$logs/environment.log"
weston --version >> "$logs/environment.log"
unset WAYLAND_DISPLAY FYNE_TEST_BACKGROUND_SOCKET FYNE_EXPECT_TRANSPARENT FYNE_EXPECT_SYSTEM_BLUR
if [ "${2:-}" != wayland-only ]; then
    check x11-no-compositor dbus-run-session -- xvfb-run -a go test -race -tags migrated_fynedo,x11 -run '^Test(NativeWindowVisuals|WindowVisualsLinuxPolicy)' -count=1 -v ./internal/driver/glfw
    check x11 dbus-run-session -- xvfb-run -a sh scripts/validate-linux-backgrounds.sh "$work" x11-fixture gl
    check x11-gles dbus-run-session -- xvfb-run -a sh scripts/validate-linux-backgrounds.sh "$work" x11-fixture gles
fi
protocol=deps/glfw/v3.5/glfw/glfw/deps/wayland
mkdir -p "$work/protocol"
for name in ext-background-effect-v1 fractional-scale-v1; do
    wayland-scanner server-header "$protocol/$name.xml" "$work/protocol/$name-server-protocol.h"
    wayland-scanner private-code "$protocol/$name.xml" "$work/protocol/$name-server-code.c"
done
cc -Wall -Wextra -Wno-unused-parameter -shared -fPIC -I"$work/protocol" $(pkg-config --cflags libweston-10 wayland-server) \
    scripts/weston-background-test-module.c "$work/protocol/ext-background-effect-v1-server-code.c" "$work/protocol/fractional-scale-v1-server-code.c" \
    -o "$work/native/background-test.so" $(pkg-config --libs libweston-10 wayland-server)
runtime=$(mktemp -d "$work/background-runtime.XXXXXX"); chmod 700 "$runtime"
export XDG_RUNTIME_DIR="$runtime" WAYLAND_DISPLAY=background-test FYNE_EXPECT_TRANSPARENT=1
weston --backend=headless-backend.so --use-pixman --scale=2 --width=1280 --height=720 --idle-time=0 --no-config --socket=background-test --log="$logs/weston-standard.log" > "$logs/weston-standard-console.log" 2>&1 &
compositor_pid=$!
i=0
while [ ! -S "$runtime/background-test" ]; do i=$((i+1)); if ! kill -0 "$compositor_pid" 2>/dev/null || [ "$i" -ge 100 ]; then cat "$logs/weston-standard-console.log"; exit 1; fi; sleep .05; done
check wayland-no-blur dbus-run-session -- env -u DISPLAY go test -race -tags migrated_fynedo,wayland -run '^Test(NativeWindowVisuals|WindowVisualsLinuxPolicy)' -count=1 -v ./internal/driver/glfw
kill "$compositor_pid"; wait "$compositor_pid" 2>/dev/null || true; compositor_pid=
export FYNE_TEST_BACKGROUND_SOCKET="$runtime/control" FYNE_EXPECT_SYSTEM_BLUR=1
weston --backend=headless-backend.so --use-pixman --scale=2 --width=1280 --height=720 --idle-time=0 --no-config --socket=background-fixture --modules="$work/native/background-test.so" --log="$logs/weston-fixture.log" > "$logs/weston-fixture-console.log" 2>&1 &
compositor_pid=$!
export WAYLAND_DISPLAY=background-fixture
i=0
while [ ! -S "$runtime/background-fixture" ] || [ ! -S "$runtime/control" ]; do i=$((i+1)); if ! kill -0 "$compositor_pid" 2>/dev/null || [ "$i" -ge 100 ]; then cat "$logs/weston-fixture-console.log"; exit 1; fi; sleep .05; done
check wayland dbus-run-session -- env -u DISPLAY go test -race -gcflags=all=-d=checkptr=2 -tags migrated_fynedo,wayland -run '^Test(NativeWindowVisuals|WindowVisualsLinuxPolicy)' -count=1 -v ./internal/driver/glfw
check wayland-gles dbus-run-session -- env -u DISPLAY MESA_GLES_VERSION_OVERRIDE=2.0 go test -race -gcflags=all=-d=checkptr=2 -tags migrated_fynedo,wayland,gles -run '^Test(NativeWindowVisuals|WindowVisualsLinuxPolicy)' -count=1 -v ./internal/driver/glfw
check glfw-wayland dbus-run-session -- env -u DISPLAY go test -race -tags wayland,glfw_integration -run '^TestBackgroundNativeLifecycle$' -count=1 -v github.com/Uzer007-admin/glfw/v3.5/glfw
check gallery-transparent dbus-run-session -- env -u DISPLAY go run -tags migrated_fynedo,wayland ./cmd/granio_gallery -smoke -backdrop transparent
check gallery-blur dbus-run-session -- env -u DISPLAY go run -tags migrated_fynedo,wayland ./cmd/granio_gallery -smoke -backdrop system-blur
unset FYNE_TEST_BACKGROUND_SOCKET FYNE_EXPECT_SYSTEM_BLUR FYNE_EXPECT_TRANSPARENT WAYLAND_DISPLAY
export FYNE_TEST_WINDOW_VISUALS=0
check root-race dbus-run-session -- xvfb-run -a go test -race -tags migrated_fynedo ./...
check wasm env GOOS=js GOARCH=wasm CGO_ENABLED=0 go build -tags migrated_fynedo -o "$work/background-gallery.wasm" ./cmd/granio_gallery
