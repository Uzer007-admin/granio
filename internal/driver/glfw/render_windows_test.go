//go:build windows && !no_glfw && !ci && !mobile && !wasm && !test_web_driver

package glfw

import (
	"image"
	"image/color"
	"testing"
	"unsafe"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/stretchr/testify/require"
)

func TestSoftwareFrameGDIColorOrder(t *testing.T) {
	// Проверяем DIB в памяти, без чтения рабочего стола или окон приложений.
	dc, _, _ := softwareGDI32.NewProc("CreateCompatibleDC").Call(0)
	require.NotZero(t, dc)
	defer softwareGDI32.NewProc("DeleteDC").Call(dc)
	info := softwareBitmapInfo{size: 40, width: 2, height: -2, planes: 1, bits: 32}
	var pixels uintptr
	bitmap, _, _ := softwareGDI32.NewProc("CreateDIBSection").Call(dc, uintptr(unsafe.Pointer(&info)), 0, uintptr(unsafe.Pointer(&pixels)), 0, 0)
	require.NotZero(t, bitmap)
	defer softwareGDI32.NewProc("DeleteObject").Call(bitmap)
	old, _, _ := softwareGDI32.NewProc("SelectObject").Call(dc, bitmap)
	defer softwareGDI32.NewProc("SelectObject").Call(dc, old)
	frame := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	frame.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	frame.SetNRGBA(1, 0, color.NRGBA{G: 255, A: 255})
	frame.SetNRGBA(0, 1, color.NRGBA{B: 255, A: 255})
	require.NoError(t, drawSoftwareFrame(dc, frame))
	for _, sample := range []struct{ x, y, rgb uintptr }{{0, 0, 0xff}, {1, 0, 0xff00}, {0, 1, 0xff0000}} {
		rgb, _, _ := softwareGDI32.NewProc("GetPixel").Call(dc, sample.x, sample.y)
		require.Equal(t, sample.rgb, rgb)
	}
}

func TestSoftwareNativeWindowCaptureResizeAndVisualFallback(t *testing.T) {
	if !softwareRendererRequested() {
		t.Skip("run with GRANIO_RENDERER=software")
	}
	var raw *window
	var err error
	runOnMain(func() {
		raw = d.CreateWindow("Проверка программного окна").(*window)
		_, err = raw.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropMica, Fallback: desktop.VisualFallbackOpaque})
		raw.SetPadded(false)
		raw.SetContent(canvas.NewRectangle(color.NRGBA{R: 255, A: 255}))
		raw.Resize(granio.NewSize(120, 80))
		raw.Show()
	})
	w := &safeWindow{window: raw}
	defer w.Close()
	require.NoError(t, err)
	ensureCanvasSize(t, w, granio.NewSize(120, 80))
	var img image.Image
	var api int
	var context *glfw.Window
	runOnMain(func() {
		api = w.viewport.GetAttrib(glfw.ClientAPI)
		context = glfw.GetCurrentContext()
		img, err = w.canvas.CaptureWithError()
	})
	require.Equal(t, glfw.NoAPI, api)
	require.Nil(t, context)
	require.NoError(t, err)
	require.NotNil(t, img)
	r, g, b, a := img.At(img.Bounds().Dx()/2, img.Bounds().Dy()/2).RGBA()
	require.Equal(t, uint32(65535), r)
	require.Zero(t, g)
	require.Zero(t, b)
	require.Equal(t, uint32(65535), a)
	status := w.VisualStatus()
	require.Equal(t, desktop.BackdropOpaque, status.Selected)
	require.Equal(t, desktop.VisualUnsupportedBackend, status.Reason)
	require.True(t, status.SurfaceReady)
	require.True(t, status.FirstFrameBeforeShow)
	w.Resize(granio.NewSize(180, 120))
	ensureCanvasSize(t, w, granio.NewSize(180, 120))
	runOnMain(func() { img, err = w.canvas.CaptureWithError() })
	require.NoError(t, err)
	require.Greater(t, img.Bounds().Dx(), 120)
	require.Greater(t, img.Bounds().Dy(), 80)
	// Неподдерживаемый эффект сохраняет непрозрачность без пересоздания окна,
	// которое не может предоставить alpha-поверхность композитору.
	_, err = w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropTransparent, Fallback: desktop.VisualFallbackOpaque})
	require.NoError(t, err)
	runOnMain(func() {})
	require.Equal(t, desktop.VisualUnsupportedBackend, w.VisualStatus().Reason)
	runOnMain(func() {
		w.viewport.Iconify()
		_, err = w.canvas.CaptureWithError()
		w.viewport.Restore()
	})
	require.ErrorIs(t, err, desktop.ErrCaptureUnavailable)
	runOnMain(func() { img, err = w.canvas.CaptureWithError() })
	require.NoError(t, err)
	require.NotEmpty(t, img.Bounds())
	w.Close()
	_, err = w.canvas.CaptureWithError()
	require.ErrorIs(t, err, desktop.ErrCaptureUnavailable)
}
