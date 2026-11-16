//go:build (linux || windows) && !android && !wasm && !test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v3/internal/painter/gl"
	"github.com/Uzer007-admin/granio/v3/internal/painter/software"
)

const softwareRendererAvailable = true

func (w *window) softwareRendering() bool { return w.software }

func (w *window) softwareFrameSize() (int, int) { return w.framebuffer.X, w.framebuffer.Y }

func (w *window) softwareCaptureUnavailable() bool {
	if !w.software {
		return false
	}
	return w.framebuffer.X <= 0 || w.framebuffer.Y <= 0 || w.viewport.GetAttrib(glfw.Iconified) != glfw.False
}

type softwareWindowPainter struct{ *software.FramePainter }

func (*softwareWindowPainter) SetWindowSurface(mode gl.WindowSurfaceMode) bool {
	return mode == gl.WindowSurfaceLegacy || mode == gl.WindowSurfaceOpaque
}
func (*softwareWindowPainter) WindowSurfaceAlphaBits() int { return 0 }

func newWindowPainter(w *window) gl.Painter {
	if w.software {
		return &softwareWindowPainter{software.NewFramePainter(w.canvas)}
	}
	return gl.NewPainter(w.canvas, w)
}
