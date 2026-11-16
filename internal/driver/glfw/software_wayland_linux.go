//go:build linux && !android && !test_web_driver && (!x11 || wayland)

package glfw

// #include "software_linux.h"
import "C"

import (
	"unsafe"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
)

func (w *window) blitWayland(pixels unsafe.Pointer, width, height, stride int) int {
	if w.viewport.GetAttrib(glfw.Visible) == glfw.False {
		return C.GRANIO_SOFTWARE_DEFERRED
	}
	display := unsafe.Pointer(glfw.GetWaylandDisplay())
	surface := unsafe.Pointer(w.viewport.GetWaylandWindow())
	surfaceWidth, surfaceHeight := w.viewport.GetSize()
	return int(C.granioPresentSoftwareWayland(w.presentation.software, display, surface,
		(*C.uchar)(pixels), C.int(width), C.int(height), C.int(stride), C.int(surfaceWidth), C.int(surfaceHeight)))
}
