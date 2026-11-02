//go:build windows && !no_glfw && !ci && !mobile && !wasm && !test_web_driver

package glfw

import (
	"image/color"
	"os"
	"syscall"
	"testing"
	"unsafe"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/stretchr/testify/require"
)

func requireComposition(t *testing.T, w *window) *wglDXGIPresenter {
	t.Helper()
	var p *wglDXGIPresenter
	runOnMain(func() { p = w.presentation.dxgi })
	if p == nil || p.composition == nil {
		if os.Getenv("GRANIO_REQUIRE_DXGI") == "1" {
			t.Fatal("hardware DirectComposition presenter is required")
		}
		t.Skip("hardware DirectComposition interop is unavailable")
	}
	return p
}

func TestWGLDCompAlphaFirstFrameResizeAndRecovery(t *testing.T) {
	requireVisualTest(t)
	w := newVisualTestWindow("Проверка alpha DirectComposition")
	defer runOnMain(w.Close)
	_, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropTransparent})
	require.NoError(t, err)
	runOnMain(func() {
		w.SetPadded(false)
		w.SetContent(visualAlphaScene())
		w.Resize(granio.NewSize(360, 290))
		w.created = true
		w.create()
	})
	p := requireComposition(t, w)
	s := w.VisualStatus()
	require.Equal(t, desktop.VisualReady, s.Phase)
	require.Equal(t, desktop.VisualBackendDirectComposition, s.Backend)
	require.True(t, s.FirstFrameBeforeShow)
	require.True(t, s.SurfaceReady)
	require.True(t, p.composition.attached)
	require.Greater(t, p.width, 1)
	require.Greater(t, p.height, 1)
	var desc graphicsCompositionDesc
	var descErr error
	runOnMain(func() {
		descErr = graphicsResult("test swap description", p.swap.call(18, uintptr(unsafe.Pointer(&desc))))
	})
	require.NoError(t, descErr)
	require.Equal(t, uint32(1), desc.alpha)
	require.Equal(t, uint32(3), desc.effect)
	require.Equal(t, uint32(p.width), desc.width)
	require.Equal(t, uint32(p.height), desc.height)
	assertVisualAlpha(t, visualFrame(w), true)
	checkPixels := func() {
		t.Helper()
		for i, alpha := range []uint8{0, 64, 128, 192, 255} {
			var pixel color.NRGBA
			var sampleErr error
			runOnMain(func() {
				w.RunWithContext(func() {
					x, y := w.canvas.PixelCoordinateForPosition(granio.NewPos(float32(40+i*60), 40))
					pixel, sampleErr = sampleDXGIPixel(p, uint32(x), uint32(y))
				})
			})
			require.NoError(t, sampleErr)
			require.InDelta(t, alpha, pixel.A, 1, "alpha swatch %d", i)
			for channel, value := range []uint8{pixel.R, pixel.G, pixel.B} {
				require.InDelta(t, float64([]uint8{32, 176, 240}[channel])*float64(alpha)/255, value, 1, "premultiplied swatch %d channel %d", i, channel)
			}
		}
	}
	checkPixels() // Проверяется первый скрытый кадр, до Show.
	runOnMain(w.Show)
	for _, size := range []granio.Size{granio.NewSize(410, 310), granio.NewSize(360, 290)} {
		runOnMain(func() { w.Resize(size); w.RunWithContext(func() { d.repaintWindow(w) }) })
		ensureCanvasSize(t, &safeWindow{w}, size)
		runOnMain(func() { w.RunWithContext(func() { d.repaintWindow(w) }) })
		checkPixels()
	}
	var zeroErr, limitErr error
	var before, after uint64
	runOnMain(func() {
		w.RunWithContext(func() {
			before = p.frames
			zeroErr = p.present(0, p.height)
			limitErr = p.present(p.maxDimension+1, p.height)
			after = p.frames
		})
	})
	require.NoError(t, zeroErr)
	require.Error(t, limitErr)
	require.Equal(t, before, after, "zero and oversized frames must preserve the last valid frame")
	var opaqueID uint64
	runOnMain(func() {
		opaqueID, err = w.ConfigureVisuals(desktop.VisualOptions{PrepareAlpha: true})
	})
	require.NoError(t, err)
	runOnMain(func() {})
	require.Equal(t, opaqueID, w.VisualStatus().RequestID)
	assertVisualAlpha(t, visualFrame(w), false)
	var opaquePixel color.NRGBA
	runOnMain(func() {
		w.RunWithContext(func() { opaquePixel, err = sampleDXGIPixel(p, 10, 70) })
	})
	require.NoError(t, err)
	require.Equal(t, uint8(255), opaquePixel.A, "the presented fallback must also be opaque")
	_, err = w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropTransparent})
	require.NoError(t, err)
	runOnMain(func() {})
	checkPixels()
	runOnMain(func() {
		w.viewport.Iconify()
		w.RunWithContext(w.presentFrame)
		w.viewport.Restore()
		w.RunWithContext(func() { d.repaintWindow(w) })
	})
	checkPixels()
	removed := syscall.NewCallback(func(_, _, _ uintptr) uintptr { return 0x887a0005 })
	old := p
	runOnMain(func() {
		w.RunWithContext(func() {
			old.presentMethod = removed
			w.presentFrame()
			p = w.presentation.dxgi
			d.repaintWindow(w)
		})
	})
	require.NotNil(t, p)
	require.NotSame(t, old, p)
	require.Nil(t, old.composition)
	require.Nil(t, old.device)
	require.NotNil(t, p.composition)
	checkPixels()
	// Потеря устройства в первом скрытом Present не должна навсегда
	// оставить Show без готовой поверхности после успешного восстановления.
	var prepared bool
	runOnMain(func() {
		w.Hide()
		w.RunWithContext(func() {
			w.releasePresentation()
			w.initPresentation()
			w.presentation.dxgi.presentMethod = removed
		})
		prepared = w.applyWindowVisuals()
		p = w.presentation.dxgi
	})
	require.True(t, prepared)
	require.True(t, w.VisualStatus().SurfaceReady)
	require.NotNil(t, p.composition)
	checkPixels()
	runOnMain(func() { w.releasePresentation(); w.releasePresentation() })
	require.Nil(t, p.composition)
	require.Nil(t, p.swap)
	require.Zero(t, p.d3d)
}

