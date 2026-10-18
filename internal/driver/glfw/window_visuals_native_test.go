//go:build !no_glfw && !wasm && !test_web_driver && !mobile

package glfw

import (
	"image"
	"image/color"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal/async"
	"github.com/Uzer007-admin/granio/v2/internal/painter/gl"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/require"
)

func newVisualTestWindow(title string) *window {
	var w *window
	runOnMain(func() { w = d.CreateWindow(title).(*window) })
	return w
}

func requireVisualTest(t *testing.T) {
	if os.Getenv("FYNE_TEST_WINDOW_VISUALS") != "1" {
		t.Skip("real native window visuals acceptance")
	}
}

// Flat alpha swatches exercise straight shape colours. The rest of this scene
// covers the glyph atlas, strokes, image/SVG textures, shadows and GPU materials.
func visualAlphaScene() granio.CanvasObject {
	objects := make([]granio.CanvasObject, 0, 12)
	for i, a := range []uint8{0, 64, 128, 192, 255} {
		r := canvas.NewRectangle(color.NRGBA{R: 32, G: 176, B: 240, A: a})
		r.Move(granio.NewPos(20+float32(i)*60, 20))
		r.Resize(granio.NewSize(40, 40))
		objects = append(objects, r)
	}
	text := canvas.NewText("Alpha · Русский · 日本語", color.NRGBA{R: 224, G: 220, B: 200, A: 128})
	text.Move(granio.NewPos(20, 85))
	text.TextSize = 18
	objects = append(objects, text)
	line := canvas.NewLine(color.NRGBA{R: 220, G: 25, B: 120, A: 192})
	line.Position1 = granio.NewPos(20, 120)
	line.Position2 = granio.NewPos(280, 143)
	line.StrokeWidth = 2
	objects = append(objects, line)
	r := canvas.NewRectangle(color.NRGBA{R: 220, G: 220, B: 240, A: 64})
	r.Move(granio.NewPos(20, 160))
	r.Resize(granio.NewSize(60, 35))
	r.CornerRadius = 8
	r.Shadow = canvas.Shadow{Color: color.NRGBA{R: 4, G: 5, B: 8, A: 128}, BlurRadius: 8, Offset: granio.NewPos(5, 5)}
	objects = append(objects, r)
	bitmap := image.NewNRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			bitmap.SetNRGBA(x, y, color.NRGBA{R: 30, G: 240, B: 60, A: uint8(x * 17)})
		}
	}
	img := canvas.NewImageFromImage(bitmap)
	img.Move(granio.NewPos(110, 160))
	img.Resize(granio.NewSize(48, 35))
	objects = append(objects, img)
	svg := canvas.NewImageFromResource(granio.NewStaticResource("alpha.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="40" height="40"><circle cx="20" cy="20" r="17" fill="#fa7040" fill-opacity="0.5"/></svg>`)))
	svg.Move(granio.NewPos(180, 160))
	svg.Resize(granio.NewSize(40, 35))
	objects = append(objects, svg)
	m := canvas.NewMaterial(color.NRGBA{A: 255})
	m.Move(granio.NewPos(240, 160))
	m.Resize(granio.NewSize(45, 35))
	m.Tint = color.NRGBA{R: 30, G: 100, B: 180, A: 40}
	m.BlurRadius = 2
	objects = append(objects, m)
	button := widget.NewButton("Fyne", nil)
	button.Move(granio.NewPos(20, 225))
	button.Resize(granio.NewSize(100, 32))
	objects = append(objects, button)
	return container.NewWithoutLayout(objects...)
}

func visualFrame(w *window) image.Image {
	var img image.Image
	runOnMain(func() { img = w.canvas.Capture() })
	return img
}

func assertVisualAlpha(t *testing.T, img image.Image, transparent bool) {
	require.NotNil(t, img)
	bounds := img.Bounds()
	require.Greater(t, bounds.Dx(), 300)
	require.Greater(t, bounds.Dy(), 200)
	invalid := 0
	dirtyTransparent := 0
	zeros := 0
	nonOpaque := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			if r > a+257 || g > a+257 || b > a+257 {
				invalid++
			}
			if a == 0 {
				zeros++
				if r != 0 || g != 0 || b != 0 {
					dirtyTransparent++
				}
			}
			if a != 65535 {
				nonOpaque++
			}
		}
	}
	require.Zero(t, invalid, "every stock pass must write premultiplied RGB")
	require.Zero(t, dirtyTransparent, "RGB must be exactly zero at zero alpha")
	if transparent {
		require.Greater(t, zeros, 100)
		require.Greater(t, nonOpaque, 100)
	} else {
		require.Zero(t, nonOpaque, "fallback is fully opaque, including edges")
	}
}

