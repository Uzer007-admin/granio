//go:build !windows || wasm || test_web_driver

package glfw

import "github.com/Uzer007-admin/granio/v2/internal/painter/gl"

func softwareRendererRequested() bool { return false }

func (*window) softwareRendering() bool { return false }

func (*window) softwareCaptureUnavailable() bool { return false }

func newWindowPainter(w *window) gl.Painter { return gl.NewPainter(w.canvas, w) }

func (w *window) presentFrame() { w.viewport.SwapBuffers() }
