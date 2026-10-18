//go:build windows && !no_glfw && !wasm && !test_web_driver && !mobile

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
	"testing"
	"time"
	"unsafe"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
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
	foreground, _, _ := visualUser32.NewProc("GetForegroundWindow").Call()
	if foreground != hwnd {
		return nil, fmt.Errorf("our Fyne test window 0x%x is not foreground (0x%x)", hwnd, foreground)
	}
	var rect struct{ left, top, right, bottom int32 }
	ok, _, err := visualUser32.NewProc("GetClientRect").Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	if ok == 0 {
		return nil, err
	}
	var origin struct{ x, y int32 }
	ok, _, err = visualUser32.NewProc("ClientToScreen").Call(hwnd, uintptr(unsafe.Pointer(&origin)))
	if ok == 0 {
		return nil, err
	}
	width, height := int(rect.right), int(rect.bottom)
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
		intersects := area.left < origin.x+int32(width) && area.right > origin.x && area.top < origin.y+int32(height) && area.bottom > origin.y
		if visible != 0 && intersects {
			break
		}
		under, _, _ = visualUser32.NewProc("GetWindow").Call(under, 2)
	}
	if under != background {
		return nil, fmt.Errorf("owned background 0x%x is not immediately behind our client (0x%x)", background, under)
	}
	for y := 4; y < height-4; y += 16 {
		for x := 4; x < width-4; x += 16 {
			point := uint64(uint32(origin.x+int32(x))) | uint64(uint32(origin.y+int32(y)))<<32
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
	ok, _, err = visualGDI32.NewProc("BitBlt").Call(memory, 0, 0, uintptr(width), uintptr(height), dc, uintptr(origin.x), uintptr(origin.y), 0x00cc0020)
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
		focusVisualTestWindow(w)
		visualDWM.NewProc("DwmFlush").Call()
		img, err = captureVisualClient(uintptr(unsafe.Pointer(w.viewport.GetWin32Window())), uintptr(unsafe.Pointer(bg.viewport.GetWin32Window())))
	})
	return img, err
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

