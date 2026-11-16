//go:build windows && !no_glfw && !wasm && !test_web_driver && !mobile && !ci

package glfw

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/driver/desktop"
	"github.com/Uzer007-admin/granio/v3/theme"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

var visualUser32 = windows.NewLazySystemDLL("user32.dll")
var visualGDI32 = windows.NewLazySystemDLL("gdi32.dll")
var visualDWM = windows.NewLazySystemDLL("dwmapi.dll")

// This diagnostic reads only our own foreground client rectangle. It refuses
// to read the screen if another top-level window owns any sampled point. It is
// not a renderer or a per-frame transfer path; DWM output cannot be read by GL.
func captureVisualClient(hwnd, background uintptr) (*image.NRGBA, error) {
	bounds, err := visualClientBounds(hwnd)
	if err != nil {
		return nil, err
	}
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 || width > 4096 || height > 4096 {
		return nil, fmt.Errorf("invalid owned client size %dx%d", width, height)
	}
	// Transparent pixels must reveal our background, not an application which
	// the desktop user raised between the two fixture windows.
	under, _, _ := visualUser32.NewProc("GetWindow").Call(hwnd, 2)
	for under != 0 {
		visible, _, _ := visualUser32.NewProc("IsWindowVisible").Call(under)
		var area struct{ left, top, right, bottom int32 }
		visualUser32.NewProc("GetWindowRect").Call(under, uintptr(unsafe.Pointer(&area)))
		intersects := bounds.Overlaps(image.Rect(int(area.left), int(area.top), int(area.right), int(area.bottom)))
		if visible != 0 && intersects {
			// A cloaked window keeps WS_VISIBLE and its bounds but is absent
			// from the desktop image. It cannot occlude the owned source.
			var cloaked uint32
			hr, _, _ := visualDWM.NewProc("DwmGetWindowAttribute").Call(under, 14, uintptr(unsafe.Pointer(&cloaked)), 4)
			if int32(hr) < 0 || cloaked == 0 {
				break
			}
		}
		under, _, _ = visualUser32.NewProc("GetWindow").Call(under, 2)
	}
	if background != 0 && under != background {
		return nil, fmt.Errorf("owned background 0x%x is not immediately behind our client (0x%x)", background, under)
	}
	for y := 4; y < height-4; y += 16 {
		for x := 4; x < width-4; x += 16 {
			point := uint64(uint32(bounds.Min.X+x)) | uint64(uint32(bounds.Min.Y+y))<<32
			owner, _, _ := visualUser32.NewProc("WindowFromPoint").Call(uintptr(point))
			root, _, _ := visualUser32.NewProc("GetAncestor").Call(owner, 2)
			if root != hwnd {
				return nil, fmt.Errorf("owned client is occluded at %d,%d", x, y)
			}
		}
	}
	dc, _, err := visualUser32.NewProc("GetDC").Call(0)
	if dc == 0 {
		return nil, err
	}
	defer visualUser32.NewProc("ReleaseDC").Call(0, dc)
	memory, _, err := visualGDI32.NewProc("CreateCompatibleDC").Call(dc)
	if memory == 0 {
		return nil, err
	}
	defer visualGDI32.NewProc("DeleteDC").Call(memory)
	bitmap, _, err := visualGDI32.NewProc("CreateCompatibleBitmap").Call(dc, uintptr(width), uintptr(height))
	if bitmap == 0 {
		return nil, err
	}
	defer visualGDI32.NewProc("DeleteObject").Call(bitmap)
	old, _, _ := visualGDI32.NewProc("SelectObject").Call(memory, bitmap)
	const captureComposedPixels = 0x00cc0020 | 0x40000000 // SRCCOPY | CAPTUREBLT
	ok, _, err := visualGDI32.NewProc("BitBlt").Call(memory, 0, 0, uintptr(width), uintptr(height), dc, uintptr(bounds.Min.X), uintptr(bounds.Min.Y), captureComposedPixels)
	visualGDI32.NewProc("SelectObject").Call(memory, old)
	if ok == 0 {
		return nil, err
	}
	var info struct {
		size                   uint32
		width, height          int32
		planes, bits           uint16
		compression, sizeImage uint32
		xppm, yppm             int32
		used, important        uint32
	}
	info.size = 40
	info.width = int32(width)
	info.height = -int32(height)
	info.planes = 1
	info.bits = 32
	image := image.NewNRGBA(image.Rect(0, 0, width, height))
	rows, _, err := visualGDI32.NewProc("GetDIBits").Call(dc, bitmap, 0, uintptr(height), uintptr(unsafe.Pointer(unsafe.SliceData(image.Pix))), uintptr(unsafe.Pointer(&info)), 0)
	if int(rows) != height {
		return nil, fmt.Errorf("GetDIBits %d/%d: %v", rows, height, err)
	}
	for i := 0; i < len(image.Pix); i += 4 {
		image.Pix[i], image.Pix[i+2] = image.Pix[i+2], image.Pix[i]
		image.Pix[i+3] = 255
	}
	return image, nil
}

