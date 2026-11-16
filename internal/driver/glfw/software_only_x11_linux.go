//go:build linux && !android && !test_web_driver && x11 && !wayland

package glfw

// #include "software_linux.h"
import "C"

import "unsafe"

func (*window) blitWayland(unsafe.Pointer, int, int, int) int { return C.GRANIO_SOFTWARE_UNSUPPORTED }
