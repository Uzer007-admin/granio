//go:build (linux || windows) && !android && !ci && !no_glfw && !mobile && !wasm && !test_web_driver

package glfw

import (
	"image"
	"image/color"
	"testing"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/driver"
	"github.com/stretchr/testify/require"
)

func TestNativeWindowsCanUseDifferentRenderModes(t *testing.T) {
	modes := [2]driver.RenderMode{driver.RenderSoftware, driver.RenderOpenGL}
	colors := [2]color.NRGBA{{R: 184, G: 61, B: 108, A: 255}, {R: 44, G: 152, B: 205, A: 255}}
	var windows [2]*window
	defer runOnMain(func() {
		for _, w := range windows {
			if w != nil {
				w.Close()
			}
		}
	})

	for i, mode := range modes {
		var setErr, lockedErr error
		var active driver.RenderMode
		var api int
		runOnMain(func() {
			w := d.CreateWindow("Granio render mode").(*window)
			windows[i] = w
			setErr = w.SetRenderMode(mode)
			if setErr != nil {
				return
			}
			w.SetPadded(false)
			w.SetContent(canvas.NewRectangle(colors[i]))
			w.Resize(granio.NewSize(120, 80))
			w.Show()
			active = w.RenderMode()
			api = w.viewport.GetAttrib(glfw.ClientAPI)
			lockedErr = w.SetRenderMode(driver.RenderAuto)
		})
		require.NoError(t, setErr)
		require.Equal(t, mode, active)
		require.Equal(t, mode == driver.RenderSoftware, api == glfw.NoAPI)
		require.ErrorIs(t, lockedErr, driver.ErrRenderModeLocked)
	}

	// Both color buffers must remain correct after creating the second window.
	for i, w := range windows {
		var frame image.Image
		var err error
		runOnMain(func() { frame, err = w.canvas.CaptureWithError() })
		require.NoError(t, err)
		require.GreaterOrEqual(t, frame.Bounds().Dx(), 120)
		require.GreaterOrEqual(t, frame.Bounds().Dy(), 80)
		c := color.NRGBAModel.Convert(frame.At(frame.Bounds().Dx()/2, frame.Bounds().Dy()/2))
		require.Equal(t, colors[i], c)
	}
}
