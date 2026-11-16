//go:build windows && !no_glfw && !wasm && !test_web_driver && !mobile && !ci

package glfw

import (
	"bufio"
	"context"
	"fmt"
	"image"
	"image/color"
	"io"
	"math"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/driver/desktop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestNativeAcrylicFirstFrameAfterDXGIRecovery(t *testing.T) {
	requireVisualTest(t)
	if !windowsAcrylicRefreshRequired() || os.Getenv("FYNE_THEME") != "dark" {
		t.Skip("requires the dark Acrylic workaround on Windows 11 26300.9974")
	}
	setVisualTestTheme(t)
	bg := createWindowWithResizeCallback("Проверка первого кадра Acrylic — фон")
	defer bg.Close()
	w := newVisualTestWindow("Проверка первого кадра Acrylic после восстановления")
	defer runOnMain(w.Close)
	request, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropDesktopAcrylic})
	require.NoError(t, err)
	marker := canvas.NewRectangle(color.NRGBA{R: 240, G: 30, B: 170, A: 255})
	marker.Move(granio.NewPos(255, 225))
	marker.Resize(granio.NewSize(36, 36))
	checker := canvas.NewRasterWithPixels(func(x, y, _, _ int) color.Color {
		if (x/16+y/16)%2 == 0 {
			return color.NRGBA{R: 255, G: 40, B: 30, A: 255}
		}
		return color.NRGBA{R: 45, G: 10, B: 5, A: 255}
	})
	var hwnd uintptr
	placeBehind := func() error {
		background := uintptr(unsafe.Pointer(bg.viewport.GetWin32Window()))
		return stackVisualWindows(hwnd, background)
	}
	runOnMain(func() {
		bg.window.SetPadded(false)
		bg.window.SetContent(checker)
		bg.window.Resize(granio.NewSize(360, 290))
		bg.window.Show()
		bg.viewport.SetPos(220, 200)
		bg.viewport.SetAttrib(glfw.Floating, glfw.True)
		w.SetPadded(false)
		scene := visualAlphaScene().(*granio.Container)
		scene.Add(marker)
		w.SetContent(scene)
		w.Resize(granio.NewSize(360, 290))
		w.created = true
		w.create()
		w.viewport.SetPos(220, 200)
		w.viewport.SetAttrib(glfw.Floating, glfw.True)
		hwnd = uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
		disabled := int32(1)
		visualDWM.NewProc("DwmSetWindowAttribute").Call(hwnd, 3, uintptr(unsafe.Pointer(&disabled)), 4)
		w.Show()
		focusVisualTestWindow(w)
		err = placeBehind()
	})
	require.NoError(t, err)
	requireComposition(t, w)
	require.Equal(t, desktop.BackdropDesktopAcrylic, w.VisualStatus().Selected)

	// Establish the live response to the checker. A dark fallback would also
	// have low contrast, but it has a different colour from the sampled source.
	var before *image.NRGBA
	var captureErr error
	require.EventuallyWithT(t, func(c *assert.CollectT) {
		before, captureErr = visualScreen(w, bg)
		require.NoError(c, captureErr)
		require.NotNil(c, before)
		require.Equal(c, [3]float64{240, 30, 170}, visualRGB(before, 270, 240), "marker before device loss")
		base := visualRGB(before, 40, 75)
		require.Greater(c, base[0]-base[1], 15.0, "live red checker in Acrylic: %v", base)
		require.Less(c, visualDeviation(before), 25.0, "initial checker must be blurred")
	}, time.Second, 5*time.Millisecond)
	want := visualRGB(before, 40, 75)
	removed := syscall.NewCallback(func(_, _, _ uintptr) uintptr { return 0x887a0005 })
	for _, tc := range []struct {
		name   string
		hidden bool
		marker color.NRGBA
	}{
		{"visible", false, color.NRGBA{R: 30, G: 240, B: 80, A: 255}},
		{"hidden", true, color.NRGBA{R: 70, G: 105, B: 230, A: 255}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var old, recovered *wglDXGIPresenter
			var currentHWND uintptr
			var stackErr error
			runOnMain(func() {
				if tc.hidden {
					w.Hide()
				}
				w.RunWithContext(func() {
					old = w.presentation.dxgi
					if old == nil {
						return
					}
					old.presentMethod = removed
					w.presentFrame()
					recovered = w.presentation.dxgi
				})
				if recovered == nil || recovered == old {
					return
				}
				marker.FillColor = tc.marker
				marker.Refresh()
				if tc.hidden {
					w.Show()
					focusVisualTestWindow(w)
					stackErr = placeBehind()
				} else {
					w.RunWithContext(func() { d.repaintWindow(w) })
				}
				// Observe the new frame before a periodic AccentPolicy refresh
				// can replace its original backdrop.
				w.presentation.acrylicRefresh.next = time.Now().Add(time.Second)
				currentHWND = uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
			})
			require.NotNil(t, recovered)
			require.NotSame(t, old, recovered)
			require.Equal(t, hwnd, currentHWND)
			require.NoError(t, stackErr)
			markerRGB := [3]float64{float64(tc.marker.R), float64(tc.marker.G), float64(tc.marker.B)}
			var first *image.NRGBA
			require.Eventually(t, func() bool {
				first, _ = visualScreen(w, bg)
				return first != nil && visualRGB(first, 270, 240) == markerRGB
			}, 500*time.Millisecond, 4*time.Millisecond, "the first recovered canvas frame must become visible")
			got := visualRGB(first, 40, 75)
			t.Logf("first recovered Acrylic: %v, before DXGI loss: %v, checker contrast %.3f", got, want, visualDeviation(first))
			require.Less(t, visualDeviation(first), 25.0, "the first recovered frame must blur the source")
			require.LessOrEqual(t, visualDelta(got, want), 5.0, "the first recovered frame must sample the background")
			assertVisualScreenAlpha(t, first)
			require.Equal(t, request, w.VisualStatus().RequestID)
			require.True(t, w.VisualStatus().SurfaceReady)
			assertVisualAlpha(t, visualFrame(w), true)
		})
	}
}

