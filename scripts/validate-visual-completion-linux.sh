#!/bin/sh
# Short functional validation only. No benchmark/stress/performance loops.
set -eu
work=$(realpath "$1")
cd "$work/source"
mkdir -p "$work/logs" "$work/cache" "$work/tmp" "$work/modules" "$work/config" "$work/data"
export XDG_CONFIG_HOME="$work/config" XDG_DATA_HOME="$work/data" GOCACHE="$work/cache" GOTMPDIR="$work/tmp" GOMODCACHE="$work/modules"
export GOPROXY="file://$work/proxy" GOSUMDB=off GOTOOLCHAIN=local
export GOEXPERIMENT=cgocheck2 GOMAXPROCS=2 GOFLAGS=-p=2
export FYNE_TEST_WINDOW_VISUALS=0
unset WAYLAND_DISPLAY
run_check() {
    name=$1
    shift
    if "$@" > "$work/logs/$name.log" 2>&1; then printf 'PASS %s\n' "$name"; else cat "$work/logs/$name.log"; return 1; fi
}
go version > "$work/logs/environment.log"
uname -a >> "$work/logs/environment.log"
if [ "${2-}" != "after-root" ]; then
run_check dbus-portal dbus-run-session -- go test -race -count=1 -timeout 2m ./deps/dbus/...
run_check glfw-auto-fallback xvfb-run -a go test -race -tags glfw_integration -run '^TestAutoPlatformFallbackWithStaleWayland$' -count=1 -v ./deps/glfw/v3.5/glfw
run_check native-portal dbus-run-session -- xvfb-run -a go test -race -tags migrated_fynedo,system_appearance_integration -run '^TestNativeXDGAppearanceProviderLifecycle$' -count=1 -v ./app
run_check root-race dbus-run-session -- xvfb-run -a go test -p 1 -race -tags migrated_fynedo -timeout 4m ./...
else
run_check foldermeta-retest go test -race -tags migrated_fynedo -count=1 -v ./internal/foldermeta
run_check glfw-auto-fallback xvfb-run -a go test -race -tags glfw_integration -run '^TestAutoPlatformFallbackWithStaleWayland$' -count=1 -v ./deps/glfw/v3.5/glfw
fi
run_check reference-icons-software go test -race -tags migrated_fynedo,ci -run '^TestReference' -count=1 -v ./cmd/granio_gallery
run_check native-functional-x11 dbus-run-session -- xvfb-run -a env FYNE_TEST_TZ_FUNCTIONAL=1 go test -race -tags migrated_fynedo -run '^TestNative(ModernScopeVisibilityAndEscape|VisualHundredLifecycleAndResizeCycles|ModernLiveTextScalePreservesState)$' -count=1 -v -timeout 3m ./internal/driver/glfw
run_check materials-x11 dbus-run-session -- xvfb-run -a env FYNE_TEST_MATERIAL=1 go test -race -tags migrated_fynedo -run '^TestNativeCanvasAndPopupMaterials$' -count=1 -v ./internal/driver/glfw
run_check vet go vet -tags migrated_fynedo ./app ./theme ./widget ./canvas ./driver/desktop ./internal/animation ./internal/driver/glfw ./dialog ./cmd/granio_gallery
run_check gallery-x11 dbus-run-session -- xvfb-run -a go run -tags migrated_fynedo ./cmd/granio_gallery -smoke -profile modern -appearance-report "$work/logs/gallery-x11.json"
run_check gallery-grid-x11 dbus-run-session -- xvfb-run -a go run -tags migrated_fynedo ./cmd/granio_gallery -smoke -sample data-grid -capture-path "$work/logs/gallery-grid-x11.png" -appearance-report "$work/logs/gallery-grid-x11.json"
runtime=$(mktemp -d "$work/visual-runtime.XXXXXX")
chmod 700 "$runtime"
weston_pid=
cleanup() {
    if [ -n "$weston_pid" ]; then kill "$weston_pid" 2>/dev/null || true; wait "$weston_pid" 2>/dev/null || true; fi
    case "$runtime" in "$work"/visual-runtime.*) rm -rf -- "$runtime" ;; esac
}
trap cleanup EXIT HUP INT TERM
XDG_RUNTIME_DIR="$runtime" weston --backend=headless-backend.so --use-pixman --scale=2 --width=1280 --height=720 --idle-time=0 --no-config --socket=completion --log="$work/logs/weston.log" > "$work/logs/weston-console.log" 2>&1 &
weston_pid=$!
i=0
while [ ! -S "$runtime/completion" ]; do
    i=$((i+1))
    if ! kill -0 "$weston_pid" 2>/dev/null || [ "$i" -ge 100 ]; then cat "$work/logs/weston-console.log"; exit 1; fi
    sleep .05
done
run_check native-functional-wayland dbus-run-session -- env -u DISPLAY XDG_RUNTIME_DIR="$runtime" WAYLAND_DISPLAY=completion MESA_GLES_VERSION_OVERRIDE=2.0 FYNE_TEST_TZ_FUNCTIONAL=1 go test -race -tags migrated_fynedo,wayland,gles -run '^TestNative(ModernScopeVisibilityAndEscape|VisualHundredLifecycleAndResizeCycles|ModernLiveTextScalePreservesState)$' -count=1 -v -timeout 3m ./internal/driver/glfw
run_check gallery-wayland-gles2 dbus-run-session -- env -u DISPLAY XDG_RUNTIME_DIR="$runtime" WAYLAND_DISPLAY=completion MESA_GLES_VERSION_OVERRIDE=2.0 go run -tags migrated_fynedo,wayland,gles ./cmd/granio_gallery -smoke -profile materials -backdrop transparent -appearance-report "$work/logs/gallery-wayland-gles2.json"
run_check gallery-system-wayland dbus-run-session -- env -u DISPLAY XDG_RUNTIME_DIR="$runtime" WAYLAND_DISPLAY=completion FYNE_PLATFORM=wayland MESA_GLES_VERSION_OVERRIDE=2.0 go run -tags migrated_fynedo,gles ./cmd/granio_gallery -smoke -sample system -system-palette -appearance-report "$work/logs/gallery-system-wayland.json"
run_check wasm env GOOS=js GOARCH=wasm CGO_ENABLED=0 go build -tags migrated_fynedo -o "$work/gallery.wasm" ./cmd/granio_gallery
run_check macos-software-compile env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -tags migrated_fynedo,ci -o "$work/gallery-macos-software" ./cmd/granio_gallery
