//go:build !no_glfw && !wasm && !test_web_driver && !mobile

package glfw

import (
	"image"
	"image/color"
	"os"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/require"
)

func TestNativeCanvasAndPopupMaterials(t *testing.T) {
	if os.Getenv("FYNE_TEST_MATERIAL") != "1" {
		t.Skip("native material FBO acceptance; unsupported devices use opaque fallback")
	}
	w := createWindowWithResizeCallback("Canvas material integration")
	defer w.Close()
	base := canvas.NewRasterWithPixels(func(x, y, width, height int) color.Color {
		return color.NRGBA{R: uint8(x * 251 / max(1, width)), G: uint8(y * 239 / max(1, height)), B: 31, A: 255}
	})
	m := canvas.NewMaterial(color.NRGBA{R: 200, G: 10, B: 20, A: 255})
	m.BlurRadius = 0
	m.CornerRadius = 0
	m.Tint = color.NRGBA{}
	m.Downsample = 1
	var reference, filtered image.Image
	var stats map[string]uint64
	runOnMain(func() { w.window.SetContent(base); w.window.Resize(granio.NewSize(360, 240)); w.window.Show() })
	require.Eventually(t, func() bool {
		ready := false
		runOnMain(func() { ready = w.canvas.Size().Width > 300 && w.canvas.Size().Height > 200 })
		return ready
	}, time.Second, 10*time.Millisecond)
	runOnMain(func() {
		reference = w.canvas.Capture()
		m.Resize(granio.NewSize(160, 100))
		m.Move(granio.NewPos(40, 50))
		overlay := container.NewWithoutLayout(m)
		w.canvas.Overlays().Add(overlay)
		filtered = w.canvas.Capture()
		w.canvas.Capture()
		stats = w.canvas.Painter().(interface{ MaterialStatistics() map[string]uint64 }).MaterialStatistics()
	})
	require.NotEmpty(t, stats, "native GPU pipeline must run, not silently fall back")
	require.Greater(t, stats["cache_hits"], uint64(0))
	require.LessOrEqual(t, stats["pool_bytes"]+stats["capture_bytes"], stats["budget_bytes"])
	for _, pos := range []image.Point{{60, 70}, {160, 120}, {100, 95}} {
		require.Equal(t, reference.At(pos.X, pos.Y), filtered.At(pos.X, pos.Y), "GPU snapshot orientation and lower-layer composition at %v", pos)
	}
	var fallback image.Image
	runOnMain(func() { m.Enabled = false; m.Refresh(); fallback = w.canvas.Capture() })
	r, g, b, a := fallback.At(100, 95).RGBA()
	require.Equal(t, []uint32{200 * 257, 10 * 257, 20 * 257, 65535}, []uint32{r, g, b, a})
	var popup *widget.PopUp
	runOnMain(func() {
		for _, overlay := range w.canvas.Overlays().List() {
			w.canvas.Overlays().Remove(overlay)
		}
		popup = widget.NewPopUp(widget.NewLabel("Material popup"), w.canvas)
		popup.Material = canvas.NewMaterial(color.NRGBA{A: 255})
		popup.ShowAtPosition(granio.NewPos(30, 20))
		popup.Resize(granio.NewSize(180, 80))
		w.canvas.Capture()
		popup.Hide()
		w.canvas.Capture()
	})
}