type acrylicSource struct {
	hwnd  uintptr
	input io.Writer
	lines *bufio.Scanner
}

func (s acrylicSource) read(t *testing.T, prefix string) string {
	t.Helper()
	for s.lines.Scan() {
		line := s.lines.Text()
		if value, ok := strings.CutPrefix(line, prefix); ok {
			return value
		}
		t.Log(line)
	}
	t.Fatalf("source closed before %q: %v", prefix, s.lines.Err())
	return ""
}

func startAcrylicSource(t *testing.T) acrylicSource {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	t.Cleanup(cancel)
	var command *exec.Cmd
	if native := os.Getenv("FYNE_DWM_NATIVE_SOURCE"); native != "" {
		command = exec.CommandContext(ctx, native)
	} else {
		executable, err := os.Executable()
		require.NoError(t, err)
		command = exec.CommandContext(ctx, executable, "-test.run", "^TestNativeDWMBackgroundSource$")
	}
	command.Env = append(os.Environ(), "FYNE_DWM_CHILD=full")
	stdin, err := command.StdinPipe()
	require.NoError(t, err)
	stdout, err := command.StdoutPipe()
	require.NoError(t, err)
	command.Stderr = command.Stdout
	require.NoError(t, command.Start())
	t.Cleanup(func() {
		_, _ = io.WriteString(stdin, "q\n")
		_ = stdin.Close()
		_ = command.Wait()
	})
	source := acrylicSource{input: stdin, lines: bufio.NewScanner(stdout)}
	_, err = fmt.Sscanf(source.read(t, "HWND "), "%d", &source.hwnd)
	require.NoError(t, err)
	require.NotZero(t, source.hwnd)
	var pid uint32
	visualUser32.NewProc("GetWindowThreadProcessId").Call(source.hwnd, uintptr(unsafe.Pointer(&pid)))
	require.Equal(t, uint32(command.Process.Pid), pid)
	return source
}

