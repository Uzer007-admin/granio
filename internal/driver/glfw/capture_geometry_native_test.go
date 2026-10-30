//go:build !no_glfw && !ci && !mobile && !wasm && !test_web_driver

package glfw

import (
	"image"
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/stretchr/testify/require"
)

func TestNativeCaptureAfterZeroSizeNotification(t *testing.T) {
	w := createWindowWithResizeCallback("Capture geometry")
	defer w.Close()
	w.SetPadded(false)
	w.SetContent(canvas.NewRectangle(color.NRGBA{G: 255, A: 255}))
	w.Resize(granio.NewSize(80, 60))
	runOnMain(func() { w.window.Show() })
	ensureCanvasSize(t, w, granio.NewSize(80, 60))
	var capture image.Image
	var size granio.Size
	runOnMain(func() {
		// Win32 sends this exact callback from WM_SIZE/SIZE_MINIMIZED.
		w.window.resized(w.viewport, 0, 0)
		size = w.window.Canvas().Size()
		capture = w.window.Canvas().Capture()
	})
	require.Equal(t, granio.NewSize(80, 60), size, "a suspended native surface must not erase canvas geometry")
	require.NotNil(t, capture)
	require.Equal(t, image.Rect(0, 0, 80, 60), capture.Bounds())
	r, g, b, _ := capture.At(40, 30).RGBA()
	require.Less(t, r, uint32(0xff))
	require.Greater(t, g, uint32(0xff00))
	require.Less(t, b, uint32(0xff))
	w.Resize(granio.NewSize(101, 79))
	capture, err := w.glCanvasCaptureForTest()
	require.NoError(t, err)
	require.Equal(t, image.Rect(0, 0, 101, 79), capture.Bounds(), "the first frame after a positive resize must use the new geometry")
	_, g, _, _ = capture.At(50, 39).RGBA()
	require.Greater(t, g, uint32(0xff00))
}
