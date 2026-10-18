//go:build windows && !no_glfw && !wasm && !test_web_driver && !mobile

package glfw

import (
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/stretchr/testify/require"
)

func TestNativeWindowVisualsWindowsPortable(t *testing.T) {
	requireVisualTest(t)
	w := newVisualTestWindow("Fyne portable WGL background")
	defer runOnMain(w.Close)
	_, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropTransparent})
	require.NoError(t, err)
	runOnMain(func() {
		w.SetPadded(false)
		w.SetContent(visualAlphaScene())
		w.Resize(granio.NewSize(360, 290))
		w.Show()
	})
	ensureCanvasSize(t, &safeWindow{w}, granio.NewSize(360, 290))
	for _, mode := range []desktop.WindowBackdrop{desktop.BackdropTransparent, desktop.BackdropSystemBlur, desktop.BackdropOpaque} {
		id, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: mode})
		require.NoError(t, err)
		require.Eventually(t, func() bool { s := w.VisualStatus(); return s.RequestID == id && s.Phase != desktop.VisualPending }, 3*time.Second, 10*time.Millisecond)
		s := w.VisualStatus()
		t.Logf("%s: %+v caps=%+v", mode, s, w.VisualCapabilities())
		require.True(t, s.SurfaceReady)
		if s.Reason == desktop.VisualReasonNone {
			require.Equal(t, mode, s.Selected)
			require.Equal(t, desktop.VisualReady, s.Phase)
		}
		assertVisualAlpha(t, visualFrame(w), s.Selected != desktop.BackdropOpaque)
		require.Equal(t, uint64(1), w.VisualCapabilities().Generation)
	}
}
