//go:build windows && !no_glfw && !ci && !mobile && !wasm && !test_web_driver

package glfw

import (
	"image"
	"image/color"
	"testing"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/stretchr/testify/require"
)

func TestNativeCaptureMinimizedWin32(t *testing.T) {
	w := createWindowWithResizeCallback("Capture minimized")
	defer w.Close()
	w.SetPadded(false)
	w.SetContent(canvas.NewRectangle(color.NRGBA{G: 255, A: 255}))
	w.Resize(granio.NewSize(80, 60))
	runOnMain(func() { w.window.Show() })
	ensureCanvasSize(t, w, granio.NewSize(80, 60))
	var capture image.Image
	var size granio.Size
	var minimized bool
	runOnMain(func() {
		w.viewport.Iconify()
		minimized = w.viewport.GetAttrib(glfw.Iconified) == glfw.True
		size = w.window.Canvas().Size()
		capture = w.window.Canvas().Capture()
		w.viewport.Restore()
	})
	require.True(t, minimized)
	require.Equal(t, granio.NewSize(80, 60), size)
	require.NotNil(t, capture)
	require.Equal(t, image.Rect(0, 0, 80, 60), capture.Bounds())
	r, g, b, _ := capture.At(40, 30).RGBA()
	require.Less(t, r, uint32(0xff))
	require.Greater(t, g, uint32(0xff00))
	require.Less(t, b, uint32(0xff))
}
