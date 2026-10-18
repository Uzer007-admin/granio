//go:build accessibility && linux && !no_glfw && !mobile

package glfw

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Uzer007-admin/dbus/v5"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/require"
)

func TestATSPINativeOffscreenHierarchyAndCache(t *testing.T) {
	if os.Getenv("FYNE_TEST_ATSPI") != "1" {
		t.Skip("requires AT-SPI bus")
	}
	l := widget.NewList(func() int { return 100000 }, func() granio.CanvasObject { return widget.NewLabel("") }, func(id widget.ListItemID, object granio.CanvasObject) {
		object.(*widget.Label).SetText(fmt.Sprintf("Item %d 日本", id))
	})
	l.SelectionMode = widget.SelectionMultiple
	var w *window
	runOnMain(func() {
		w = d.CreateWindow("AT-SPI offscreen acceptance").(*window)
		w.create()
		w.SetContent(l)
		w.Resize(granio.NewSize(360, 180))
		w.Show()
		l.SetSelection([]int{90000})
	})
	t.Cleanup(func() { runOnMain(func() { w.Close() }) })
	m := linuxAccessibility
	require.Eventually(t, func() bool { m.mu.RLock(); defer m.mu.RUnlock(); return m.conn != nil }, 5*time.Second, 10*time.Millisecond)
	runOnMain(func() { w.updateAccessibility() })
	session, err := dbus.ConnectSessionBus()
	require.NoError(t, err)
	defer session.Close()
	var address string
	require.NoError(t, session.Object("org.a11y.Bus", "/org/a11y/bus").Call("org.a11y.Bus.GetAddress", 0).Store(&address))
	client, err := dbus.Connect(address)
	require.NoError(t, err)
	defer client.Close()
	path := m.pathForKey(w, accessibleKey{object: l})
	m.mu.RLock()
	bus := m.busName
	m.mu.RUnlock()
	object := client.Object(bus, path)
	var count dbus.Variant
	require.NoError(t, object.Call("org.freedesktop.DBus.Properties.Get", 0, atspiAccessible, "ChildCount").Store(&count))
	require.Equal(t, int32(100000), count.Value())
	var selected atspiReference
	require.NoError(t, object.Call(atspiSelection+".GetSelectedChild", 0, int32(0)).Store(&selected))
	var name dbus.Variant
	require.NoError(t, client.Object(bus, selected.Path).Call("org.freedesktop.DBus.Properties.Get", 0, atspiAccessible, "Name").Store(&name))
	require.Equal(t, "Item 90000 日本", name.Value())
	var last atspiReference
	require.NoError(t, object.Call(atspiAccessible+".GetChildAtIndex", 0, int32(99999)).Store(&last))
	var index int32
	require.NoError(t, client.Object(bus, last.Path).Call(atspiAccessible+".GetIndexInParent", 0).Store(&index))
	require.Equal(t, int32(99999), index)
	var states []uint32
	require.NoError(t, client.Object(bus, last.Path).Call(atspiAccessible+".GetState", 0).Store(&states))
	require.Zero(t, states[0]&(1<<25))
	var cache []atspiCacheItem
	require.NoError(t, client.Object(bus, atspiCachePath).Call(atspiCache+".GetItems", 0).Store(&cache))
	require.Less(t, len(cache), 200, "bulk cache must not realize 100000 rows")
	runOnMain(func() { l.ScrollTo(500); w.updateAccessibility() })
	var again atspiReference
	require.NoError(t, object.Call(atspiSelection+".GetSelectedChild", 0, int32(0)).Store(&again))
	require.Equal(t, selected.Path, again.Path)
}