func visualScreen(w *window, bg *safeWindow) (*image.NRGBA, error) {
	var img *image.NRGBA
	var err error
	runOnMain(func() {
		hwnd := uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
		background := uintptr(unsafe.Pointer(bg.viewport.GetWin32Window()))
		visualDWM.NewProc("DwmFlush").Call()
		img, err = captureVisualClient(hwnd, background)
	})
	return img, err
}

// Keep the two owned windows adjacent when other topmost windows are present.
func stackVisualWindows(foreground, background uintptr) error {
	batch, _, err := visualUser32.NewProc("BeginDeferWindowPos").Call(2)
	if batch == 0 {
		return fmt.Errorf("BeginDeferWindowPos: %w", err)
	}
	deferPos := visualUser32.NewProc("DeferWindowPos")
	const flags = 0x213 // SWP_NOSIZE | SWP_NOMOVE | SWP_NOACTIVATE | SWP_NOOWNERZORDER
	for _, hwnd := range []uintptr{background, foreground} {
		batch, _, err = deferPos.Call(batch, hwnd, ^uintptr(0), 0, 0, 0, 0, flags) // HWND_TOPMOST
		if batch == 0 {
			return fmt.Errorf("DeferWindowPos: %w", err)
		}
	}
	if ok, _, err := visualUser32.NewProc("EndDeferWindowPos").Call(batch); ok == 0 {
		return fmt.Errorf("EndDeferWindowPos: %w", err)
	}
	return nil
}

// Test-only focus setup. Native tests are launched from a background terminal,
// so Windows may reject a plain SetForegroundWindow. Attach only for the focus
// call and immediately detach; no hooks, input injection or production changes.
func focusVisualTestWindow(w *window) {
	foreground, _, _ := visualUser32.NewProc("GetForegroundWindow").Call()
	thread, _, _ := visualUser32.NewProc("GetWindowThreadProcessId").Call(foreground, 0)
	current := uintptr(windows.GetCurrentThreadId())
	if thread != 0 && thread != current {
		attached, _, _ := visualUser32.NewProc("AttachThreadInput").Call(current, thread, 1)
		if attached != 0 {
			defer visualUser32.NewProc("AttachThreadInput").Call(current, thread, 0)
		}
	}
	w.viewport.Focus()
}

func visualRGB(img image.Image, x, y float64) [3]float64 {
	width, height := img.Bounds().Dx(), img.Bounds().Dy()
	r, g, b, _ := img.At(int(x*float64(width)/360), int(y*float64(height)/290)).RGBA()
	return [3]float64{float64(r) / 257, float64(g) / 257, float64(b) / 257}
}

func visualClientBounds(hwnd uintptr) (image.Rectangle, error) {
	var rect struct{ left, top, right, bottom int32 }
	ok, _, err := visualUser32.NewProc("GetClientRect").Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	if ok == 0 {
		return image.Rectangle{}, err
	}
	var origin struct{ x, y int32 }
	ok, _, err = visualUser32.NewProc("ClientToScreen").Call(hwnd, uintptr(unsafe.Pointer(&origin)))
	if ok == 0 {
		return image.Rectangle{}, err
	}
	return image.Rect(0, 0, int(rect.right), int(rect.bottom)).Add(image.Pt(int(origin.x), int(origin.y))), nil
}

func visualBackgroundColor(w *window) (pixel color.NRGBA, frames uint64, err error) {
	w.RunWithContext(func() {
		p := w.presentation.dxgi
		if p == nil || p.shared == nil || p.width <= 20 || p.height <= 100 {
			err = fmt.Errorf("background fixture has no full DXGI frame")
			return
		}
		pixel, err = sampleDXGIPixel(p, 20, 100)
		frames = p.frames
	})
	return
}

func visualDelta(a, b [3]float64) float64 {
	return (math.Abs(a[0]-b[0]) + math.Abs(a[1]-b[1]) + math.Abs(a[2]-b[2])) / 3
}

func visualDeviation(img image.Image) float64 {
	var sum, squares, count float64
	for y := 65; y < 80; y++ {
		for x := 20; x < 320; x++ {
			v := visualRGB(img, float64(x), float64(y))[0]
			sum += v
			squares += v * v
			count++
		}
	}
	return math.Sqrt(math.Max(0, squares/count-sum*sum/count/count))
}

func assertVisualScreenAlpha(t *testing.T, frame *image.NRGBA) {
	t.Helper()
	for i, alpha := range []float64{0, .25, .5, .75, 1} {
		base := visualRGB(frame, float64(40+i*60), 75)
		pixel := visualRGB(frame, float64(40+i*60), 40)
		for channel, source := range [3]float64{32, 176, 240} {
			require.InDelta(t, alpha*source+(1-alpha)*base[channel], pixel[channel], 8, "screen alpha in the first frame")
		}
	}
}

func saveVisualImage(directory, name string, img image.Image) error {
	if directory == "" {
		return nil
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(directory, name+".png"))
	if err != nil {
		return err
	}
	err = png.Encode(f, img)
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}

