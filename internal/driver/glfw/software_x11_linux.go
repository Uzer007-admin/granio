//go:build linux && !android && !test_web_driver && (!wayland || x11)

package glfw

// #include "software_linux.h"
import "C"

import (
	"unsafe"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
)

func (w *window) blitX11(pixels unsafe.Pointer, width, height, stride int) int {
	display := unsafe.Pointer(glfw.GetX11Display())
	id := C.uintptr_t(w.viewport.GetX11Window())
	return int(C.granioPresentSoftwareX11(w.presentation.software, display, id,
		(*C.uchar)(pixels), C.int(width), C.int(height), C.int(stride)))
}
