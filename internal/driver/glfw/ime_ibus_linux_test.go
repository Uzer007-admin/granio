//go:build linux && !no_glfw && !wasm && !test_web_driver && !mobile

package glfw

import (
	"os"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/require"
)

func TestIBusNativePasswordAndLifecycle(t *testing.T) {
	if os.Getenv("FYNE_TEST_REAL_IBUS") != "1" {
		t.Skip("requires real IBus session")
	}
	w := createWindowWithResizeCallback("IBus privacy and lifetime")
	entry := widget.NewPasswordEntry()
	entry.SetText("secret 日本 😀")
	var input *ibusInput
	var redacted, private bool
	runOnMain(func() {
		w.window.SetContent(entry)
		w.window.Resize(granio.NewSize(360, 180))
		w.window.Show()
		w.canvas.Focus(entry)
		curWindow = w.window
		w.window.updateInputMethod()
		if stored, ok := ibusWindows.Load(w.window); ok {
			input = stored.(*ibusInput)
			redacted = input.text == "" && input.cursor == 0 && input.anchor == 0
			private = input.sensitive
		}
	})
	require.NotNil(t, input, "must connect to real IBus daemon")
	require.True(t, redacted)
	require.True(t, private)
	runOnMain(func() { w.window.Close() })
	require.Eventually(t, input.closed.Load, time.Second, time.Millisecond)
	_, retained := ibusWindows.Load(w.window)
	require.False(t, retained, "closed context must not retain the window")
}
