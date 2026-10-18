//go:build accessibility && linux && !no_glfw && !mobile

package glfw

import (
	"context"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Uzer007-admin/dbus/v5"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/require"
)

// A second bus connection models a native assistive-technology client. Run with
// FYNE_TEST_ATSPI=1 under dbus-run-session with at-spi2-core installed.
func TestATSPINativeClient(t *testing.T) {
	if os.Getenv("FYNE_TEST_ATSPI") != "1" {
		t.Skip("requires the native AT-SPI accessibility bus")
	}
	var win *window
	var clicks atomic.Int32
	var entry, password *widget.Entry
	var button *widget.Button
	runOnMain(func() {
		entry = widget.NewEntry()
		entry.SetText("a界z")
		button = widget.NewButton("Apply", func() { clicks.Add(1) })
		password = widget.NewPasswordEntry()
		password.SetText("secret")
		win = d.CreateWindow("Accessible native fixture").(*window)
		win.create()
		win.SetContent(container.NewVBox(button, entry, password))
		win.Show()
		win.updateAccessibility()
	})
	t.Cleanup(func() { runOnMain(func() { win.Close() }) })
	require.Eventually(t, func() bool {
		linuxAccessibility.mu.RLock()
		defer linuxAccessibility.mu.RUnlock()
		return linuxAccessibility.conn != nil
	}, 5*time.Second, 10*time.Millisecond)
	runOnMain(func() { win.updateAccessibility() })
	session, err := dbus.ConnectSessionBus()
	require.NoError(t, err)
	defer session.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var address string
	require.NoError(t, session.Object("org.a11y.Bus", "/org/a11y/bus").CallWithContext(ctx, "org.a11y.Bus.GetAddress", 0).Store(&address))
	client, err := dbus.Connect(address)
	require.NoError(t, err)
	defer client.Close()
	m := linuxAccessibility
	m.mu.RLock()
	bus := m.busName
	paths := append([]dbus.ObjectPath(nil), m.windows[win]...)
	m.mu.RUnlock()
	var buttonPath, entryPath, passwordPath dbus.ObjectPath
	for _, path := range paths {
		m.mu.RLock()
		element := m.elements[path]
		m.mu.RUnlock()
		switch element.node.object {
		case button:
			buttonPath = path
		case entry:
			entryPath = path
		case password:
			passwordPath = path
		}
	}
	require.NotEmpty(t, buttonPath)
	require.NotEmpty(t, entryPath)
	require.NotEmpty(t, passwordPath)
	var children []atspiReference
	require.NoError(t, client.Object(bus, atspiRoot).CallWithContext(ctx, atspiAccessible+".GetChildren", 0).Store(&children))
	require.NotEmpty(t, children)
	var text string
	require.NoError(t, client.Object(bus, entryPath).CallWithContext(ctx, atspiText+".GetText", 0, int32(1), int32(2)).Store(&text))
	require.Equal(t, "界", text)
	require.NoError(t, client.Object(bus, passwordPath).CallWithContext(ctx, atspiText+".GetText", 0, int32(0), int32(-1)).Store(&text))
	require.Empty(t, text)
	var accepted bool
	require.NoError(t, client.Object(bus, buttonPath).CallWithContext(ctx, atspiAction+".DoAction", 0, int32(0)).Store(&accepted))
	require.True(t, accepted)
	require.Eventually(t, func() bool { return clicks.Load() == 1 }, time.Second, time.Millisecond)
	require.NoError(t, client.Object(bus, entryPath).CallWithContext(ctx, atspiEditable+".SetTextContents", 0, "日本語").Store(&accepted))
	require.True(t, accepted)
	runOnMain(func() { require.Equal(t, "日本語", entry.Text); win.updateAccessibility() })
	require.NoError(t, client.Object(bus, entryPath).CallWithContext(ctx, atspiText+".GetText", 0, int32(0), int32(-1)).Store(&text))
	require.Equal(t, "日本語", text)
	runOnMain(func() { win.SetContent(widget.NewLabel("Removed")); win.updateAccessibility() })
	require.Error(t, client.Object(bus, entryPath).CallWithContext(ctx, atspiText+".GetText", 0, int32(0), int32(-1)).Store(&text))
}
