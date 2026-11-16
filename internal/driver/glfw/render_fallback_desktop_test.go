//go:build (linux || windows) && !android && !wasm && !test_web_driver

package glfw

import (
	"errors"
	"testing"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v3/driver"
	"github.com/stretchr/testify/require"
)

func TestSoftwareFallbackOnFirstContext(t *testing.T) {
	t.Setenv("GRANIO_RENDERER", "auto")
	d := &gLDriver{}
	d.lockRenderMode()
	w := &window{driver: d}
	w.selectRenderer()

	require.False(t, w.fallbackToSoftware(errors.New("display connection failed")))
	require.True(t, w.fallbackToSoftware(&glfw.Error{Code: glfw.FormatUnavailable}))
	require.Equal(t, driver.RenderSoftware, d.RenderMode())
	require.True(t, w.softwareRendering())
	require.False(t, (&window{driver: d}).fallbackToSoftware(errOpenGLRendererInit))
}

func TestExplicitOpenGLDisablesSoftwareFallback(t *testing.T) {
	t.Setenv("GRANIO_RENDERER", " OpenGL ")
	d := &gLDriver{}
	d.lockRenderMode()
	t.Setenv("GRANIO_RENDERER", "software")
	require.Equal(t, driver.RenderOpenGL, d.RenderMode())
	require.False(t, (&window{driver: d}).fallbackToSoftware(errOpenGLRendererInit))

	d = &gLDriver{}
	require.NoError(t, d.SetRenderMode(driver.RenderOpenGL))
	d.lockRenderMode()
	require.Equal(t, driver.RenderOpenGL, d.RenderMode())
	require.False(t, (&window{driver: d}).fallbackToSoftware(errOpenGLRendererInit))
}
