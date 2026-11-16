//go:build windows && !ci && !no_glfw && !mobile && !wasm && !test_web_driver

package glfw

import (
	"image"
	"image/color"
	"testing"
	"time"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/driver"
	"github.com/Uzer007-admin/granio/v3/driver/desktop"
	"github.com/stretchr/testify/require"
)

func TestClientTitleBarRespectsClientSize(t *testing.T) {
	var w *window
	var configErr error
	var active bool
	initial := granio.NewSize(400, 220)
	runOnMain(func() {
		w = d.CreateWindow("Client title bar size test").(*window)
		configErr = w.SetRenderMode(driver.RenderSoftware)
		if configErr == nil {
			configErr = w.SetClientTitleBar(desktop.ClientTitleBarOptions{Height: 40, LeadingReserved: 80, TrailingReserved: 100})
		}
		if configErr != nil {
			return
		}
		w.SetPadded(false)
		w.SetContent(canvas.NewRectangle(color.Black))
		w.Resize(initial)
		w.Show()
		active = w.ClientTitleBarActive()
	})
	defer runOnMain(w.Close)
	require.NoError(t, configErr)
	require.True(t, active)

	for _, size := range []granio.Size{initial, granio.NewSize(450, 260)} {
		var actual, expected image.Point
		var capture image.Image
		var captureErr error
		runOnMain(func() {
			if size != initial {
				w.Resize(size)
			}
			expected.X, expected.Y = w.screenSize(size)
			actual.X, actual.Y = w.viewport.GetSize()
			capture, captureErr = w.canvas.CaptureWithError()
		})
		require.Equal(t, expected, actual)
		require.NoError(t, captureErr)
		require.Equal(t, actual, capture.Bounds().Size())
	}
	waitMaximized := func(expected bool) {
		t.Helper()
		require.Eventually(t, func() bool {
			var maximized bool
			runOnMain(func() { maximized = w.Maximized() })
			return maximized == expected
		}, time.Second, 10*time.Millisecond)
	}
	runOnMain(w.ToggleMaximize)
	waitMaximized(true)
	runOnMain(w.ToggleMaximize)
	waitMaximized(false)
	runOnMain(w.Minimize)
	require.Eventually(t, func() bool {
		var minimized bool
		runOnMain(func() { minimized = w.viewport.GetAttrib(glfw.Iconified) == glfw.True })
		return minimized
	}, time.Second, 10*time.Millisecond)
}