type visualTestApp struct {
	granio.App
	settings granio.Settings
}

func (a visualTestApp) Settings() granio.Settings { return a.settings }

type visualTestDarkSettings struct{ granio.Settings }

func (visualTestDarkSettings) ThemeVariant() granio.ThemeVariant { return theme.VariantDark }

func setVisualTestTheme(t *testing.T) {
	t.Helper()
	if os.Getenv("FYNE_THEME") != "dark" {
		return
	}
	var original granio.App
	runOnMain(func() {
		original = granio.CurrentApp()
		granio.SetCurrentApp(visualTestApp{original, visualTestDarkSettings{original.Settings()}})
	})
	t.Cleanup(func() { runOnMain(func() { granio.SetCurrentApp(original) }) })
}

// Checks optical DWM output with two owned Fyne windows. In particular an
// accepted HRESULT does not satisfy the Acrylic and alpha composition gates.
func TestNativeWindowVisualsDWMComposition(t *testing.T) {
	requireVisualTest(t)
	setVisualTestTheme(t)
	red := color.NRGBA{R: 220, G: 40, B: 30, A: 255}
	blue := color.NRGBA{R: 30, G: 55, B: 225, A: 255}
	bg := createWindowWithResizeCallback("Fyne owned DWM test background")
	defer bg.Close()
	var bgContent *canvas.Rectangle
	runOnMain(func() {
		bgContent = canvas.NewRectangle(red)
		bg.window.SetPadded(false)
		bg.window.SetContent(bgContent)
		bg.window.Resize(granio.NewSize(360, 290))
		bg.window.Show()
		bg.viewport.SetPos(220, 200)
		bg.viewport.SetAttrib(glfw.Floating, glfw.True)
	})
	// Let the opaque source reach the desktop before covering it. Present and
	// HWND visibility can complete while DWM is still displaying an old frame.
	// Transparent captures below additionally verify the immediate underlay.
	require.Eventually(t, func() bool {
		var frame *image.NRGBA
		runOnMain(func() {
			hwnd := uintptr(unsafe.Pointer(bg.viewport.GetWin32Window()))
			visualDWM.NewProc("DwmFlush").Call()
			frame, _ = captureVisualClient(hwnd, 0)
		})
		return frame != nil && visualRGB(frame, 40, 75) == [3]float64{220, 40, 30}
	}, time.Second, 5*time.Millisecond, "the source must be visible before creating the foreground")
	w := newVisualTestWindow("Fyne WindowVisuals WGL/DWM")
	defer runOnMain(w.Close)
	_, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropDesktopAcrylic})
	require.NoError(t, err)
	var initialHWND, initialStyle uintptr
	var hidden image.Image
	runOnMain(func() {
		w.SetPadded(false)
		w.SetContent(visualAlphaScene())
		w.Resize(granio.NewSize(360, 290))
		w.created = true
		w.create()
		w.viewport.SetPos(220, 200)
		w.viewport.SetAttrib(glfw.Floating, glfw.True)
		initialHWND = uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
		initialStyle, _, _ = visualUser32.NewProc("GetWindowLongPtrW").Call(initialHWND, ^uintptr(15))
		// Иначе первый снимок попадает в системную анимацию Show: DWM
		// масштабирует и смешивает всё окно поверх фона. Отключаем переход
		// только для fixture, не меняя системные настройки или рабочий драйвер.
		disabled := int32(1)
		result, _, _ := visualDWM.NewProc("DwmSetWindowAttribute").Call(initialHWND, 3, uintptr(unsafe.Pointer(&disabled)), 4)
		if int32(result) < 0 {
			err = graphicsResult("test disable window transition", result)
		}
		w.RunWithContext(func() { hidden = w.canvas.Painter().Capture(w.canvas) })
	})
	require.NoError(t, err)
	requireComposition(t, w)
	var sourceDXGI, usingAccent bool
	runOnMain(func() {
		sourceDXGI = bg.presentation.dxgi != nil
		usingAccent = w.presentation.acrylicAccent
	})
	if !sourceDXGI {
		if os.Getenv("GRANIO_REQUIRE_DXGI") == "1" {
			t.Fatal("background fixture requires a hardware DXGI presenter")
		}
		t.Skip("background fixture requires a hardware DXGI presenter")
	}
	caps := w.VisualCapabilities()
	if !caps.OSBackdrops.Supports(desktop.BackdropDesktopAcrylic) || w.VisualStatus().Reason != desktop.VisualReasonNone {
		if os.Getenv("GRANIO_REQUIRE_DXGI") == "1" {
			t.Fatalf("native backdrop policy: %+v caps=%+v", w.VisualStatus(), caps)
		}
		t.Skipf("native policy: %+v caps=%+v", w.VisualStatus(), caps)
	}
	if caps.OSBuild == acrylicBackdropWorkaroundBuild {
		if !windowsAcrylicPowerAllowed() {
			t.Skip("Windows power policy has disabled the Acrylic workaround")
		}
		require.True(t, usingAccent, "the initial hidden frame must use the Acrylic workaround")
	}
	assertVisualAlpha(t, hidden, true)
	require.True(t, w.VisualStatus().FirstFrameBeforeShow)
	runOnMain(func() {
		w.Show()
		focusVisualTestWindow(w)
		background := uintptr(unsafe.Pointer(bg.viewport.GetWin32Window()))
		err = stackVisualWindows(initialHWND, background)
	})
	require.NoError(t, err)
	first, err := visualScreen(w, bg)
	require.NoError(t, err)
	// The HWND can own the hit-test region before the desktop has composed
	// its first frame. A screenshot entirely showing the red source at the
	// scene's sample points still belongs to the preceding desktop frame.
	// As soon as any foreground pixel appears, check that very frame below.
	stillSource := func(frame *image.NRGBA) bool {
		source := [3]float64{float64(red.R), float64(red.G), float64(red.B)}
		for _, point := range [][2]float64{{40, 40}, {100, 40}, {160, 40}, {220, 40}, {280, 40}, {40, 75}, {64, 240}} {
			if visualDelta(visualRGB(frame, point[0], point[1]), source) > 2 {
				return false
			}
		}
		return true
	}
	for deadline := time.Now().Add(300 * time.Millisecond); stillSource(first); {
		require.True(t, time.Now().Before(deadline), "DWM did not compose the prepared first frame")
		time.Sleep(4 * time.Millisecond)
		first, err = visualScreen(w, bg)
		require.NoError(t, err)
	}
	var fgBounds, bgBounds image.Rectangle
	runOnMain(func() {
		fgBounds, err = visualClientBounds(initialHWND)
		if err == nil {
			bgBounds, err = visualClientBounds(uintptr(unsafe.Pointer(bg.viewport.GetWin32Window())))
		}
	})
	require.NoError(t, err)
	require.False(t, bgBounds.Empty())
	require.True(t, bgBounds.In(fgBounds), "Acrylic client must fully cover the source client: %v over %v", fgBounds, bgBounds)
	directory := os.Getenv("FYNE_WINDOW_VISUALS_OUTPUT")
	require.NoError(t, saveVisualImage(directory, "first-show", first))
	assertVisualScreenAlpha(t, first)
	type sample struct {
		Mode                        desktop.WindowBackdrop
		Status                      desktop.VisualStatus
		RedBase, BlueBase           [3]float64
		BackgroundDelta, AlphaError float64
	}
	var samples []sample
	var lastModeBase [3]float64
	// In the diagnostic run, enqueue copies of both surfaces immediately
	// before the real Present. Map only after the first optical sample.
	var staging *graphicsCOM
	var producer *wglDXGIPresenter
	var realPresent, recordPresent uintptr
	var sampled bool
	if os.Getenv("FYNE_DWM_VALIDATE_BACKBUFFER") == "1" {
		recordPresent = syscall.NewCallback(func(chain, interval, flags uintptr) uintptr {
			copyDXGIPixel(producer, staging, 0, producer.shared, 20, 100)
			copyDXGIPixel(producer, staging, 1, producer.back, 20, 100)
			sampled = true
			result, _, _ := syscall.SyscallN(realPresent, chain, interval, flags)
			return result
		})
	}
	for _, mode := range []desktop.WindowBackdrop{desktop.BackdropMica, desktop.BackdropMicaAlt, desktop.BackdropDesktopAcrylic} {
		request, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: mode})
		require.NoError(t, err)
		require.Eventually(t, func() bool { s := w.VisualStatus(); return s.RequestID == request && s.Phase != desktop.VisualPending }, 2*time.Second, 10*time.Millisecond)
		require.Equal(t, desktop.VisualReady, w.VisualStatus().Phase)
		runOnMain(func() { usingAccent = w.presentation.acrylicAccent })
		require.Equal(t, mode == desktop.BackdropDesktopAcrylic && caps.OSBuild == acrylicBackdropWorkaroundBuild, usingAccent)
		colors := []color.NRGBA{red, blue}
		if mode == desktop.BackdropDesktopAcrylic {
			for len(colors) < 20 {
				colors = append(colors, red, blue)
			}
		}
		var screens [2]*image.NRGBA
		var previous [3]float64
		for i, c := range colors {
			var previousPresents uint64
			var diagnosticErr error
			runOnMain(func() {
				producer = bg.presentation.dxgi
				if producer != nil {
					previousPresents = producer.frames
				}
				if producer != nil && mode == desktop.BackdropDesktopAcrylic && recordPresent != 0 {
					staging, diagnosticErr = newDXGIPixelRow(producer, 2)
					if diagnosticErr == nil {
						sampled = false
						realPresent = producer.presentMethod
						defer func() { producer.presentMethod = realPresent }()
						producer.presentMethod = recordPresent
					}
				}
				bgContent.FillColor = c
				bgContent.Refresh()
				bg.RunWithContext(func() { d.repaintWindow(bg.window) })
			})
			time.Sleep(300 * time.Millisecond)
			screen, err := visualScreen(w, bg)
			if staging != nil {
				var pixels [2]color.NRGBA
				runOnMain(func() {
					defer releaseGraphicsCOM(&staging)
					if sampled && producer == bg.presentation.dxgi {
						diagnosticErr = readDXGIPixelRow(producer, staging, pixels[:])
					} else {
						diagnosticErr = fmt.Errorf("the original presenter did not submit a frame")
					}
				})
				require.NoError(t, diagnosticErr, "pre-Present readback")
				t.Logf("pre-Present background %d: expected=%v interop=%v DXGI buffer=%v", i, c, pixels[0], pixels[1])
				require.Equal(t, [2]color.NRGBA{c, c}, pixels)
			}
			require.NoError(t, diagnosticErr)
			require.NoError(t, err)
			if mode == desktop.BackdropDesktopAcrylic {
				// Check each change while the foreground is idle. A focus or
				// z-order change could refresh an otherwise stale backdrop.
				deadline := time.Now().Add(1500 * time.Millisecond)
				for {
					var got color.NRGBA
					var presented uint64
					var sampleErr error
					runOnMain(func() { got, presented, sampleErr = visualBackgroundColor(bg.window) })
					require.NoError(t, sampleErr)
					base := visualRGB(screen, 40, 75)
					source := [3]float64{float64(got.R), float64(got.G), float64(got.B)}
					want := [3]float64{float64(c.R), float64(c.G), float64(c.B)}
					// Use the direction of the red/blue change. A change in
					// tint alone can shift overall brightness.
					direction := math.Copysign(1, want[2]-want[0])
					change := ((base[2] - base[0]) - (previous[2] - previous[0])) * direction
					if presented > previousPresents && visualDelta(source, want) < 8 && (i == 0 || change > 10) {
						t.Logf("Acrylic background %d: source=%v backdrop=%v", i, source, base)
						previous = base
						break
					}
					if time.Now().After(deadline) {
						require.Greater(t, presented, previousPresents, "DXGI did not accept a new background frame")
						require.Less(t, visualDelta(source, want), 8.0, "background fixture did not update: got %v, want %v", source, want)
						t.Fatalf("Acrylic background %d stayed stale: source=%v backdrop=%v previous=%v", i, source, base, previous)
					}
					time.Sleep(40 * time.Millisecond)
					screen, err = visualScreen(w, bg)
					require.NoError(t, err)
				}
			}
			if i < len(screens) {
				screens[i] = screen
			}
			require.NoError(t, saveVisualImage(directory, fmt.Sprintf("mode-%d-background-%d", mode, i), screen))
		}
		s := sample{Mode: mode, Status: w.VisualStatus(), RedBase: visualRGB(screens[0], 40, 75), BlueBase: visualRGB(screens[1], 40, 75)}
		s.BackgroundDelta = visualDelta(s.RedBase, s.BlueBase)
		for i, alpha := range []float64{0, .25, .5, .75, 1} {
			base := visualRGB(screens[1], float64(40+i*60), 75)
			pixel := visualRGB(screens[1], float64(40+i*60), 40)
			for channel, source := range [3]float64{32, 176, 240} {
				s.AlphaError = math.Max(s.AlphaError, math.Abs(pixel[channel]-(alpha*source+(1-alpha)*base[channel])))
			}
		}
		t.Logf("DWM mode=%d red=%v blue=%v background_delta=%.2f alpha_error=%.2f", mode, s.RedBase, s.BlueBase, s.BackgroundDelta, s.AlphaError)
		samples = append(samples, s)
		lastModeBase = s.BlueBase
	}
	// Raw alpha and Acrylic have different optical behaviour. Checker contrast
	// supplies a blur gate, so live-background response alone cannot pass it.
	checker := visualChecker()
	runOnMain(func() {
		bg.window.SetContent(checker)
		bg.RunWithContext(func() { d.repaintWindow(bg.window) })
	})
	time.Sleep(350 * time.Millisecond)
	acrylic, err := visualScreen(w, bg)
	require.NoError(t, err)
	require.NoError(t, saveVisualImage(directory, "acrylic-checker", acrylic))
	id, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropTransparent})
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		s := w.VisualStatus()
		return s.RequestID == id && s.Phase == desktop.VisualReady
	}, 2*time.Second, 10*time.Millisecond)
	runOnMain(func() { usingAccent = w.presentation.acrylicAccent })
	require.False(t, usingAccent, "Transparent must remove the Acrylic policy")
	raw, err := visualScreen(w, bg)
	require.NoError(t, err)
	require.NoError(t, saveVisualImage(directory, "raw-transparent-checker", raw))
	acrylicDeviation, rawDeviation := visualDeviation(acrylic), visualDeviation(raw)
	t.Logf("checker contrast stddev: acrylic=%.3f raw-alpha=%.3f", acrylicDeviation, rawDeviation)
	// Exercise the automatic opaque fallback from an already active Acrylic.
	id, err = w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropDesktopAcrylic})
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		s := w.VisualStatus()
		return s.RequestID == id && s.Phase == desktop.VisualReady && s.Selected == desktop.BackdropDesktopAcrylic
	}, 2*time.Second, 10*time.Millisecond)
	contrast, err := theme.NewModern(theme.ModernOptions{HighContrast: true})
	require.NoError(t, err)
	var original granio.Theme
	runOnMain(func() {
		settings := granio.CurrentApp().Settings()
		original = settings.Theme()
		settings.SetTheme(contrast)
	})
	defer runOnMain(func() { granio.CurrentApp().Settings().SetTheme(original) })
	require.Eventually(t, func() bool {
		s := w.VisualStatus()
		return s.RequestID == id && s.SurfaceReady && s.Phase == desktop.VisualFallback &&
			s.Reason == desktop.VisualHighContrast && s.Selected == desktop.BackdropOpaque
	}, 2*time.Second, 10*time.Millisecond)
	runOnMain(func() { usingAccent = w.presentation.acrylicAccent })
	require.False(t, usingAccent, "opaque fallback must remove the Acrylic policy")
	assertVisualAlpha(t, visualFrame(w), false)
	time.Sleep(150 * time.Millisecond)
	opaque, err := visualScreen(w, bg)
	require.NoError(t, err)
	require.NoError(t, saveVisualImage(directory, "opaque", opaque))
	var finalHWND, finalStyle uintptr
	runOnMain(func() {
		finalHWND = uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
		finalStyle, _, _ = visualUser32.NewProc("GetWindowLongPtrW").Call(finalHWND, ^uintptr(15))
	})
	require.Equal(t, initialHWND, finalHWND)
	// Show legitimately adds WS_VISIBLE. Structural frame flags must remain.
	require.Equal(t, initialStyle&^uintptr(0x10000000), finalStyle&^uintptr(0x10000000))
	require.Equal(t, uintptr(0xc00000), finalStyle&uintptr(0xc00000), "native caption remains")
	if directory != "" {
		report := struct {
			Capabilities                                 desktop.VisualCapabilities
			Samples                                      []sample
			AcrylicCheckerDeviation, RawCheckerDeviation float64
			FirstShowBase, OpaqueBase                    [3]float64
			FirstFramePrepared, NativeWindowPreserved    bool
		}{caps, samples, acrylicDeviation, rawDeviation, visualRGB(first, 40, 75), visualRGB(opaque, 40, 75), true, initialHWND == finalHWND && initialStyle&^uintptr(0x10000000) == finalStyle&^uintptr(0x10000000)}
		data, err := json.MarshalIndent(report, "", "  ")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(directory, "dwm-composition.json"), append(data, '\n'), 0o644))
	}
	// These are optical acceptance gates, beyond surface and HRESULT checks.
	t.Run("first Acrylic", func(t *testing.T) {
		firstBase := visualRGB(first, 40, 75)
		t.Logf("first displayed Acrylic: %v, settled: %v", firstBase, samples[2].RedBase)
		require.Greater(t, firstBase[0]+firstBase[1]+firstBase[2], 10.0, "first shown frame cannot be black")
		require.LessOrEqual(t, visualDelta(firstBase, samples[2].RedBase), 4.0, "first shown frame already contains Acrylic")
	})
	for _, s := range samples {
		require.LessOrEqual(t, s.AlphaError, 8.0, "screen alpha composition mode %d", s.Mode)
		if s.Mode == desktop.BackdropMica || s.Mode == desktop.BackdropMicaAlt {
			require.Less(t, s.BackgroundDelta, 4.0, "Mica is wallpaper based")
		}
	}
	require.Greater(t, visualDelta(samples[0].BlueBase, samples[1].BlueBase), 1.0, "Mica Alt must have a distinct material")
	require.Greater(t, visualDelta(lastModeBase, visualRGB(opaque, 40, 75)), 3.0, "native material must be visible")
	require.Greater(t, rawDeviation, 30.0, "checker source must be visible through raw alpha")
	require.Less(t, acrylicDeviation, rawDeviation*.6, "Desktop Acrylic must reduce checker contrast")
	require.Greater(t, visualDelta(visualRGB(acrylic, 40, 75), lastModeBase), 5.0, "Desktop Acrylic must reflect the new checker source")
}

