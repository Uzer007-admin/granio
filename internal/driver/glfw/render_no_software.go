//go:build (!linux && !windows) || android || wasm || test_web_driver

package glfw

import "github.com/Uzer007-admin/granio/v3/internal/painter/gl"

const softwareRendererAvailable = false

type windowPresentation struct{}

func (*window) releasePresentation() {}

func (*window) softwareRendering() bool { return false }

func (*window) softwareFrameSize() (int, int) { return 0, 0 }

func (*window) softwareCaptureUnavailable() bool { return false }

func newWindowPainter(w *window) gl.Painter { return gl.NewPainter(w.canvas, w) }

func (w *window) presentFrame() { w.viewport.SwapBuffers() }
