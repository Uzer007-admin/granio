"""External libatspi client, also used while the real Orca process is running."""
import json
import pyatspi
import time

desktop = pyatspi.Registry.getDesktop(0)
application = None
for _ in range(50):
    for child in desktop:
        if child and child.name == "org.fyne.native-input-probe":
            application = child
            break
    if application:
        break
    time.sleep(0.1)
assert application is not None, "application is absent from the real AT-SPI registry"
window = application[0]
def visit(obj):
    yield obj
    if obj.getRole() == pyatspi.ROLE_LIST:
        return
    for child in obj:
        if child:
            yield from visit(child)
objects = list(visit(window))
entry = next(o for o in objects if o.name == "CJK input")
password = next(o for o in objects if o.name == "Private input")
collection = next(o for o in objects if o.getRole() == pyatspi.ROLE_LIST)
assert password.queryText().getText(0, -1) == "", "password contents exposed"
assert password.queryText().characterCount == 0, "password character count exposed"
assert collection.childCount == 100000
selection = collection.querySelection()
assert selection.nSelectedChildren == 1
selected = selection.getSelectedChild(0)
assert selected.name == "Item 90000 日本"
last = collection[99999]
assert last.name == "Item 99999 日本"
assert last.getIndexInParent() == 99999
assert not last.getState().contains(pyatspi.STATE_SHOWING)
print(json.dumps({"application": application.name, "window": window.name,
    "entry_text": entry.queryText().getText(0, -1), "password_redacted": True,
    "child_count": collection.childCount, "selected": selected.name,
    "last": last.name, "last_showing": False}, ensure_ascii=False))
