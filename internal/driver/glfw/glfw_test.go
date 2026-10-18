//go:build !no_glfw && !mobile

package glfw

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Uzer007-admin/granio/v2"
)

func assertCanvasSize(t *testing.T, w *safeWindow, size granio.Size) {
	if runtime.GOOS == "linux" || runtime.GOOS == "windows" {
		// TODO: find the root cause for these problems and solve them without additional repaint
		// fixes issues where the window does not have the correct size
		waitForCanvasSize(t, w, size, runtime.GOOS == "windows")
	}
	assert.Equal(t, size, w.Canvas().Size())
}

func ensureCanvasSize(t *testing.T, w *safeWindow, size granio.Size) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		// TODO: find the root cause for these problems and solve them without additional repaint
		// fixes issues where the window does not have the correct size
		waitForCanvasSize(t, w, size, true)
	}
	require.Equal(t, size, w.Canvas().Size())
}

func repaintWindow(w *safeWindow) {
	runOnMain(func() {
		w.RunWithContext(func() {
			d.repaintWindow(w.window)
		})
	})

	time.Sleep(time.Millisecond * 150) // wait for the frames to be rendered... o
}

func waitForCanvasSize(t *testing.T, w *safeWindow, size granio.Size, resizeIfNecessary bool) {
	attempts := 0
	for w.Canvas().Size() != size {
		attempts++
		require.Less(t, attempts, 100, "canvas did not get correct size in time")

		if resizeIfNecessary && attempts%20 == 0 {
			// sometimes the resize does not seem to reach the actual window at all
			w.Resize(size)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
