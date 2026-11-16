//go:build linux && !android && !test_web_driver && wayland && !x11

package glfw

// #include "software_linux.h"
import "C"

import "unsafe"

func (*window) blitX11(unsafe.Pointer, int, int, int) int { return C.GRANIO_SOFTWARE_UNSUPPORTED }
