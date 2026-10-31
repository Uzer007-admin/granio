//go:build windows && !wasm && !test_web_driver

package glfw

import (
	"fmt"
	"image"
	"os"
	"runtime"
	"strings"
	"unsafe"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/painter/gl"
	"github.com/Uzer007-admin/granio/v2/internal/painter/software"
	"golang.org/x/sys/windows"
)

func softwareRendererRequested() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("GRANIO_RENDERER")), "software")
}

func (w *window) softwareRendering() bool { return w.software }

type windowPresentation struct {
	requested bool
	dxgi      *wglDXGIPresenter
	failed    bool
}

func (w *window) preparePresentation() {
	w.presentation.requested = !w.software && strings.EqualFold(strings.TrimSpace(os.Getenv("GRANIO_WGL_PRESENT")), "dxgi")
}

func (w *window) opaqueRendering() bool { return w.software || w.presentation.requested }

func (w *window) initPresentation() {
	if !w.presentation.requested {
		return
	}
	p, err := newWGLDXGIPresenter(uintptr(unsafe.Pointer(w.viewport.GetWin32Window())))
	if err != nil {
		granio.LogError("Failed to initialise WGL/DXGI presentation; using WGL", err)
		return
	}
	w.presentation.dxgi = p
}

func (w *window) releasePresentation() {
	if p := w.presentation.dxgi; p != nil {
		restore := w.activateContext()
		defer restore()
		p.destroy()
		w.presentation.dxgi = nil
	}
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
	if !w.software {
		if p := w.presentation.dxgi; p != nil {
			if w.viewport.GetAttrib(glfw.Iconified) == glfw.True {
				return
			}
			width, height := w.viewport.GetFramebufferSize()
			if err := p.present(width, height); err != nil && !w.presentation.failed {
				w.presentation.failed = true
				granio.LogError("Failed to present WGL/DXGI frame", err)
			}
			return
		}
		w.viewport.SwapBuffers()
		return
	}
	p := w.canvas.Painter().(*softwareWindowPainter)
	if err := presentSoftwareFrame(uintptr(unsafe.Pointer(w.viewport.GetWin32Window())), p.Frame()); err != nil {
		granio.LogError("Failed to present software frame", err)
	}
}
