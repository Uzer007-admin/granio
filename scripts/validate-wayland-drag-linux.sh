#!/bin/sh
set -eu
work=$(realpath "$1")
export XDG_RUNTIME_DIR="$work/wayland-runtime" WAYLAND_DISPLAY=fyne-drag-test
mkdir -p "$XDG_RUNTIME_DIR" "$work/logs/wayland"
chmod 700 "$XDG_RUNTIME_DIR"
protocol="$work/source/deps/glfw/v3.5/glfw/glfw/deps/wayland/xdg-shell.xml"
wayland-scanner client-header "$protocol" "$work/xdg-shell-client-protocol.h"
wayland-scanner private-code "$protocol" "$work/xdg-shell-protocol.c"
cc -O2 -I"$work" "$work/source/scripts/native-wayland-drag-source.c" "$work/xdg-shell-protocol.c" $(pkg-config --cflags --libs wayland-client) -o "$work/wayland-source"
cc -O2 "$work/source/scripts/native-drag-pointer.c" -lX11 -lXtst -o "$work/drag-pointer"
cc -shared -fPIC -O2 "$work/source/scripts/weston-input-test-module.c" $(pkg-config --cflags --libs libweston-10 wayland-server) -o "$work/weston-input-test.so"
export FYNE_WESTON_INPUT_SOCKET="$work/input-test.sock"
rm -f "$FYNE_WESTON_INPUT_SOCKET"
weston --backend=x11-backend.so --modules="$work/weston-input-test.so" --use-pixman --width=1000 --height=750 --idle-time=0 --no-config --socket=fyne-drag-test --log="$work/logs/wayland/weston.txt" > "$work/logs/wayland/weston-console.txt" 2>&1 & weston_pid=$!
cleanup(){ if [ -n "${probe_pid-}" ]; then kill "$probe_pid" 2>/dev/null || true;i=0;while kill -0 "$probe_pid" 2>/dev/null && [ "$i" -lt 30 ]; do sleep 0.1;i=$((i+1));done;kill -KILL "$probe_pid" 2>/dev/null || true;fi;kill ${source_pid-} "$weston_pid" 2>/dev/null || true;wait 2>/dev/null || true; }
trap cleanup EXIT HUP INT TERM
i=0;while [ ! -S "$XDG_RUNTIME_DIR/$WAYLAND_DISPLAY" ]; do i=$((i+1));[ "$i" -lt 100 ] || exit 1;sleep 0.1;done
WAYLAND_DEBUG=1 FYNE_PLATFORM=wayland "$work/input-probe-wayland" -drop-target -seconds 90 > "$work/logs/wayland/received.jsonl" 2> "$work/logs/wayland/target-stderr.txt" & probe_pid=$!
i=0;while ! grep -q '"event":"ready"' "$work/logs/wayland/received.jsonl"; do i=$((i+1));[ "$i" -lt 100 ] || exit 1;sleep 0.1;done
python3 - "$work" <<'PY'
from pathlib import Path
import sys
root=Path(sys.argv[1]); (root/'unicode.payload').write_bytes('Привет 日本 😀'.encode());(root/'html.payload').write_bytes('<b>日本 😀</b>'.encode());(root/'files.payload').write_bytes(b'file:///tmp/example%20file.txt\r\n');(root/'binary.payload').write_bytes(bytes([0,7,255,17])*131072)
PY
for kind in unicode html files binary; do
    case "$kind" in unicode) mime=text/plain;;html)mime=text/html;;files)mime=text/uri-list;;binary)mime=application/x-fyne-drag-test;;esac
    "$work/wayland-source" "$mime" "$work/$kind.payload" > "$work/logs/wayland/source-$kind.txt" 2>&1 & source_pid=$!
    i=0;while ! grep -q ready "$work/logs/wayland/source-$kind.txt";do i=$((i+1));[ "$i" -lt 100 ] || exit 1;kill -0 "$source_pid";sleep 0.1;done
    sleep 1
    GDK_BACKEND=x11 python3 - "$work" <<'PY'
import gi,sys
gi.require_version('Gdk','3.0')
from gi.repository import Gdk
root=Gdk.get_default_root_window();width,height=root.get_width(),root.get_height();pixels=Gdk.pixbuf_get_from_window(root,0,0,width,height);pixels.savev(sys.argv[1]+'/logs/wayland/compositor.png','png',[],[])
PY
    "$work/drag-pointer" > "$work/logs/wayland/gesture-$kind.txt"
    wait "$source_pid";source_pid=
    grep -q finished "$work/logs/wayland/source-$kind.txt"
    sleep 0.3
done
"$work/wayland-source" application/x-not-accepted "$work/unicode.payload" > "$work/logs/wayland/source-rejected.txt" 2>&1 & source_pid=$!
i=0;while ! grep -q ready "$work/logs/wayland/source-rejected.txt";do i=$((i+1));[ "$i" -lt 100 ] || exit 1;sleep 0.1;done
sleep 1
"$work/drag-pointer" > "$work/logs/wayland/gesture-rejected.txt"
if wait "$source_pid";then printf '%s\n' 'unsupported MIME unexpectedly accepted' >&2;exit 1;else result=$?;[ "$result" -eq 8 ] || exit "$result";fi
source_pid=
grep -q cancelled "$work/logs/wayland/source-rejected.txt"
python3 - "$work" <<'PY'
import base64,json,sys
from pathlib import Path
root=Path(sys.argv[1]);events=[json.loads(s) for s in (root/'logs/wayland/received.jsonl').read_text().splitlines() if s.startswith('{')];drops=[e['data'] for e in events if e.get('event')=='drop'];assert len(drops)==4,drops
for data,kind,mime in zip(drops,('unicode','html','files','binary'),('text/plain','text/html','text/uri-list','application/x-fyne-drag-test')):
    actual=base64.b64decode(data[mime]);assert actual==(root/(kind+'.payload')).read_bytes();print(f'external Wayland {kind}: {len(actual)} bytes, source finished, PASS')
print('external Wayland unsupported MIME: rejected, PASS')
PY
if [ "${FYNE_TEST_MATERIAL-}" = 1 ]; then
    (cd "$work/source" && FYNE_PLATFORM=wayland FYNE_TEST_ATSPI=1 go test -p 2 -race -timeout 3m -tags wayland,accessibility,migrated_fynedo -count=1 -v -run '^TestNativeCanvasAndPopupMaterials$|^TestATSPINativeOffscreenHierarchyAndCache$|^TestATSPINativeClient$' ./internal/driver/glfw) > "$work/logs/material-wayland.txt" 2>&1
    cat "$work/logs/material-wayland.txt"
fi
