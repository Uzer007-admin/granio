//go:build !no_glfw && !mobile && !wasm && !test_web_driver

package glfw

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/stretchr/testify/require"
)

func TestExactScaleMatrix(t *testing.T) {
	for _, system := range []float32{1, 1.25, 1.5, 1.75, 2, 2.25, 3, 4} {
		for _, user := range []float32{0.8, 1, 1.2, 1.37} {
			require.InDelta(t, float64(system)*float64(user), calculateScaleWithPolicy(desktop.ScaleExact, user, system, 1), 0.00001)
		}
	}
	for _, bad := range []float32{0, -2, float32(math.NaN()), float32(math.Inf(1))} {
		require.True(t, validScale(calculateScaleWithPolicy(desktop.ScaleExact, bad, bad, bad)))
	}
	require.Equal(t, float32(1.3), calculateScale(1, 1.25, 1))
	require.Equal(t, float32(0.1), calculateScale(0.001, 1, 1))
	require.True(t, validScale(calculateScale(1, math.MaxFloat32, 1)))
}

func TestNativeDisplayResolutionMatrix(t *testing.T) {
	if os.Getenv("FYNE_TEST_DISPLAY_MATRIX") != "1" {
		t.Skip("requires native large-framebuffer acceptance")
	}
	for _, resolution := range [][2]int{{2048, 1080}, {2560, 1440}, {3840, 2160}, {5120, 2880}} {
		for _, factor := range []string{"1", "1.25", "2"} {
			t.Run(fmt.Sprintf("%dx%d-scale%s", resolution[0], resolution[1], factor), func(t *testing.T) {
				t.Setenv("FYNE_SCALE", factor)
				w := createWindowWithResizeCallback("Display acceptance")
				defer w.Close()
				w.SetContent(canvas.NewRectangle(color.NRGBA{R: 31, G: 80, B: 130, A: 255}))
				var setupError error
				runOnMain(func() {
					w.window.SetPadded(false)
					setupError = w.window.SetScalePolicy(desktop.ScaleExact)
					w.window.SetFixedSize(true)
					w.window.Show()
					w.viewport.SetPos(0, 0)
					density := w.canvas.Scale() * w.canvas.texScale
					w.window.Resize(granio.NewSize(float32(resolution[0])/density, float32(resolution[1])/density))
				})
				require.NoError(t, setupError)
				require.Eventually(t, func() bool {
					ready := false
					runOnMain(func() {
						width, height := w.viewport.GetFramebufferSize()
						ready = width == resolution[0] && height == resolution[1] && w.frame.ready()
					})
					return ready
				}, 3*time.Second, 10*time.Millisecond)
				var captured image.Image
				var info desktop.WindowScaleInfo
				runOnMain(func() {
					w.window.RunWithContext(func() { d.repaintWindow(w.window) })
					captured = w.window.canvas.Capture()
					info = w.window.ScaleInfo()
				})
				require.Equal(t, resolution[0], captured.Bounds().Dx())
				require.Equal(t, resolution[1], captured.Bounds().Dy())
				r, g, b, a := captured.At(resolution[0]-2, resolution[1]-2).RGBA()
				t.Logf("sample origin=%v centre=%v corner=%v", captured.At(1, 1), captured.At(resolution[0]/2, resolution[1]/2), captured.At(resolution[0]-2, resolution[1]-2))
				require.EqualValues(t, 31*257, r)
				require.EqualValues(t, 80*257, g)
				require.EqualValues(t, 130*257, b)
				require.EqualValues(t, 65535, a)
				t.Logf("framebuffer=%v scale=%+v", captured.Bounds(), info)
			})
		}
	}
}