func TestWGLDCompUnavailableUsesOpaqueDXGI(t *testing.T) {
	requireVisualTest(t)
	w := newVisualTestWindow("Проверка непрозрачного fallback DXGI")
	defer runOnMain(w.Close)
	_, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropTransparent})
	require.NoError(t, err)
	runOnMain(func() {
		w.SetPadded(false)
		w.SetContent(visualAlphaScene())
		w.Resize(granio.NewSize(360, 290))
		w.created = true
		w.create()
	})
	requireComposition(t, w)
	var blocker, fallback *wglDXGIPresenter
	var blockErr error
	var retry bool
	runOnMain(func() {
		w.RunWithContext(func() {
			w.releasePresentation()
			// Реальный отказ CreateTargetForHwnd: занимаем тот же слой HWND,
			// не присоединяя к нему visual и не меняя настройки системы.
			blocker, blockErr = newWGLDXGIPresenter(uintptr(unsafe.Pointer(w.viewport.GetWin32Window())), true)
			if blockErr == nil {
				w.initPresentation()
				retry = w.retryPresentationWindow(false)
				blocker.destroy()
				blocker = nil
			}
		})
		if retry {
			// Повторяет обработку отказа до показа окна: флаг неизменяем,
			// поэтому fallback получает обычный, ещё скрытый HWND.
			w.viewport.Destroy()
			w.viewport = nil
			w.create()
			fallback = w.presentation.dxgi
		}
	})
	defer runOnMain(func() {
		if blocker != nil {
			w.RunWithContext(blocker.destroy)
		}
	})
	require.NoError(t, blockErr)
	require.True(t, retry)
	require.NotNil(t, fallback)
	require.Nil(t, fallback.composition)
	s := w.VisualStatus()
	require.Equal(t, desktop.VisualFallback, s.Phase)
	require.Equal(t, desktop.BackdropOpaque, s.Selected)
	require.Equal(t, desktop.VisualUnsupportedBackend, s.Reason)
	require.True(t, s.SurfaceReady)
	assertVisualAlpha(t, visualFrame(w), false)
	_, err = w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropTransparent, Fallback: desktop.VisualFallbackFail})
	require.NoError(t, err)
	runOnMain(func() {})
	require.Equal(t, desktop.VisualFailed, w.VisualStatus().Phase)
}
