package desktop

import "errors"

// ErrClientTitleBarUnavailable means the window cannot use a client title bar.
var ErrClientTitleBarUnavailable = errors.New("client title bar unavailable")

// ClientTitleBarOptions describes the application's title area in logical
// canvas units. The two reserved widths keep controls at either end clickable.
type ClientTitleBarOptions struct {
	Height           float32
	LeadingReserved  float32
	TrailingReserved float32
}

// ClientTitleBarWindow is an optional extension for windows that can draw
// their own title bar while retaining system moving and resizing.
type ClientTitleBarWindow interface {
	// SetClientTitleBar configures a decorated window before its first Show.
	// Call it on the UI thread. The area between the reserved widths is draggable.
	SetClientTitleBar(ClientTitleBarOptions) error
	// ClientTitleBarActive reports whether the native frame was configured.
	// Call it on the UI thread after Show.
	ClientTitleBarActive() bool
	// Minimize asks the window manager to iconify the window.
	Minimize()
	// ToggleMaximize maximizes or restores the window.
	ToggleMaximize()
	// Maximized reports the current native state. Call it on the UI thread.
	Maximized() bool
}
