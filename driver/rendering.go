package driver

import "errors"

// RenderMode selects how Granio draws application windows.
type RenderMode uint8

const (
	// RenderAuto uses the platform default and honors GRANIO_RENDERER. On
	// Windows and Linux it can fall back to CPU drawing at the first window.
	RenderAuto RenderMode = iota
	// RenderOpenGL uses the platform GL or GLES implementation. Acceleration
	// depends on the graphics driver installed on the device.
	RenderOpenGL
	// RenderSoftware draws the application canvas on the CPU.
	RenderSoftware
)

var (
	// ErrRenderModeUnavailable indicates that the driver cannot use the requested mode.
	ErrRenderModeUnavailable = errors.New("render mode unavailable")
	// ErrRenderModeLocked indicates that the renderer has already been selected.
	ErrRenderModeLocked = errors.New("render mode is already locked")
)

// RenderingDriver is an optional interface implemented by native drivers.
type RenderingDriver interface {
	// SetRenderMode must be called on the main goroutine before NewWindow or Run.
	SetRenderMode(RenderMode) error
	// RenderMode reports the drawing path selected for new windows.
	RenderMode() RenderMode
}

// RenderingWindow is implemented by desktop windows with individual render mode selection.
type RenderingWindow interface {
	// SetRenderMode must be called on the main goroutine before the first Show.
	// RenderAuto uses the driver default.
	SetRenderMode(RenderMode) error
	// RenderMode reports the window's selected drawing path. Call it on the
	// main goroutine.
	RenderMode() RenderMode
}
