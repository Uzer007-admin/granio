//go:build linux && !no_glfw && !wasm && !test_web_driver && !mobile

package glfw

import (
	"errors"
	"image"
	"net"
	"os"
	"testing"
	"time"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/stretchr/testify/require"
)

func TestWindowVisualsLinuxPolicy(t *testing.T) {
	base := windowVisualPolicy{platform: glfw.BackgroundWayland, composition: true, alpha: true, blur: true}
	baseCaps := desktop.VisualCapabilities{RendererBackdrops: 0x3f, TransparentSurface: true}
	for _, test := range []struct {
		name   string
		change func(*windowVisualPolicy, *desktop.VisualCapabilities)
		want   desktop.VisualFallbackReason
	}{
		{"supported", nil, desktop.VisualReasonNone},
		{"compositor_off", func(p *windowVisualPolicy, _ *desktop.VisualCapabilities) { p.composition = false }, desktop.VisualCompositorUnavailable},
		{"blur_off", func(p *windowVisualPolicy, _ *desktop.VisualCapabilities) { p.blur = false }, desktop.VisualBlurUnavailable},
		{"high_contrast", func(p *windowVisualPolicy, _ *desktop.VisualCapabilities) { p.highContrast = true }, desktop.VisualHighContrast},
		{"transparency_off", func(p *windowVisualPolicy, _ *desktop.VisualCapabilities) {
			p.transparencyKnown = true
			p.transparent = false
		}, desktop.VisualTransparencyDisabled},
		{"renderer", func(_ *windowVisualPolicy, c *desktop.VisualCapabilities) { c.RendererBackdrops = 1 }, desktop.VisualUnsupportedBackend},
		{"format", func(_ *windowVisualPolicy, c *desktop.VisualCapabilities) { c.TransparentSurface = false }, desktop.VisualSurfaceNotTransparent},
	} {
		t.Run(test.name, func(t *testing.T) {
			p, c := base, baseCaps
			if test.change != nil {
				test.change(&p, &c)
			}
			require.Equal(t, test.want, windowBackdropRejection(desktop.VisualOptions{Backdrop: desktop.BackdropSystemBlur}, p, &c, nil))
			require.Equal(t, desktop.VisualReasonNone, windowBackdropRejection(desktop.VisualOptions{}, p, &c, errors.New("query failed")))
		})
	}
	base.blur = false
	require.Equal(t, desktop.VisualReasonNone, windowBackdropRejection(desktop.VisualOptions{Backdrop: desktop.BackdropTransparent}, base, &baseCaps, nil), "transparency is independent of blur support")
}

// CaptureBack reads the already prepared buffer, without causing a repaint.
func preparedVisualFrame(w *window) image.Image {
	var result image.Image
	runOnMain(func() {
		w.RunWithContext(func() {
			result = w.canvas.Painter().(interface {
				CaptureBack(granio.Canvas) image.Image
			}).CaptureBack(w.canvas)
		})
	})
	return result
}

func TestNativeWindowVisualsLinuxPortable(t *testing.T) {
	requireVisualTest(t)
	for _, backdrop := range []desktop.WindowBackdrop{desktop.BackdropTransparent, desktop.BackdropSystemBlur} {
		t.Run(backdrop.String(), func(t *testing.T) {
			w := newVisualTestWindow("Fyne portable background")
			defer runOnMain(w.Close)
			_, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: backdrop})
			require.NoError(t, err)
			runOnMain(func() {
				w.SetPadded(false)
				_ = w.SetScalePolicy(desktop.ScaleExact)
				w.SetContent(visualAlphaScene())
				w.Resize(granio.NewSize(360, 290))
				w.created = true
				w.create()
			})
			status, caps := w.VisualStatus(), w.VisualCapabilities()
			t.Logf("hidden: %+v caps=%+v", status, caps)
			require.True(t, status.SurfaceReady)
			require.True(t, status.FirstFrameBeforeShow)
			if os.Getenv("FYNE_EXPECT_TRANSPARENT") == "1" {
				require.True(t, caps.TransparentSurface)
			}
			if backdrop == desktop.BackdropSystemBlur && os.Getenv("FYNE_EXPECT_SYSTEM_BLUR") == "1" {
				require.Equal(t, desktop.VisualReady, status.Phase)
			}
			selected := status.Selected != desktop.BackdropOpaque
			assertVisualAlpha(t, preparedVisualFrame(w), selected)
			var viewport *glfw.Window
			runOnMain(func() { viewport = w.viewport; w.Show() })
			ensureCanvasSize(t, &safeWindow{w}, granio.NewSize(360, 290))
			assertVisualAlpha(t, visualFrame(w), selected)
			if selected {
				frame := visualFrame(w)
				for i, a := range []uint32{0, 64, 128, 192, 255} {
					var x, y int
					runOnMain(func() { x, y = w.canvas.PixelCoordinateForPosition(granio.NewPos(float32(40+i*60), 40)) })
					r, g, b, alpha := frame.At(x, y).RGBA()
					require.InDelta(t, float64(a), float64(alpha)/257, 1)
					for channel, value := range []uint32{r, g, b} {
						require.InDelta(t, []float64{32, 176, 240}[channel]*float64(a)/255, float64(value)/257, 1)
					}
				}
			}
			runOnMain(func() { w.Resize(granio.NewSize(410, 310)); w.Hide(); w.Show() })
			ensureCanvasSize(t, &safeWindow{w}, granio.NewSize(410, 310))
			require.Equal(t, uint64(1), w.VisualCapabilities().Generation)
			var currentViewport *glfw.Window
			runOnMain(func() { currentViewport = w.viewport })
			require.Same(t, viewport, currentViewport)
			assertVisualAlpha(t, visualFrame(w), selected)
			// Stable capabilities must not keep resubmitting prepared frames.
			before := w.VisualStatus().PreparedFrames
			time.Sleep(1100 * time.Millisecond)
			require.Equal(t, before, w.VisualStatus().PreparedFrames)
			var original granio.Theme
			contrast, err := theme.NewModern(theme.ModernOptions{HighContrast: true})
			require.NoError(t, err)
			runOnMain(func() {
				original = granio.CurrentApp().Settings().Theme()
				granio.CurrentApp().Settings().SetTheme(contrast)
			})
			defer runOnMain(func() { granio.CurrentApp().Settings().SetTheme(original) })
			require.Eventually(t, func() bool {
				s := w.VisualStatus()
				return s.Phase == desktop.VisualFallback && s.Reason == desktop.VisualHighContrast
			}, 3*time.Second, 10*time.Millisecond)
			assertVisualAlpha(t, visualFrame(w), false)
		})
	}
}