func TestNativeAcrylicFallbackWithoutAccent(t *testing.T) {
	requireVisualTest(t)
	var original *windows.LazyProc
	runOnMain(func() {
		original = setWindowCompositionAttribute
		setWindowCompositionAttribute = visualUser32.NewProc("GranioTestUnknownCompositionEntry")
	})
	defer runOnMain(func() { setWindowCompositionAttribute = original })

	w := newVisualTestWindow("Проверка резервного Desktop Acrylic")
	defer runOnMain(w.Close)
	_, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropDesktopAcrylic})
	require.NoError(t, err)
	runOnMain(func() {
		w.SetContent(visualAlphaScene())
		w.Resize(granio.NewSize(360, 290))
		w.Show()
	})
	requireComposition(t, w)
	if w.VisualCapabilities().OSBuild != acrylicBackdropWorkaroundBuild {
		t.Skip("this Windows build uses the system backdrop by default")
	}
	status := w.VisualStatus()
	require.Equal(t, desktop.VisualReady, status.Phase)
	require.Equal(t, desktop.BackdropDesktopAcrylic, status.Selected)
	require.True(t, status.NativeAccepted)
	require.True(t, status.FirstFrameBeforeShow)

	var state glfw.WindowVisualState
	var active bool
	runOnMain(func() {
		state, err = w.viewport.QueryWindowVisuals()
		active = w.presentation.acrylicAccent
	})
	require.NoError(t, err)
	require.False(t, active)
	require.True(t, state.BackdropKnown)
	require.Equal(t, glfw.BackdropTransientWindow, state.Backdrop)
	assertVisualAlpha(t, visualFrame(w), true)
}