func (s acrylicSource) submit(t *testing.T, command string) {
	t.Helper()
	_, err := fmt.Fprintln(s.input, command)
	require.NoError(t, err)
	require.Equal(t, "00000000", s.read(t, "FRAME "))
}

func TestNativeAcrylicWithIndependentSource(t *testing.T) {
	requireVisualTest(t)
	if os.Getenv("FYNE_DWM_CROSS_PROCESS") != "1" {
		t.Skip("requires independent on-screen Acrylic acceptance")
	}
	setVisualTestTheme(t)
	source := startAcrylicSource(t)
	w := newVisualTestWindow("Проверка Desktop Acrylic с отдельным источником")
	defer runOnMain(w.Close)
	_, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropDesktopAcrylic})
	require.NoError(t, err)
	var hwnd uintptr
	var positioned uintptr
	var sourceBounds, foregroundBounds image.Rectangle
	runOnMain(func() {
		w.SetPadded(false)
		w.SetContent(visualAlphaScene())
		w.Resize(granio.NewSize(360, 290))
		w.created = true
		w.create()
		w.viewport.SetPos(220, 200)
		w.viewport.SetAttrib(glfw.Floating, glfw.True)
		hwnd = uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
		disabled := int32(1)
		visualDWM.NewProc("DwmSetWindowAttribute").Call(hwnd, 3, uintptr(unsafe.Pointer(&disabled)), 4)
		w.Show()
		var rect struct{ left, top, right, bottom int32 }
		visualUser32.NewProc("GetWindowRect").Call(hwnd, uintptr(unsafe.Pointer(&rect)))
		positioned, _, _ = visualUser32.NewProc("SetWindowPos").Call(source.hwnd, hwnd, uintptr(rect.left), uintptr(rect.top), 0, 0, 0x11)
		focusVisualTestWindow(w)
		if err = stackVisualWindows(hwnd, source.hwnd); err != nil {
			return
		}
		sourceBounds, err = visualClientBounds(source.hwnd)
		if err == nil {
			foregroundBounds, err = visualClientBounds(hwnd)
		}
	})
	require.NoError(t, err)
	require.NotZero(t, positioned)
	requireComposition(t, w)
	require.True(t, sourceBounds.In(foregroundBounds), "%v must fully cover %v", foregroundBounds, sourceBounds)
	status := w.VisualStatus()
	require.Equal(t, desktop.VisualReady, status.Phase)
	require.Equal(t, desktop.BackdropDesktopAcrylic, status.Selected)
	require.True(t, status.SurfaceReady && status.FirstFrameBeforeShow)
	assertVisualAlpha(t, visualFrame(w), true)
	capture := func() *image.NRGBA {
		var frame *image.NRGBA
		runOnMain(func() { frame, err = captureVisualClient(hwnd, source.hwnd) })
		require.NoError(t, err)
		return frame
	}

	// Three distinct colours expose alternating stale frames even when the
	// frame before the stale one used a different swap-chain buffer.
	const frames = 20
	type tint struct {
		r, g, b  uint32
		dominant int
	}
	colors := []tint{{30, 55, 225, 2}, {220, 40, 30, 0}, {35, 220, 40, 1}}
	updates := 0
	for i := 0; i < frames; i++ {
		c := colors[i%len(colors)]
		source.submit(t, fmt.Sprintf("%08x", c.r|c.g<<8|c.b<<16|0xff000000))
		// Outside the UI thread. Give the compositor a whole interval before
		// changing the source again, even if the new colour appears immediately.
		time.Sleep(170 * time.Millisecond)
		deadline := time.Now().Add(550 * time.Millisecond)
		for {
			seen := visualRGB(capture(), 40, 100)
			if seen[c.dominant]-math.Max(seen[(c.dominant+1)%3], seen[(c.dominant+2)%3]) > 12 {
				updates++
				break
			}
			if time.Now().After(deadline) {
				t.Logf("frame %d stayed stale: input %v, composed %v", i+1, c, seen)
				break
			}
			time.Sleep(25 * time.Millisecond)
		}
	}
	source.submit(t, "k")
	time.Sleep(300 * time.Millisecond)
	// Inspect the stable source across several refresh periods. A temporary
	// disappearance of Acrylic exposes either the checker or a flat flash.
	checker := capture()
	base := visualRGB(checker, 40, 75)
	maxContrast := visualDeviation(checker)
	maxVariation := 0.0
	samples := 1
	for until := time.Now().Add(600 * time.Millisecond); time.Now().Before(until); samples++ {
		time.Sleep(13 * time.Millisecond)
		checker = capture()
		maxContrast = max(maxContrast, visualDeviation(checker))
		maxVariation = max(maxVariation, visualDelta(base, visualRGB(checker, 40, 75)))
	}
	t.Logf("%s updates=%d/%d checker samples=%d max contrast=%.3f max variation=%.3f", os.Getenv("FYNE_THEME"), updates, frames, samples, maxContrast, maxVariation)
	require.Equal(t, frames, updates)
	require.Less(t, maxContrast, 40.0, "the material must keep the source blurred")
	require.Less(t, maxVariation, 4.0, "Acrylic must stay visually stable between source frames")
}

