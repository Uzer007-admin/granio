//go:build !no_glfw && !ci && !mobile && !wasm && !test_web_driver && !gles && !arm && !arm64

package glfw

import (
	"image"
	"image/color"
	"testing"
	"unsafe"

	gl "github.com/Uzer007-admin/gl/v2.1/fyne"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/stretchr/testify/require"
)

func TestPainterTablesSurviveFailedGlobalReload(t *testing.T) {
	a, b := createWindowWithResizeCallback("dispatch A"), createWindowWithResizeCallback("dispatch B")
	defer a.Close()
	defer b.Close()
	a.SetContent(canvas.NewRectangle(color.NRGBA{G: 255, A: 255}))
	b.SetContent(canvas.NewRectangle(color.NRGBA{R: 255, A: 255}))
	a.Resize(granio.NewSize(80, 60))
	b.Resize(granio.NewSize(96, 64))
	runOnMain(func() { a.Show(); b.Show() })
	ensureCanvasSize(t, a, granio.NewSize(80, 60))
	ensureCanvasSize(t, b, granio.NewSize(96, 64))
	repaintWindow(a)
	repaintWindow(b)
	runOnMain(func() { a.window.Show(); b.window.Show() })
	// Unlike createWindow's layout-only fixture, this test needs native resize
	// callbacks and mapped GLX drawables before checking actual pixel contents.
	greenBefore, redBefore := a.Canvas().Capture(), b.Canvas().Capture()
	_, green, _, _ := greenBefore.At(greenBefore.Bounds().Dx()/2, greenBefore.Bounds().Dy()/2).RGBA()
	red, _, _, _ := redBefore.At(redBefore.Bounds().Dx()/2, redBefore.Bounds().Dy()/2).RGBA()
	require.Greater(t, green, uint32(0xff00), "baseline native green drawable")
	require.Greater(t, red, uint32(0xff00), "baseline native red drawable")
	// Both painters initialize private tables. A failed legacy/global reload
	// must not affect their programs, captures, or subsequent window switching.
	var reloadErr error
	runOnMain(func() { reloadErr = gl.InitWithProcAddrFunc(func(string) unsafe.Pointer { return nil }) })
	require.Error(t, reloadErr)
	for range 2 {
		var green, red image.Image
		runOnMain(func() { green, red = a.window.Canvas().Capture(), b.window.Canvas().Capture() })
		gBounds, rBounds := green.Bounds(), red.Bounds()
		gr, gg, gb, _ := green.At(gBounds.Min.X+gBounds.Dx()/2, gBounds.Min.Y+gBounds.Dy()/2).RGBA()
		rr, rg, rb, _ := red.At(rBounds.Min.X+rBounds.Dx()/2, rBounds.Min.Y+rBounds.Dy()/2).RGBA()
		require.Greater(t, gg, uint32(0xff00))
		require.Less(t, gr, uint32(0xff))
		require.Less(t, gb, uint32(0xff))
		require.Greater(t, rr, uint32(0xff00))
		require.Less(t, rg, uint32(0xff))
		require.Less(t, rb, uint32(0xff))
	}
}