func TestNativeAcrylicRemovalFailureDoesNotApplyMica(t *testing.T) {
	requireVisualTest(t)
	var originalAccent *windows.LazyProc
	var originalPower func() bool
	runOnMain(func() {
		originalAccent, originalPower = setWindowCompositionAttribute, windowsAcrylicPowerAllowed
		windowsAcrylicPowerAllowed = func() bool { return true }
	})
	defer runOnMain(func() {
		setWindowCompositionAttribute, windowsAcrylicPowerAllowed = originalAccent, originalPower
	})
	w := newVisualTestWindow("Проверка снятия Desktop Acrylic")
	defer runOnMain(w.Close)
	_, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropDesktopAcrylic})
	require.NoError(t, err)
	runOnMain(func() {
		w.SetContent(visualAlphaScene())
		w.Resize(granio.NewSize(360, 290))
		w.Show()
	})
	requireComposition(t, w)
	if w.VisualCapabilities().OSBuild != acrylicBackdropWorkaroundBuild {
		t.Skip("the legacy accent path is limited to the target build")
	}
	var active bool
	runOnMain(func() {
		active = w.presentation.acrylicAccent
		setWindowCompositionAttribute = visualUser32.NewProc("GranioTestUnknownCompositionEntry")
	})
	require.True(t, active)
	for _, failure := range []desktop.VisualFallbackPolicy{desktop.VisualFallbackOpaque, desktop.VisualFallbackFail} {
		id, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropMica, Fallback: failure})
		require.NoError(t, err)
		want := desktop.VisualFallback
		if failure == desktop.VisualFallbackFail {
			want = desktop.VisualFailed
		}
		require.Eventually(t, func() bool {
			s := w.VisualStatus()
			return s.RequestID == id && s.Phase == want && s.SurfaceReady && !s.NativeAccepted &&
				s.Selected == desktop.BackdropOpaque && s.Reason == desktop.VisualNativeFailure
		}, 2*time.Second, 10*time.Millisecond)
		var state glfw.WindowVisualState
		runOnMain(func() { state, err = w.viewport.QueryWindowVisuals() })
		require.NoError(t, err)
		require.Equal(t, glfw.BackdropNone, state.Backdrop, "Mica must wait for accent removal")
		assertVisualAlpha(t, visualFrame(w), false)
	}
	runOnMain(func() { setWindowCompositionAttribute = originalAccent })
	id, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropMica})
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		s := w.VisualStatus()
		return s.RequestID == id && s.Phase == desktop.VisualReady && s.Selected == desktop.BackdropMica
	}, 2*time.Second, 10*time.Millisecond)
	assertVisualAlpha(t, visualFrame(w), true)
}

