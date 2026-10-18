"""Owned Xvfb fixture: real xcompmgr alpha plus KWin property/capability tests.

It advertises KWin's property solely for protocol/lifecycle tests. xcompmgr does
not implement native background blur. Never run this against a user's desktop.
"""
import ctypes as c
import os
from pathlib import Path
import signal
import socket
import subprocess

lib = c.CDLL("libX11.so.6")
def bind(name, result, *args):
    function = getattr(lib, name)
    function.restype = result
    function.argtypes = list(args)
    return function
open_display = bind("XOpenDisplay", c.c_void_p, c.c_char_p)
root_window = bind("XDefaultRootWindow", c.c_ulong, c.c_void_p)
intern = bind("XInternAtom", c.c_ulong, c.c_void_p, c.c_char_p, c.c_int)
change = bind("XChangeProperty", c.c_int, c.c_void_p, c.c_ulong, c.c_ulong, c.c_ulong, c.c_int, c.c_int, c.c_void_p, c.c_int)
delete = bind("XDeleteProperty", c.c_int, c.c_void_p, c.c_ulong, c.c_ulong)
flush = bind("XSync", c.c_int, c.c_void_p, c.c_int)
display = open_display(None)
if not display:
    raise SystemExit("owned Xvfb display unavailable")
root = root_window(display)
atom = intern(display, b"_KDE_NET_WM_BLUR_BEHIND_REGION", 0)
cardinal = intern(display, b"CARDINAL", 0)
control_path = os.environ["FYNE_TEST_BACKGROUND_SOCKET"]
compositor_path = os.environ["FYNE_TEST_XCOMPMGR"]
child = None
def advertise(enabled):
    if enabled:
        change(display, root, atom, cardinal, 32, 0, None, 0)
    else:
        delete(display, root, atom)
    flush(display, 0)
def start():
    global child
    child = subprocess.Popen([compositor_path, "-a"])
    advertise(True)
def stop():
    global child
    advertise(False)
    if child is not None:
        child.terminate()
        child.wait(timeout=10)
        child = None
def terminated(signum, frame):
    raise SystemExit(0)
signal.signal(signal.SIGTERM, terminated)
signal.signal(signal.SIGINT, terminated)
control = socket.socket(socket.AF_UNIX, socket.SOCK_DGRAM)
try:
    start()
    control.bind(control_path)
    print("owned X11 background fixture ready (xcompmgr alpha, KWin protocol fixture)", flush=True)
    while True:
        command = control.recv(80).decode()
        print(f"fixture command {command}", flush=True)
        if command == "blur off": advertise(False)
        elif command == "blur on": advertise(True)
        elif command == "remove": stop()
        elif command == "add": start()
finally:
    stop()
    control.close()
    Path(control_path).unlink(missing_ok=True)
