//go:build !windows

package glfw

import (
	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/internal/scale"
)

func (*window) setDarkMode() {
}

func (w *window) computeCanvasSize(width, height int) granio.Size {
	return granio.NewSize(scale.ToFyneCoordinate(w.canvas, width), scale.ToFyneCoordinate(w.canvas, height))
}