func TestNativeAcrylicFollowsPowerPolicy(t *testing.T) {
	requireVisualTest(t)
	var original func() bool
	runOnMain(func() {
		original = windowsAcrylicPowerAllowed
		windowsAcrylicPowerAllowed = func() bool { return true }
	})
	defer runOnMain(func() { windowsAcrylicPowerAllowed = original })
	w := newVisualTestWindow("Проверка Acrylic при энергосбережении")
	defer runOnMain(w.Close)
	request, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropDesktopAcrylic})
	require.NoError(t, err)
	runOnMain(func() {
		w.SetContent(visualAlphaScene())
		w.Resize(granio.NewSize(360, 290))
		w.Show()
	})
	requireComposition(t, w)
	caps := w.VisualCapabilities()
	if caps.OSBuild != acrylicBackdropWorkaroundBuild {
		t.Skip("the Acrylic workaround is limited to build 26300")
	}

	checkBackdrop := func(wantAccent bool, want glfw.SystemBackdrop) {
		t.Helper()
		require.Eventually(t, func() bool {
			var applied bool
			runOnMain(func() {
				state, err := w.viewport.QueryWindowVisuals()
				applied = err == nil && state.BackdropKnown && state.Backdrop == want &&
					w.presentation.acrylicAccent == wantAccent
			})
			return applied
		}, 2*time.Second, 10*time.Millisecond, "native backdrop %v, accent=%t", want, wantAccent)
		status := w.VisualStatus()
		require.Equal(t, request, status.RequestID)
		require.Equal(t, desktop.VisualReady, status.Phase)
		require.True(t, status.SurfaceReady && status.NativeAccepted)
		require.Equal(t, caps.Generation, w.VisualCapabilities().Generation)
	}
	checkBackdrop(true, glfw.BackdropNone)
	// Substitute only the reported power state. Windows settings stay untouched.
	runOnMain(func() { windowsAcrylicPowerAllowed = func() bool { return false } })
	checkBackdrop(false, glfw.BackdropTransientWindow)
	runOnMain(func() { windowsAcrylicPowerAllowed = func() bool { return true } })
	checkBackdrop(true, glfw.BackdropNone)
}

