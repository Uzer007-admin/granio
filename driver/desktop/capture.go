package desktop

import (
	"errors"
	"image"
)

// ErrCaptureUnavailable means that no usable native surface/context exists,
// for example before a window is created or after it is closed.
var ErrCaptureUnavailable = errors.New("canvas capture: native surface unavailable")

// CanvasCapture is an optional native canvas capability. Unlike Capture, it
// returns framebuffer/allocation/readback failures to the caller. A successful
// capture has positive dimensions; unavailable images are never replaced with
// a fabricated 1x1 image or a stale frame.
type CanvasCapture interface {
	CaptureWithError() (image.Image, error)
}
