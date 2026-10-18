//go:build accessibility && linux && !no_glfw && !mobile

package glfw

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Uzer007-admin/dbus/v5"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/require"
)

func TestATSPINativeCollectionAndValuePatterns(t *testing.T) {
	if os.Getenv("FYNE_TEST_ATSPI") != "1" {
		t.Skip("requires the native AT-SPI accessibility bus")
	}
	list := widget.NewList(func() int { return 1000 }, func() granio.CanvasObject { return widget.NewLabel("") },
		func(id widget.ListItemID, object granio.CanvasObject) {
			object.(*widget.Label).SetText(fmt.Sprintf("Item %d", id))
		})
	list.SelectionMode = widget.SelectionMultiple
	table := widget.NewTable(func() (int, int) { return 8, 3 }, func() granio.CanvasObject { return widget.NewLabel("cell") },
		func(id widget.TableCellID, object granio.CanvasObject) {
			object.(*widget.Label).SetText(fmt.Sprintf("%d:%d", id.Row, id.Col))
		})
	table.SelectionMode = widget.SelectionMultiple
	slider := widget.NewSlider(0, 10)
	var win *window
	runOnMain(func() {
		win = d.CreateWindow("AT-SPI pattern fixture").(*window)
		win.create()
		win.SetContent(container.NewBorder(nil, slider, nil, nil, container.NewGridWithColumns(2, list, table)))
		win.Resize(granio.NewSize(600, 400))
		win.Show()
		list.ScrollTo(90)
		table.SetSelection([]widget.TableCellID{{Row: 1, Col: 0}, {Row: 1, Col: 1}, {Row: 1, Col: 2}})
		win.updateAccessibility()
	})
	t.Cleanup(func() { runOnMain(func() { win.Close() }) })
	m := linuxAccessibility
	require.Eventually(t, func() bool { m.mu.RLock(); defer m.mu.RUnlock(); return m.conn != nil }, 5*time.Second, 10*time.Millisecond)
	runOnMain(func() { win.updateAccessibility() })
	session, err := dbus.ConnectSessionBus()
	require.NoError(t, err)
	defer session.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var address string
	require.NoError(t, session.Object("org.a11y.Bus", "/org/a11y/bus").CallWithContext(ctx, "org.a11y.Bus.GetAddress", 0).Store(&address))
	client, err := dbus.Connect(address)
	require.NoError(t, err)
	defer client.Close()
	paths := make(map[granio.CanvasObject]dbus.ObjectPath)
	m.mu.RLock()
	bus := m.busName
	for _, path := range m.windows[win] {
		paths[m.elements[path].node.object] = path
	}
	listElement := m.elements[paths[list]]
	childIndex, logicalIndex := -1, -1
	for index, path := range listElement.children {
		node := m.elements[path].node
		if node.selectionOwner == list {
			childIndex, logicalIndex = index, node.selectionIndex
			break
		}
	}
	m.mu.RUnlock()
	require.NotEmpty(t, paths[list])
	require.Greater(t, logicalIndex, 0, "fixture must expose recycled rows after scrolling")
	// Complete virtual child order now uses source indices, independently of
	// which renderer happens to be first in the viewport.
	childIndex = logicalIndex
	var accepted bool
	listClient := client.Object(bus, paths[list])
	require.NoError(t, listClient.CallWithContext(ctx, atspiSelection+".SelectChild", 0, int32(childIndex)).Store(&accepted))
	require.True(t, accepted)
	var selected []int
	runOnMain(func() { selected = list.Selected() })
	require.Equal(t, []int{logicalIndex}, selected)
	var reference atspiReference
	require.NoError(t, listClient.CallWithContext(ctx, atspiSelection+".GetSelectedChild", 0, int32(0)).Store(&reference))
	m.mu.RLock()
	actual := m.elements[reference.Path].node.selectionIndex
	m.mu.RUnlock()
	require.Equal(t, logicalIndex, actual)

	tableClient := client.Object(bus, paths[table])
	var rows, columns []int32
	require.NoError(t, tableClient.CallWithContext(ctx, atspiTable+".GetSelectedRows", 0).Store(&rows))
	require.Equal(t, []int32{1}, rows)
	var count dbus.Variant
	require.NoError(t, tableClient.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, atspiTable, "NSelectedRows").Store(&count))
	require.Equal(t, int32(1), count.Value())
	require.NoError(t, tableClient.CallWithContext(ctx, atspiTable+".AddColumnSelection", 0, int32(2)).Store(&accepted))
	require.True(t, accepted)
	require.NoError(t, tableClient.CallWithContext(ctx, atspiTable+".GetSelectedColumns", 0).Store(&columns))
	require.Equal(t, []int32{2}, columns)
	require.NoError(t, tableClient.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, atspiTable, "NSelectedColumns").Store(&count))
	require.Equal(t, int32(1), count.Value())
	require.NoError(t, tableClient.CallWithContext(ctx, atspiTable+".GetAccessibleAt", 0, int32(7), int32(2)).Store(&reference))
	m.mu.RLock()
	cell := m.elements[reference.Path].node
	m.mu.RUnlock()
	require.Equal(t, 7, cell.row)
	require.Equal(t, 2, cell.column)

	sliderPath := paths[slider]
	signals := make(chan *dbus.Signal, 16)
	client.Signal(signals)
	defer client.RemoveSignal(signals)
	require.NoError(t, client.AddMatchSignal(dbus.WithMatchInterface("org.a11y.atspi.Event.Object"), dbus.WithMatchMember("PropertyChange"), dbus.WithMatchObjectPath(sliderPath)))
	sliderClient := client.Object(bus, sliderPath)
	require.NoError(t, sliderClient.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Set", 0, atspiValue, "CurrentValue", dbus.MakeVariant(float64(7))).Err)
	require.NoError(t, sliderClient.CallWithContext(ctx, "org.freedesktop.DBus.Properties.Get", 0, atspiValue, "CurrentValue").Store(&count))
	require.Equal(t, float64(7), count.Value())
	select {
	case signal := <-signals:
		require.Equal(t, "accessible-value", signal.Body[0])
		require.Equal(t, float64(7), signal.Body[3].(dbus.Variant).Value())
	case <-time.After(2 * time.Second):
		t.Fatal("native numeric property-change event was not delivered")
	}
}
