#!/bin/sh
set -eu
work=$(realpath "$1")
export XDG_CONFIG_HOME="$work/anthy-profile" XDG_CACHE_HOME="$work/anthy-cache" XDG_DATA_DIRS="$work/anthy/share:/usr/local/share:/usr/share"
export GI_TYPELIB_PATH="$work/anthy/lib/girepository-1.0" LD_LIBRARY_PATH="$work/anthy/lib" GSETTINGS_SCHEMA_DIR="$work/anthy/share/glib-2.0/schemas" GSETTINGS_BACKEND=keyfile
export LANG=C.UTF-8 LC_ALL=C.UTF-8 XMODIFIERS=@im=ibus FYNE_PLATFORM=x11 FYNE_LOG_IME=1
mkdir -p "$work/logs"
gsettings set org.freedesktop.ibus.engine.anthy.common input-mode 0
export IBUS_ADDRESS="unix:path=$work/ibus-test.sock"
rm -f "$work/ibus-test.sock"
ibus-daemon --address="$IBUS_ADDRESS" --panel=disable --config=/usr/libexec/ibus-memconf --verbose > "$work/logs/ibus-anthy.txt" 2>&1 & im_pid=$!
cleanup() {
    if [ -n "${probe_pid-}" ]; then
        kill "$probe_pid" 2>/dev/null || true
        i=0
        while kill -0 "$probe_pid" 2>/dev/null && [ "$i" -lt 30 ]; do sleep 0.1; i=$((i+1)); done
        kill -KILL "$probe_pid" 2>/dev/null || true
    fi
    kill "$im_pid" ${engine_pid-} ${xim_pid-} 2>/dev/null || true
    wait 2>/dev/null || true
}
trap cleanup EXIT HUP INT TERM
sleep 2
GDK_BACKEND=x11 /usr/libexec/ibus-x11 --server-name=ibus --locale-append=C.UTF-8,C.utf8,ru_RU.utf8 --debug=3 > "$work/logs/anthy-xim-server.txt" 2>&1 & xim_pid=$!
sleep 1
xprop -root _XIM_SERVERS > "$work/logs/anthy-xim-servers.txt"
"$work/anthy/libexec/ibus-engine-anthy" > "$work/logs/anthy-engine.txt" 2>&1 & engine_pid=$!
sleep 1
ibus list-engine > "$work/logs/ibus-engines.txt"
if [ "${FYNE_IME_BASELINE-}" = gtk ];then
    GDK_BACKEND=x11 GTK_IM_MODULE=xim python3 "$work/source/scripts/gtk-native-ime-probe.py" > "$work/logs/ime-anthy.jsonl" 2> "$work/logs/anthy-probe-stderr.txt" & probe_pid=$!
else
    "$work/input-probe" -seconds 30 > "$work/logs/ime-anthy.jsonl" 2> "$work/logs/anthy-probe-stderr.txt" & probe_pid=$!
fi
i=0;while ! grep -q ready "$work/logs/ime-anthy.jsonl";do i=$((i+1));[ "$i" -lt 100 ]||exit 1;sleep 0.1;done
title='^Fyne native input acceptance$';if [ "${FYNE_IME_BASELINE-}" = gtk ];then title='^GTK XIM input baseline$';fi
target=$(xdotool search --name "$title" | head -1);xdotool windowfocus --sync "$target"
ibus engine anthy
python3 - <<'PY'
import gi
gi.require_version('IBus','1.0')
from gi.repository import IBus
bus=IBus.Bus();print('IBus engine',bus.get_global_engine().get_name(),'enabled',bus.is_global_engine_enabled(),'context',bus.current_input_context())
context=IBus.InputContext.new(bus.current_input_context(),bus.get_connection(),None)
context.property_activate('InputMode.Hiragana',IBus.PropState.CHECKED)
bus.get_connection().flush_sync(None)
PY
sleep 0.5
xdotool type --delay 120 nihongo
sleep 1
xdotool key space
sleep 1
xdotool key Return
sleep 1
python3 - <<'PY'
import pyatspi
def visit(obj):
    yield obj
    if obj.getRole()==pyatspi.ROLE_LIST:return
    for child in obj:
        if child:yield from visit(child)
application=next(o for o in pyatspi.Registry.getDesktop(0) if o.name=='org.fyne.native-input-probe')
button=next(o for o in visit(application) if o.name=='Undo input')
assert button.queryAction().doAction(0)
PY
sleep 1
xdotool type --delay 120 nihon
sleep 1
xdotool key Escape Escape
sleep 1
cat "$work/logs/ime-anthy.jsonl"
python3 - "$work/logs/ime-anthy.jsonl" <<'PY'
import json,sys
events=[json.loads(s) for s in open(sys.argv[1]) if s.startswith('{')]
assert any(e.get('text')=='日本語' for e in events),events
assert any(e.get('composition') and e.get('text')=='' for e in events),events
assert any(e.get('event')=='undo' and e.get('text')=='' for e in events),events
assert events[-1].get('text')=='' and not events[-1].get('composition'),events[-1]
print('real IBus Anthy Japanese commit: PASS')
PY
if [ "${FYNE_TEST_REAL_IBUS-}" = 1 ]; then
    (cd "$work/source" && go test -p 1 -race -timeout 2m -tags accessibility,migrated_fynedo -count=1 -v -run '^TestIBusNativePasswordAndLifecycle$' ./internal/driver/glfw) > "$work/logs/ibus-privacy-lifetime.txt" 2>&1
    cat "$work/logs/ibus-privacy-lifetime.txt"
fi