func TestNativeWindowsBackdropTransitions(t *testing.T) {
	requireVisualTest(t)
	setVisualTestTheme(t)
	w := newVisualTestWindow("Проверка переходов между материалами Granio")
	defer runOnMain(w.Close)
	initial := desktop.BackdropTransparent
	request, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: initial})
	require.NoError(t, err)
	runOnMain(func() {
		w.SetPadded(false)
		w.SetContent(visualAlphaScene())
		w.Resize(granio.NewSize(360, 290))
		w.Show()
	})
	requireComposition(t, w)
	caps := w.VisualCapabilities()
	if !caps.OSBackdrops.Supports(desktop.BackdropDesktopAcrylic) || w.VisualStatus().Reason != desktop.VisualReasonNone {
		if os.Getenv("GRANIO_REQUIRE_DXGI") == "1" {
			t.Fatalf("native backdrop policy: %+v caps=%+v", w.VisualStatus(), caps)
		}
		t.Skipf("native backdrop policy: %+v caps=%+v", w.VisualStatus(), caps)
	}
	require.True(t, w.VisualStatus().FirstFrameBeforeShow)

	contrast, err := theme.NewModern(theme.ModernOptions{HighContrast: true})
	require.NoError(t, err)
	var original granio.Theme
	var hwnd uintptr
	runOnMain(func() {
		original = granio.CurrentApp().Settings().Theme()
		hwnd = uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
	})
	defer runOnMain(func() { granio.CurrentApp().Settings().SetTheme(original) })

	// Opaque represents a real policy fallback on the current request.
	// Each ordered pair must be exercised on the same prepared HWND.
	type material struct {
		backdrop desktop.WindowBackdrop
		native   glfw.SystemBackdrop
	}
	acrylic := material{desktop.BackdropDesktopAcrylic, glfw.BackdropTransientWindow}
	useAccent := caps.OSBuild == acrylicBackdropWorkaroundBuild && windowsAcrylicPowerAllowed()
	if useAccent {
		acrylic.native = glfw.BackdropNone
	}
	materials := []material{
		acrylic,
		{desktop.BackdropMica, glfw.BackdropMainWindow},
		{desktop.BackdropMicaAlt, glfw.BackdropTabbedWindow},
		{desktop.BackdropTransparent, glfw.BackdropNone},
		{desktop.BackdropOpaque, glfw.BackdropNone},
	}
	current, requested := initial, initial
	apply := func(target material) {
		if current == target.backdrop {
			return
		}
		from := current
		fallback := target.backdrop == desktop.BackdropOpaque
		phase, reason := desktop.VisualReady, desktop.VisualReasonNone
		if fallback {
			phase, reason = desktop.VisualFallback, desktop.VisualHighContrast
			runOnMain(func() { granio.CurrentApp().Settings().SetTheme(contrast) })
		} else {
			request, err = w.ConfigureVisuals(desktop.VisualOptions{Backdrop: target.backdrop})
			require.NoError(t, err)
			requested = target.backdrop
			if from == desktop.BackdropOpaque {
				runOnMain(func() { granio.CurrentApp().Settings().SetTheme(original) })
			}
		}
		require.Eventually(t, func() bool {
			s := w.VisualStatus()
			return s.RequestID == request && s.Requested == requested && s.Selected == target.backdrop &&
				s.Phase == phase && s.Reason == reason && s.SurfaceReady && s.NativeAccepted == !fallback
		}, 2*time.Second, 10*time.Millisecond, "%s → %s, status %+v", from, target.backdrop, w.VisualStatus())

		var state glfw.WindowVisualState
		var actualHWND uintptr
		var accent bool
		runOnMain(func() {
			state, err = w.viewport.QueryWindowVisuals()
			actualHWND = uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
			accent = w.presentation.acrylicAccent
		})
		require.NoError(t, err)
		require.Equal(t, hwnd, actualHWND, "%s → %s", from, target.backdrop)
		require.Equal(t, caps.Generation, w.VisualCapabilities().Generation)
		require.True(t, state.BackdropKnown)
		require.Equal(t, target.native, state.Backdrop, "%s → %s", from, target.backdrop)
		require.Equal(t, useAccent && target.backdrop == desktop.BackdropDesktopAcrylic, accent)
		assertVisualAlpha(t, visualFrame(w), !fallback)
		current = target.backdrop
	}
	for _, from := range materials {
		for _, to := range materials {
			if from == to {
				continue
			}
			apply(from)
			apply(to)
			t.Logf("%s → %s", from.backdrop, to.backdrop)
		}
	}
}