func TestNativeWindowVisualsLifecycleAndAlpha(t *testing.T) {
	requireVisualTest(t)
	w := newVisualTestWindow("Fyne WGL alpha")
	defer runOnMain(w.Close)
	var callbacksMu sync.Mutex
	var callbacks []desktop.VisualStatus
	var wrongThread atomic.Bool
	cancel := w.ObserveVisualStatus(func(s desktop.VisualStatus) {
		if !async.IsMainGoroutine() {
			wrongThread.Store(true)
		}
		callbacksMu.Lock()
		callbacks = append(callbacks, s)
		callbacksMu.Unlock()
	})
	defer cancel()
	id, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropMica})
	require.NoError(t, err)
	require.Equal(t, uint64(1), id)
	runOnMain(func() {
		w.SetPadded(false)
		w.SetContent(visualAlphaScene())
		w.Resize(granio.NewSize(360, 290))
		w.Show()
	})
	ensureCanvasSize(t, &safeWindow{w}, granio.NewSize(360, 290))
	status := w.VisualStatus()
	caps := w.VisualCapabilities()
	t.Logf("first show: status=%+v capabilities=%+v", status, caps)
	require.True(t, status.SurfaceReady)
	require.True(t, status.FirstFrameBeforeShow)
	require.GreaterOrEqual(t, status.PreparedFrames, uint32(2))
	if runtime.GOOS == "windows" && caps.OSBackdrops.Supports(desktop.BackdropMica) && status.Reason == desktop.VisualReasonNone {
		require.Equal(t, desktop.VisualReady, status.Phase)
		require.True(t, caps.TransparentSurface)
		require.Greater(t, caps.AlphaBits, 0)
		frame := visualFrame(w)
		assertVisualAlpha(t, frame, true)
		scale := float64(frame.Bounds().Dx()) / 360
		for i, a := range []uint32{0, 64, 128, 192, 255} {
			r, g, b, alpha := frame.At(int(float64(40+i*60)*scale), int(40*scale)).RGBA()
			require.InDelta(t, float64(a), float64(alpha)/257, 1, "alpha swatch %d", i)
			require.InDelta(t, float64(32*a)/255, float64(r)/257, 1)
			require.InDelta(t, float64(176*a)/255, float64(g)/257, 1)
			require.InDelta(t, float64(240*a)/255, float64(b)/257, 1)
		}
	} else {
		require.Equal(t, desktop.VisualFallback, status.Phase)
		require.Equal(t, desktop.BackdropOpaque, status.Selected)
		assertVisualAlpha(t, visualFrame(w), false)
	}
	// Multiple requests before returning control to the owner thread coalesce.
	var latest uint64
	runOnMain(func() {
		_, err = w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropMicaAlt})
		if err == nil {
			latest, err = w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropDesktopAcrylic})
		}
	})
	require.NoError(t, err)
	require.Eventually(t, func() bool { s := w.VisualStatus(); return s.RequestID == latest && s.Phase != desktop.VisualPending }, 2*time.Second, 10*time.Millisecond)
	require.Equal(t, desktop.BackdropDesktopAcrylic, w.VisualStatus().Requested)
	before := w.VisualStatus()
	_, err = w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropSystemBlur + 1})
	require.ErrorIs(t, err, desktop.ErrVisualOptions)
	require.Equal(t, before.RequestID, w.VisualStatus().RequestID)
	// Explicit high contrast forces actual opaque pixels, retaining the request.
	var original granio.Theme
	contrast, err := theme.NewModern(theme.ModernOptions{HighContrast: true})
	require.NoError(t, err)
	runOnMain(func() {
		original = granio.CurrentApp().Settings().Theme()
		granio.CurrentApp().Settings().SetTheme(contrast)
		w.applyWindowVisuals()
	})
	defer runOnMain(func() { granio.CurrentApp().Settings().SetTheme(original) })
	if runtime.GOOS == "windows" && caps.OSBackdrops.Supports(desktop.BackdropMica) {
		require.Equal(t, desktop.VisualHighContrast, w.VisualStatus().Reason)
	}
	assertVisualAlpha(t, visualFrame(w), false)
	strictID, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropMica, Fallback: desktop.VisualFallbackFail})
	require.NoError(t, err)
	require.Eventually(t, func() bool { s := w.VisualStatus(); return s.RequestID == strictID && s.Phase == desktop.VisualFailed }, 2*time.Second, 10*time.Millisecond)
	assertVisualAlpha(t, visualFrame(w), false)
	// Cancellation covers notifications which were queued on the UI thread.
	var afterCancel atomic.Int32
	runOnMain(func() {
		stop := w.ObserveVisualStatus(func(desktop.VisualStatus) { afterCancel.Add(1) })
		stop()
	})
	runOnMain(func() {})
	require.Zero(t, afterCancel.Load())
	require.False(t, wrongThread.Load())
	runOnMain(w.Close)
	require.Equal(t, desktop.VisualClosed, w.VisualStatus().Phase)
	_, err = w.ConfigureVisuals(desktop.VisualOptions{})
	require.ErrorIs(t, err, desktop.ErrVisualWindowClosed)
	callbacksMu.Lock()
	defer callbacksMu.Unlock()
	require.NotEmpty(t, callbacks)
	require.Equal(t, desktop.VisualClosed, callbacks[len(callbacks)-1].Phase)
}