func TestNativeAcrylicRefreshFailure(t *testing.T) {
	requireVisualTest(t)
	if !windowsAcrylicRefreshRequired() {
		t.Skip("requires Windows 11 26300.9974")
	}
	var original *windows.LazyProc
	runOnMain(func() { original = setWindowCompositionAttribute })
	for _, fallback := range []struct {
		policy desktop.VisualFallbackPolicy
		phase  desktop.VisualPhase
	}{
		{desktop.VisualFallbackOpaque, desktop.VisualFallback},
		{desktop.VisualFallbackFail, desktop.VisualFailed},
	} {
		t.Run(fmt.Sprint(fallback.policy), func(t *testing.T) {
			w := newVisualTestWindow("Проверка fallback обновления Desktop Acrylic")
			defer runOnMain(w.Close)
			defer runOnMain(func() { setWindowCompositionAttribute = original })
			options := desktop.VisualOptions{Backdrop: desktop.BackdropDesktopAcrylic, Fallback: fallback.policy}
			request, err := w.ConfigureVisuals(options)
			require.NoError(t, err)
			var active uint64
			runOnMain(func() {
				w.SetPadded(false)
				w.SetContent(visualAlphaScene())
				w.Resize(granio.NewSize(360, 290))
				w.Show()
				active = w.presentation.acrylicRefresh.requestID
				setWindowCompositionAttribute = visualUser32.NewProc("GranioTestMissingAcrylicRefresh")
			})
			requireComposition(t, w)
			require.Equal(t, request, active)
			require.Eventually(t, func() bool {
				s := w.VisualStatus()
				return s.RequestID == request && s.Phase == fallback.phase && s.SurfaceReady && s.FirstFrameBeforeShow &&
					!s.NativeAccepted && s.Selected == desktop.BackdropOpaque && s.Reason == desktop.VisualNativeFailure
			}, 2*time.Second, 10*time.Millisecond)
			assertVisualAlpha(t, visualFrame(w), false)

			runOnMain(func() { setWindowCompositionAttribute = original })
			request, err = w.ConfigureVisuals(options)
			require.NoError(t, err)
			require.Eventually(t, func() bool {
				s := w.VisualStatus()
				return s.RequestID == request && s.Phase == desktop.VisualReady && s.Selected == options.Backdrop && s.SurfaceReady
			}, 2*time.Second, 10*time.Millisecond)
			assertVisualAlpha(t, visualFrame(w), true)
		})
	}
}