func sendBackgroundFixture(t *testing.T, command string) {
	t.Helper()
	path := os.Getenv("FYNE_TEST_BACKGROUND_SOCKET")
	if path == "" {
		t.Skip("owned compositor background fixture")
	}
	conn, err := net.DialUnix("unixgram", nil, &net.UnixAddr{Name: path, Net: "unixgram"})
	require.NoError(t, err)
	defer conn.Close()
	_, err = conn.Write([]byte(command))
	require.NoError(t, err)
}

func TestNativeWindowVisualsLinuxCapabilityChanges(t *testing.T) {
	requireVisualTest(t)
	if os.Getenv("FYNE_TEST_BACKGROUND_SOCKET") == "" {
		t.Skip("owned compositor background fixture")
	}
	w := newVisualTestWindow("Fyne blur capability changes")
	defer runOnMain(w.Close)
	id, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropSystemBlur})
	require.NoError(t, err)
	var viewport *glfw.Window
	runOnMain(func() {
		w.SetPadded(false)
		_ = w.SetScalePolicy(desktop.ScaleExact)
		w.SetContent(visualAlphaScene())
		w.Resize(granio.NewSize(360, 290))
		w.Show()
		viewport = w.viewport
	})
	ensureCanvasSize(t, &safeWindow{w}, granio.NewSize(360, 290))
	require.Equal(t, desktop.VisualReady, w.VisualStatus().Phase)
	assertVisualAlpha(t, visualFrame(w), true)
	for _, change := range []struct {
		command string
		want    desktop.VisualPhase
	}{
		{"blur off", desktop.VisualFallback}, {"blur on", desktop.VisualReady},
		{"remove", desktop.VisualFallback}, {"add", desktop.VisualReady},
	} {
		sendBackgroundFixture(t, change.command)
		require.Eventually(t, func() bool { return w.VisualStatus().Phase == change.want }, 3*time.Second, 10*time.Millisecond, change.command)
		s := w.VisualStatus()
		t.Logf("%s: %+v", change.command, s)
		require.Equal(t, id, s.RequestID)
		require.Equal(t, uint64(1), w.VisualCapabilities().Generation)
		assertVisualAlpha(t, visualFrame(w), change.want == desktop.VisualReady)
		var currentViewport *glfw.Window
		runOnMain(func() { currentViewport = w.viewport })
		require.Same(t, viewport, currentViewport)
	}
	// Real fractional-scale protocol events change the EGL buffer size. Blur
	// regions remain surface-local and do not multiply in framebuffer density.
	for _, scale := range []struct {
		command   string
		numerator float32
	}{{"scale 150", 1.25}, {"scale 240", 2}, {"scale 120", 1}} {
		sendBackgroundFixture(t, scale.command)
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			require.Eventually(t, func() bool {
				var actual float32
				runOnMain(func() { actual = w.ScaleInfo().FramebufferScale })
				return actual == scale.numerator
			}, 3*time.Second, 10*time.Millisecond)
			assertVisualAlpha(t, visualFrame(w), true)
		}
	}
	before := w.VisualStatus().PreparedFrames
	time.Sleep(1100 * time.Millisecond)
	require.Equal(t, before, w.VisualStatus().PreparedFrames)
	_, err = w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropTransparent})
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		s := w.VisualStatus()
		return s.Phase == desktop.VisualReady && s.Selected == desktop.BackdropTransparent
	}, 3*time.Second, 10*time.Millisecond)
	var state glfw.BackgroundEffectState
	runOnMain(func() { state, err = w.viewport.QueryBackgroundEffects() })
	require.NoError(t, err)
	require.False(t, state.BlurRequested)
}