func TestNativeWindowVisualsFormatReservationAndPendingClose(t *testing.T) {
	requireVisualTest(t)
	plain := createWindow("Visuals unreserved")
	defer plain.Close()
	_, err := plain.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropMica})
	if windowVisualAlphaFormat() {
		require.ErrorIs(t, err, desktop.ErrVisualNeedsRecreate)
	} else {
		require.NoError(t, err)
	}
	reserved := newVisualTestWindow("Visuals opaque reservation")
	defer runOnMain(reserved.Close)
	_, err = reserved.ConfigureVisuals(desktop.VisualOptions{PrepareAlpha: true})
	require.NoError(t, err)
	runOnMain(func() {
		reserved.SetPadded(false)
		reserved.SetContent(visualAlphaScene())
		reserved.Resize(granio.NewSize(360, 290))
		reserved.Show()
	})
	ensureCanvasSize(t, &safeWindow{reserved}, granio.NewSize(360, 290))
	require.Equal(t, desktop.BackdropOpaque, reserved.VisualStatus().Selected)
	assertVisualAlpha(t, visualFrame(reserved), false)
	_, err = reserved.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropMica})
	require.NoError(t, err, "pre-reserved opaque windows may change material without recreation")
	runOnMain(func() { reserved.Hide(); reserved.Resize(granio.NewSize(410, 310)); reserved.Show() })
	ensureCanvasSize(t, &safeWindow{reserved}, granio.NewSize(410, 310))
	runOnMain(func() { reserved.Resize(granio.NewSize(360, 290)) })
	ensureCanvasSize(t, &safeWindow{reserved}, granio.NewSize(360, 290))
	require.Equal(t, uint64(1), reserved.VisualCapabilities().Generation)
	assertVisualAlpha(t, visualFrame(reserved), reserved.VisualStatus().Selected != desktop.BackdropOpaque)
	// Creating another ordinary window must not inherit the global alpha hint.
	second := createWindow("Visuals hint isolation")
	defer second.Close()
	var leaked bool
	runOnMain(func() { leaked = second.viewport.GetAttrib(glfw.TransparentFramebuffer) != glfw.False })
	require.False(t, leaked)
	var closed atomic.Int32
	pending := newVisualTestWindow("Visuals pending close")
	runOnMain(func() { pending.SetOnClosed(func() { closed.Add(1) }) })
	_, err = pending.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropMica})
	require.NoError(t, err)
	runOnMain(func() { pending.Close(); pending.Show() })
	require.Equal(t, desktop.VisualClosed, pending.VisualStatus().Phase)
	require.Equal(t, int32(1), closed.Load())
	var pendingCreated bool
	runOnMain(func() { pendingCreated = pending.viewport != nil })
	require.False(t, pendingCreated)
}

