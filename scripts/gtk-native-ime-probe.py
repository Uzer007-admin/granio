"""Native GTK/XIM baseline in the same private session as the Fyne probe."""
import gi
gi.require_version('Gtk','3.0')
from gi.repository import Gtk,GLib
import json
window=Gtk.Window(title='GTK XIM input baseline')
entry=Gtk.Entry();entry.connect('changed',lambda obj:print(json.dumps({'text':obj.get_text()},ensure_ascii=False),flush=True))
window.add(entry);window.set_default_size(400,160);window.show_all();entry.grab_focus()
GLib.timeout_add_seconds(30,lambda:Gtk.main_quit())
print('ready',flush=True)
Gtk.main()
