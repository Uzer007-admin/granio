//go:build windows && !wasm && !test_web_driver

package glfw

import (
	"fmt"
	"image"
	"os"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal/painter/gl"
	"github.com/Uzer007-admin/granio/v2/internal/painter/software"
	"golang.org/x/sys/windows"
)

func softwareRendererRequested() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("GRANIO_RENDERER")), "software")
}

func (w *window) softwareRendering() bool { return w.software }

type windowPresentation struct {
	requested         bool
	automatic         bool
	dxgi              *wglDXGIPresenter
	failure           error
	recoveryAttempted bool
	checkComposition  bool
	retryAfter        time.Time
	noRedirection     bool
	alphaUnavailable  bool
}

func (w *window) preparePresentation() {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("GRANIO_WGL_PRESENT")))
	w.presentation.requested = !w.software && mode == "dxgi"
	w.presentation.automatic = !w.software && (mode == "" || mode == "auto")
}

func (w *window) opaqueRendering() bool { return w.software }

func (w *window) preparePresentationWindow(alpha, desktopGL bool) {
	w.presentation.noRedirection = alpha && desktopGL && !w.presentation.alphaUnavailable && (w.presentation.requested || w.presentation.automatic)
	value := glfw.False
	if w.presentation.noRedirection {
		value = glfw.True
	}
	glfw.WindowHint(glfw.Win32NoRedirectionBitmap, value)
}

func (*window) resetPresentationWindowHint() {
	glfw.WindowHint(glfw.Win32NoRedirectionBitmap, glfw.False)
}

func (w *window) retryPresentationWindow(creationFailed bool) bool {
	if !w.presentation.noRedirection || !creationFailed && w.presentation.dxgi != nil && w.presentation.dxgi.composition != nil {
		return false
	}
	w.presentation.alphaUnavailable = true
	w.presentation.noRedirection = false
	return true
}

func (w *window) presentationReady() bool {
	if w.presentation.failure != nil {
		return false
	}
	p := w.presentation.dxgi
	return p == nil || p.shared != nil && (p.composition == nil || p.composition.attached)
}

func (w *window) presentationFailure() (uint32, bool) {
	native, _ := w.presentation.failure.(graphicsError)
	return native.code, w.presentation.failure != nil
}

func (w *window) requestPresentationCheck() {
	if w.presentation.noRedirection {
		w.presentation.checkComposition = true
	}
}

func (w *window) presentationMaxTextureSize() int {
	if p := w.presentation.dxgi; p != nil {
		return p.maxDimension
	}
	return 0
}

func (w *window) initPresentation() {
	if !w.presentation.requested && !w.presentation.automatic {
		return
	}
	if w.presentation.automatic {
		// GLES и драйверы без interop не создают устройство D3D11.
		if w.viewport.GetAttrib(glfw.ClientAPI) != glfw.OpenGLAPI || !glfw.ExtensionSupported("WGL_NV_DX_interop2") {
			return
		}
	}
	hwnd := uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
	p, err := newWGLDXGIPresenter(hwnd, w.presentation.noRedirection)
	if err != nil {
		if w.presentation.failure != nil {
			granio.LogError("Failed to restore WGL/DXGI presentation", err)
		} else if w.presentation.noRedirection {
			granio.LogError("Failed to initialise DirectComposition; recreating an opaque window", err)
		} else {
			granio.LogError("Failed to initialise WGL/DXGI presentation; using WGL", err)
		}
		return
	}
	w.presentation.dxgi = p
	w.presentation.failure = nil
}

func (w *window) releasePresentation() {
	if p := w.presentation.dxgi; p != nil {
		restore := w.activateContext()
		defer restore()
		p.destroy()
		w.presentation.dxgi = nil
	}
}

func (w *window) preparePresentationFirstFrame() {
	if w.visible || w.presentationReady() {
		return
	}
	w.RunWithContext(func() {
		for range 2 {
			w.driver.repaintWindow(w)
			w.presentFrame()
			if w.presentationReady() || w.presentation.failure != nil {
				break
			}
		}
	})
}

func (w *window) softwareCaptureUnavailable() bool {
	if !w.software {
		return false
	}
	width, height := w.viewport.GetFramebufferSize()
	return width <= 0 || height <= 0 || w.viewport.GetAttrib(glfw.Iconified) != glfw.False
}

type softwareWindowPainter struct{ *software.FramePainter }

func (*softwareWindowPainter) SetWindowSurface(mode gl.WindowSurfaceMode) bool {
	return mode == gl.WindowSurfaceLegacy || mode == gl.WindowSurfaceOpaque
}
func (*softwareWindowPainter) WindowSurfaceAlphaBits() int { return 0 }

func newWindowPainter(w *window) gl.Painter {
	if w.software {
		return &softwareWindowPainter{software.NewFramePainter(w.canvas)}
	}
	return gl.NewPainter(w.canvas, w)
}