func TestNativeWindowVisualsRendererFailureStaysHidden(t *testing.T) {
	requireVisualTest(t)
	w := newVisualTestWindow("Visuals renderer failure")
	defer runOnMain(w.Close)
	_, err := w.ConfigureVisuals(desktop.VisualOptions{PrepareAlpha: true})
	require.NoError(t, err)
	var working gl.Painter
	runOnMain(func() {
		w.SetPadded(false)
		w.SetContent(visualAlphaScene())
		w.Resize(granio.NewSize(360, 290))
		w.Show()
		w.Hide()
		working = w.canvas.Painter()
		// An uninitialised painter models a failed GL loader. It must never be
		// asked to draw, or have an undefined frame exposed by Show.
		w.canvas.SetPainter(gl.NewPainter(w.canvas, w))
	})
	defer runOnMain(func() { w.canvas.SetPainter(working) })
	request, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropMica})
	require.NoError(t, err)
	require.Eventually(t, func() bool { s := w.VisualStatus(); return s.RequestID == request && s.Phase == desktop.VisualFailed }, 2*time.Second, 10*time.Millisecond)
	var visible bool
	runOnMain(func() { w.Show(); visible = w.viewport.GetAttrib(glfw.Visible) != glfw.False })
	require.False(t, visible)
	require.False(t, w.VisualStatus().SurfaceReady)
	require.Zero(t, w.VisualStatus().PreparedFrames)
}

func TestNativeWindowVisualsMaterialSurfaceCache(t *testing.T) {
	requireVisualTest(t)
	w := createWindowWithResizeCallback("Material clear colour cache")
	defer w.Close()
	base := canvas.NewRectangle(color.NRGBA{R: 40, G: 100, B: 160, A: 64})
	m := canvas.NewMaterial(color.NRGBA{R: 255, A: 255})
	m.BlurRadius = 0
	m.CornerRadius = 0
	m.Tint = color.NRGBA{}
	m.Downsample = 1
	m.Move(granio.NewPos(40, 40))
	m.Resize(granio.NewSize(100, 70))
	runOnMain(func() {
		w.window.SetPadded(false)
		w.window.SetContent(container.NewStack(base, container.NewWithoutLayout(m)))
		w.window.Resize(granio.NewSize(360, 290))
		w.window.Show()
	})
	ensureCanvasSize(t, w, granio.NewSize(360, 290))
	var actual image.Image
	for _, mode := range []gl.WindowSurfaceMode{gl.WindowSurfaceOpaque, gl.WindowSurfaceBackdrop, gl.WindowSurfaceOpaque, gl.WindowSurfaceBackdrop} {
		var supported bool
		var x, y int
		runOnMain(func() {
			w.RunWithContext(func() { supported = w.canvas.Painter().(surfacePainter).SetWindowSurface(mode) })
			if supported {
				actual = w.canvas.Capture()
				x, y = w.canvas.PixelCoordinateForPosition(granio.NewPos(70, 60))
			}
		})
		if !supported {
			t.Skip("alpha framebuffer unavailable")
		}
		_, _, _, alpha := actual.At(x, y).RGBA()
		if mode == gl.WindowSurfaceBackdrop {
			require.InDelta(t, 64, float64(alpha)/257, 1, "cached material must not retain an opaque canvas clear")
		} else {
			require.Equal(t, uint32(65535), alpha)
		}
	}
	for _, quality := range []uint8{1, 2, 4} {
		for _, corners := range []float32{0, 12} {
			runOnMain(func() {
				m.Downsample = quality
				m.CornerRadius = corners
				m.Noise = 0
				m.Refresh()
				actual = w.canvas.Capture()
			})
			for _, point := range []granio.Position{granio.NewPos(40, 40), granio.NewPos(41, 41), granio.NewPos(42, 48), granio.NewPos(48, 42), granio.NewPos(50, 50), granio.NewPos(70, 60), granio.NewPos(139, 109)} {
				var x, y int
				runOnMain(func() { x, y = w.canvas.PixelCoordinateForPosition(point) })
				r, g, b, a := actual.At(x, y).RGBA()
				require.InDelta(t, 64, float64(a)/257, 2, "rounded mask must preserve alpha: quality=%d corners=%g point=%v", quality, corners, point)
				for channel, v := range []uint32{r, g, b} {
					require.InDelta(t, []float64{10, 25, 40}[channel], float64(v)/257, 2)
				}
			}
		}
	}
}
