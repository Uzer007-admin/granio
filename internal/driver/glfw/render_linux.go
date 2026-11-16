//go:build linux && !android && !wasm && !test_web_driver

package glfw

/*
#cgo CFLAGS: -D_GNU_SOURCE
#cgo linux,!wayland CFLAGS: -DGRANIO_X11
#cgo linux,x11,wayland CFLAGS: -DGRANIO_X11
#cgo linux,!x11 CFLAGS: -DGRANIO_WAYLAND
#cgo linux,x11,wayland CFLAGS: -DGRANIO_WAYLAND
#cgo linux,!wayland pkg-config: x11
#cgo linux,x11,wayland pkg-config: x11
#cgo linux,!x11 pkg-config: wayland-client
#cgo linux,x11,wayland pkg-config: wayland-client
#include "software_linux.h"
*/
import "C"

import (
	"errors"
	"image"
	"runtime"
	"unsafe"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/internal/build"
)

type windowPresentation struct {
	software *C.GranioSoftwareSurface
}

var errLinuxSoftwareFrame = errors.New("unable to present native software framebuffer")

func (w *window) releasePresentation() {
	C.granioDestroySoftwareSurface(w.presentation.software)
	w.presentation.software = nil
}

func (w *window) presentSoftwareFrame(frame *image.NRGBA) (bool, error) {
	width, height := frame.Rect.Dx(), frame.Rect.Dy()
	if width <= 0 || height <= 0 || width > 16384 || height > 16384 || frame.Stride != width*4 {
		return false, errLinuxSoftwareFrame
	}
	if w.presentation.software == nil {
		w.presentation.software = C.granioCreateSoftwareSurface()
		if w.presentation.software == nil {
			return false, errLinuxSoftwareFrame
		}
	}
	pixels := unsafe.Pointer(unsafe.SliceData(frame.Pix))
	var result int
	if build.IsWayland {
		result = w.blitWayland(pixels, width, height, frame.Stride)
	} else {
		result = w.blitX11(pixels, width, height, frame.Stride)
	}
	runtime.KeepAlive(frame)
	switch result {
	case C.GRANIO_SOFTWARE_PRESENTED:
		return true, nil
	case C.GRANIO_SOFTWARE_DEFERRED:
		return false, nil
	default:
		return false, errLinuxSoftwareFrame
	}
}

func (w *window) presentFrame() {
	if !w.software {
		w.viewport.SwapBuffers()
		return
	}

	frame := w.canvas.Painter().(*softwareWindowPainter).Frame()
	if frame == nil {
		w.frame.markReady()
		return
	}
	presented, err := w.presentSoftwareFrame(frame)
	if presented {
		return
	}
	w.frame.markReady()
	if err != nil {
		granio.LogError("Failed to present software frame", err)
		return
	}
	w.canvas.SetDirty() // Retry after a configure or buffer release.
}