var (
	softwareUser32    = windows.NewLazySystemDLL("user32.dll")
	softwareGDI32     = windows.NewLazySystemDLL("gdi32.dll")
	softwareGetDC     = softwareUser32.NewProc("GetDC")
	softwareReleaseDC = softwareUser32.NewProc("ReleaseDC")
	softwareSetDIBits = softwareGDI32.NewProc("SetDIBitsToDevice")
)

type softwareBitmapInfo struct {
	size                             uint32
	width, height                    int32
	planes, bits                     uint16
	compression, imageBytes          uint32
	xPixelsPerMeter, yPixelsPerMeter int32
	colorsUsed, colorsImportant      uint32
	redMask, greenMask, blueMask     uint32
}

// Маски позволяют GDI читать переиспользуемый NRGBA-кадр напрямую,
// без второго буфера и копирования всех пикселей из RGBA в BGRA.
func presentSoftwareFrame(hwnd uintptr, frame *image.NRGBA) error {
	if frame == nil || frame.Bounds().Empty() {
		return nil
	}
	dc, _, err := softwareGetDC.Call(hwnd)
	if dc == 0 {
		return fmt.Errorf("software renderer GetDC: %w", err)
	}
	defer softwareReleaseDC.Call(hwnd, dc)
	return drawSoftwareFrame(dc, frame)
}

func drawSoftwareFrame(dc uintptr, frame *image.NRGBA) error {
	if frame == nil || frame.Bounds().Empty() {
		return nil
	}
	width, height := frame.Bounds().Dx(), frame.Bounds().Dy()
	if frame.Rect.Min != (image.Point{}) || frame.Stride != width*4 || len(frame.Pix) < frame.Stride*height {
		return fmt.Errorf("software renderer requires a packed frame")
	}
	info := softwareBitmapInfo{size: 40, width: int32(width), height: -int32(height), planes: 1,
		bits: 32, compression: 3, imageBytes: uint32(len(frame.Pix)),
		redMask: 0xff, greenMask: 0xff00, blueMask: 0xff0000}
	lines, _, err := softwareSetDIBits.Call(dc, 0, 0, uintptr(width), uintptr(height), 0, 0, 0, uintptr(height),
		uintptr(unsafe.Pointer(&frame.Pix[0])), uintptr(unsafe.Pointer(&info)), 0)
	runtime.KeepAlive(frame)
	if lines == 0 {
		return fmt.Errorf("software renderer SetDIBitsToDevice: %w", err)
	}
	return nil
}

func (w *window) presentFrame() {
	if w.software {
		p := w.canvas.Painter().(*softwareWindowPainter)
		if err := presentSoftwareFrame(uintptr(unsafe.Pointer(w.viewport.GetWin32Window())), p.Frame()); err != nil {
			granio.LogError("Failed to present software frame", err)
		}
		return
	}

	state := &w.presentation
	p := state.dxgi
	if p == nil && state.failure == nil {
		w.viewport.SwapBuffers()
		return
	}
	if state.failure != nil && time.Now().Before(state.retryAfter) {
		return
	}
	if w.viewport.GetAttrib(glfw.Iconified) == glfw.True {
		return
	}
	width, height := w.viewport.GetFramebufferSize()
	if width <= 0 || height <= 0 || p != nil && (width > p.maxDimension || height > p.maxDimension) {
		return // Wait for usable geometry before allocating a new frame.
	}
	if state.failure != nil {
		state.recoveryAttempted = false
		w.recoverPresentation(state.failure)
		return
	}
	if state.checkComposition {
		state.checkComposition = false
		if err := p.checkCompositionDevice(); err != nil {
			w.recoverPresentation(err)
			return
		}
	}
	before := p.frames
	if err := p.present(width, height); err != nil {
		w.recoverPresentation(err)
		return
	}
	if p.frames != before {
		state.recoveryAttempted = false
		if !state.retryAfter.IsZero() {
			state.retryAfter = time.Time{}
			if w.visuals.active.Load() {
				w.queueVisualApply()
			}
		}
	}
}

func (w *window) recoverPresentation(err error) {
	state := &w.presentation
	state.failure = err
	granio.LogError("Failed to present WGL/DXGI frame", err)
	if !state.recoveryAttempted {
		state.recoveryAttempted = true
		w.releasePresentation()
		w.initPresentation()
		if state.dxgi != nil {
			w.canvas.SetDirty()
			return
		}
	}
	// Delay another attempt even if the application keeps requesting frames.
	state.retryAfter = time.Now().Add(time.Second)
	_, status, enabled := w.visualRequest()
	if enabled && status.Phase != desktop.VisualPending {
		status.Phase, status.Reason = desktop.VisualFailed, desktop.VisualNativeFailure
		status.NativeCode, _ = w.presentationFailure()
		status.SurfaceReady = false
		w.publishVisualStatus(status, w.VisualCapabilities())
	}
}
