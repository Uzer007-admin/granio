//go:build windows && !wasm && !test_web_driver

package glfw

/*
#include <windows.h>
int granio_client_title_bar(HWND hwnd, int height, int leading, int trailing);
int granio_title_bar_bounds(HWND hwnd, int first, int second, int position);
int granio_title_bar_limits(HWND hwnd, int min_width, int min_height, int max_width, int max_height);
*/
import "C"

import (
	"errors"
	"math"
	"unsafe"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/driver/desktop"
	"github.com/Uzer007-admin/granio/v3/internal/async"
)

var _ desktop.ClientTitleBarWindow = (*window)(nil)

func (w *window) SetClientTitleBar(options desktop.ClientTitleBarOptions) error {
	if !w.decorate {
		return desktop.ErrClientTitleBarUnavailable
	}
	if w.created || w.viewport != nil || w.windowClosed() {
		return errors.New("configure the client title bar before showing the window")
	}
	if !validScale(options.Height) || !validTitleBarInset(options.LeadingReserved) || !validTitleBarInset(options.TrailingReserved) {
		return errors.New("invalid client title bar dimensions")
	}
	w.clientTitleBar = options
	return nil
}

func validTitleBarInset(value float32) bool {
	return value == 0 || validScale(value)
}

func (w *window) updateClientTitleBar() {
	options := w.clientTitleBar
	if options.Height == 0 || w.view() == nil {
		return
	}
	factor := float64(w.canvas.Scale())
	pixels := func(value float32) C.int {
		return C.int(min(math.Round(float64(value)*factor), math.MaxInt32))
	}
	hwnd := C.HWND(unsafe.Pointer(w.viewport.GetWin32Window()))
	w.clientTitleBarActive = C.granio_client_title_bar(hwnd, max(1, pixels(options.Height)),
		pixels(options.LeadingReserved), pixels(options.TrailingReserved)) != 0
	if !w.clientTitleBarActive {
		granio.LogError("Failed to configure the client title bar", desktop.ErrClientTitleBarUnavailable)
	}
}

func (w *window) ClientTitleBarActive() bool { return w.clientTitleBarActive }

func (w *window) changeTitleBarBounds(first, second int, position bool) bool {
	if !w.clientTitleBarActive || w.FullScreen() {
		return false
	}
	hwnd := C.HWND(unsafe.Pointer(w.viewport.GetWin32Window()))
	var move C.int
	if position {
		move = 1
	}
	return C.granio_title_bar_bounds(hwnd, C.int(first), C.int(second), move) != 0
}

func (w *window) setTitleBarLimits(minWidth, minHeight, maxWidth, maxHeight int) bool {
	if !w.clientTitleBarActive {
		return false
	}
	hwnd := C.HWND(unsafe.Pointer(w.viewport.GetWin32Window()))
	return C.granio_title_bar_limits(hwnd, C.int(minWidth), C.int(minHeight), C.int(maxWidth), C.int(maxHeight)) != 0
}

func (w *window) Minimize() {
	async.EnsureMain(func() {
		if view := w.view(); view != nil {
			view.Iconify()
		}
	})
}

func (w *window) ToggleMaximize() {
	async.EnsureMain(func() {
		if view := w.view(); view != nil && !w.FullScreen() && !w.fixedSize {
			if w.Maximized() {
				view.Restore()
			} else {
				view.Maximize()
			}
		}
	})
}

func (w *window) Maximized() bool {
	return w.view() != nil && w.viewport.GetAttrib(glfw.Maximized) == glfw.True
}
