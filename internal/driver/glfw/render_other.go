//go:build !windows || wasm || test_web_driver

package glfw

import "github.com/Uzer007-admin/granio/v2/internal/painter/gl"

type windowPresentation struct{}

func (*window) preparePresentationFirstFrame() {}

func (*window) preparePresentation()  {}
func (*window) initPresentation()     {}
func (*window) releasePresentation()  {}
func (*window) opaqueRendering() bool { return false }

func (*window) presentationMaxTextureSize() int { return 0 }

func softwareRendererRequested() bool { return false }

func (*window) softwareRendering() bool { return false }

func (*window) softwareCaptureUnavailable() bool { return false }

func newWindowPainter(w *window) gl.Painter { return gl.NewPainter(w.canvas, w) }

func (w *window) presentFrame() { w.viewport.SwapBuffers() }
