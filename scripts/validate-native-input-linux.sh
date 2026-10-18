#!/bin/sh
# Independent XIM/Orca processes and real XTest key events in an isolated session.
set -eu
work=$(realpath "$1")
export XDG_CONFIG_HOME="$work/profile" XDG_CACHE_HOME="$work/user-cache" GSETTINGS_BACKEND=memory
export LANG=C.UTF-8 LC_ALL=C.UTF-8 XMODIFIERS=@im=fcitx FYNE_PLATFORM=x11 FYNE_LOG_IME=1
mkdir -p "$work/profile/fcitx" "$work/orca" "$work/logs"
cat > "$work/profile/fcitx/profile" <<'EOF'
[Profile]
IMName=pinyin
EnabledIMList=fcitx-keyboard-us:True,pinyin:True
PreeditStringInClientWindow=True
EOF
cat > "$work/orca/user-settings.conf" <<'EOF'
{"general":{"startingProfile":["Default","default"],"activeProfile":["Default","default"],"enableSpeech":true,"enableBraille":false,"enableMouseReview":false},"profiles":{"default":{"profile":["Default","default"],"enableSpeech":true,"enableBraille":false}},"keybindings":{},"pronunciations":{}}
EOF
fcitx -D -s 0 > "$work/logs/fcitx.txt" 2>&1 & im_pid=$!
orca --user-prefs "$work/orca" --enable speech --disable braille --debug-file "$work/logs/orca-debug.txt" > "$work/logs/orca-console.txt" 2>&1 & reader_pid=$!
cleanup(){
    if [ -n "${probe_pid-}" ]; then kill "$probe_pid" 2>/dev/null || true; i=0; while kill -0 "$probe_pid" 2>/dev/null && [ "$i" -lt 30 ]; do sleep 0.1; i=$((i+1)); done; kill -KILL "$probe_pid" 2>/dev/null || true; fi
    kill "$im_pid" "$reader_pid" 2>/dev/null || true
    i=0
    while { kill -0 "$im_pid" 2>/dev/null || kill -0 "$reader_pid" 2>/dev/null; } && [ "$i" -lt 30 ]; do sleep 0.1; i=$((i+1)); done
    kill -KILL "$im_pid" "$reader_pid" 2>/dev/null || true
    wait 2>/dev/null || true
}
trap cleanup EXIT HUP INT TERM
sleep 2
"$work/input-probe" -seconds 45 > "$work/logs/ime-pinyin.jsonl" 2> "$work/logs/probe-stderr.txt" & probe_pid=$!
i=0
while ! grep -q '"event":"ready"' "$work/logs/ime-pinyin.jsonl"; do i=$((i+1)); [ "$i" -lt 100 ] || exit 1; kill -0 "$probe_pid"; sleep 0.1; done
target=$(xdotool search --name '^Fyne native input acceptance$' | head -1)
xdotool windowfocus --sync "$target"
fcitx-remote -s pinyin
fcitx-remote -o
xdotool type --delay 100 nihao
sleep 1
xdotool key space
sleep 1
xdotool key ctrl+z
sleep 1
xdotool type --delay 100 zhongguo
sleep 1
xdotool key space
sleep 1
xdotool key Tab Tab space
sleep 2
python3 "$work/source/scripts/atspi-reader-probe.py" > "$work/logs/reader-client.json" 2> "$work/logs/reader-client-stderr.txt"
sleep 2
grep -E 'SPEECH OUTPUT|SPEECH:|ERROR|Traceback' "$work/logs/orca-debug.txt" | tail -40 || true
cat "$work/logs/ime-pinyin.jsonl"
python3 - "$work/logs" <<'PY'
import json,sys
from pathlib import Path
logs=Path(sys.argv[1])
events=[json.loads(s) for s in (logs/'ime-pinyin.jsonl').read_text().splitlines() if s.startswith('{')]
assert any('你好' in e.get('text','') for e in events),events
assert any('中国' in e.get('text','') for e in events),events
reader=json.loads((logs/'reader-client.json').read_text())
assert reader['password_redacted'] and reader['child_count']==100000 and not reader['last_showing'],reader
speech=[s for s in (logs/'orca-debug.txt').read_text().splitlines() if 'SPEECH OUTPUT' in s]
assert any('CJK input' in s for s in speech),speech
print('real Fcitx Pinyin commits, Orca field speech and external AT-SPI offscreen/password checks: PASS')
PY