// Checks optical DWM output with two owned Fyne windows. In particular an
// accepted HRESULT does not satisfy the Acrylic and alpha composition gates.
func TestNativeWindowVisualsDWMComposition(t *testing.T) {
	requireVisualTest(t)
	bg := createWindowWithResizeCallback("Fyne owned DWM test background")
	defer bg.Close()
	var bgContent *canvas.Rectangle
	runOnMain(func() {
		bgContent = canvas.NewRectangle(color.NRGBA{R: 220, G: 40, B: 30, A: 255})
		bg.window.SetPadded(false)
		bg.window.SetContent(bgContent)
		bg.window.Resize(granio.NewSize(780, 580))
		bg.window.Show()
		bg.viewport.SetPos(80, 80)
		bg.viewport.SetAttrib(glfw.Floating, glfw.True)
	})
	w := newVisualTestWindow("Fyne WindowVisuals WGL/DWM")
	defer runOnMain(w.Close)
	_, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropMica})
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
		w.RunWithContext(func() { hidden = w.canvas.Painter().Capture(w.canvas) })
	})
	caps := w.VisualCapabilities()
	if !caps.OSBackdrops.Supports(desktop.BackdropMica) || w.VisualStatus().Reason != desktop.VisualReasonNone {
		t.Skipf("native policy: %+v caps=%+v", w.VisualStatus(), caps)
	}
	assertVisualAlpha(t, hidden, true)
	require.True(t, w.VisualStatus().FirstFrameBeforeShow)
	runOnMain(func() { w.Show(); focusVisualTestWindow(w) })
	first, err := visualScreen(w, bg)
	require.NoError(t, err)
	directory := os.Getenv("FYNE_WINDOW_VISUALS_OUTPUT")
	require.NoError(t, saveVisualImage(directory, "first-show", first))
	type sample struct {
		Mode                        desktop.WindowBackdrop
		Status                      desktop.VisualStatus
		RedBase, BlueBase           [3]float64
		BackgroundDelta, AlphaError float64
	}
	var samples []sample
	var lastModeBase [3]float64
	for _, mode := range []desktop.WindowBackdrop{desktop.BackdropMica, desktop.BackdropMicaAlt, desktop.BackdropDesktopAcrylic} {
		request, err := w.ConfigureVisuals(desktop.VisualOptions{Backdrop: mode})
		require.NoError(t, err)
		require.Eventually(t, func() bool { s := w.VisualStatus(); return s.RequestID == request && s.Phase != desktop.VisualPending }, 2*time.Second, 10*time.Millisecond)
		require.Equal(t, desktop.VisualReady, w.VisualStatus().Phase)
		var screens [2]*image.NRGBA
		for i, c := range []color.NRGBA{{R: 220, G: 40, B: 30, A: 255}, {R: 30, G: 55, B: 225, A: 255}} {
			runOnMain(func() {
				bgContent.FillColor = c
				bgContent.Refresh()
				bg.RunWithContext(func() { d.repaintWindow(bg.window) })
				focusVisualTestWindow(w)
			})
			time.Sleep(300 * time.Millisecond)
			screens[i], err = visualScreen(w, bg)
			require.NoError(t, err)
			require.NoError(t, saveVisualImage(directory, fmt.Sprintf("mode-%d-background-%d", mode, i), screens[i]))
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
	checker := canvas.NewRasterWithPixels(func(x, y, width, height int) color.Color {
		if (x/16+y/16)%2 == 0 {
			return color.NRGBA{R: 250, G: 250, B: 250, A: 255}
		}
		return color.NRGBA{R: 12, G: 12, B: 12, A: 255}
	})
	runOnMain(func() {
		bg.window.SetContent(checker)
		bg.RunWithContext(func() { d.repaintWindow(bg.window) })
		focusVisualTestWindow(w)
	})
	time.Sleep(350 * time.Millisecond)
	acrylic, err := visualScreen(w, bg)
	require.NoError(t, err)
	require.NoError(t, saveVisualImage(directory, "acrylic-checker", acrylic))
	runOnMain(func() {
		_, err = w.viewport.ApplyWindowVisuals(glfw.WindowVisualOptions{Attributes: glfw.VisualBackdrop, Backdrop: glfw.BackdropNone})
	})
	require.NoError(t, err)
	time.Sleep(200 * time.Millisecond)
	raw, err := visualScreen(w, bg)
	require.NoError(t, err)
	require.NoError(t, saveVisualImage(directory, "raw-transparent-checker", raw))
	acrylicDeviation, rawDeviation := visualDeviation(acrylic), visualDeviation(raw)
	t.Logf("checker contrast stddev: acrylic=%.3f raw-alpha=%.3f", acrylicDeviation, rawDeviation)
	// Restore through the public interface before the opaque fallback check.
	_, err = w.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropOpaque, PrepareAlpha: true})
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		return w.VisualStatus().Phase == desktop.VisualReady && w.VisualStatus().Requested == desktop.BackdropOpaque
	}, 2*time.Second, 10*time.Millisecond)
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
	for _, s := range samples {
		require.LessOrEqual(t, s.AlphaError, 8.0, "screen alpha composition mode %d", s.Mode)
		if s.Mode == desktop.BackdropDesktopAcrylic {
			require.Greater(t, s.BackgroundDelta, 5.0, "Acrylic must use the live background")
		}
		if s.Mode == desktop.BackdropMica || s.Mode == desktop.BackdropMicaAlt {
			require.Less(t, s.BackgroundDelta, 4.0, "Mica is wallpaper based")
		}
	}
	require.Greater(t, visualDelta(samples[0].BlueBase, samples[1].BlueBase), 1.0, "Mica Alt must have a distinct material")
	require.Greater(t, visualDelta(lastModeBase, visualRGB(opaque, 40, 75)), 3.0, "native material must be visible")
	require.Greater(t, rawDeviation, 30.0, "checker source must be visible through raw alpha")
	require.Less(t, acrylicDeviation, rawDeviation*.6, "Desktop Acrylic must reduce checker contrast")
	firstBase := visualRGB(first, 40, 75)
	require.Greater(t, firstBase[0]+firstBase[1]+firstBase[2], 10.0, "first shown frame cannot be black")
	require.LessOrEqual(t, visualDelta(firstBase, samples[0].RedBase), 2.0, "first shown frame already contains Mica")
}
