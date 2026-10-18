"""Run Windows D-Bus units that do not require a native session/system daemon."""

from pathlib import Path
import re
import subprocess

ROOT = Path(__file__).resolve().parent.parent
FILES = (
    "codec_regression_test.go", "encoder_test.go", "decoder_test.go",
    "numeric_array_test.go", "decoder_fd_test.go", "store_test.go",
    "store_signature_test.go", "sig_test.go", "variant_test.go",
    "message_test.go", "match_test.go", "escape_test.go", "sequence_test.go",
    "property_context_test.go", "send_regression_test.go",
    "signal_subscription_test.go", "conn_windows_test.go", "transport_no_tcp_test.go",
)

names = []
for filename in FILES:
    names.extend(re.findall(r"^func (Test\w+)\(", (ROOT / "deps/dbus" / filename).read_text(encoding="utf-8"), re.M))
pattern = "^(?:" + "|".join(sorted(set(names))) + ")$"
for flags in ([], ["-tags", "godbus_no_tcp"]):
    subprocess.run(["go", "test", "-race", *flags, "-run", pattern, ".", "./portal"], cwd=ROOT / "deps/dbus", check=True)
    # The root names filter deliberately excludes portal tests; run its complete
    # platform-eligible unit suite separately. Real daemon cases are Linux-only.
    subprocess.run(["go", "test", "-race", *flags, "./portal"], cwd=ROOT / "deps/dbus", check=True)
print("PASS: Windows daemon-independent units, normal and godbus_no_tcp; native daemon interoperability not claimed")
