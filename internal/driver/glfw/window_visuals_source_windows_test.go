//go:build windows && !no_glfw && !wasm && !test_web_driver && !mobile && !ci

package glfw

import (
	"bufio"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"testing"
	"unsafe"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/stretchr/testify/require"
)

func visualChecker() *canvas.Raster {
	return canvas.NewRasterWithPixels(func(x, y, _, _ int) color.Color {
		shade := uint8(250)
		if (x/16+y/16)%2 != 0 {
			shade = 12
		}
		return color.NRGBA{R: shade, G: shade, B: shade, A: 255}
	})
}

// TestNativeDWMBackgroundSource is an opt-in child process for composition
// tests. It submits exactly one GPU frame per acknowledged stdin command.
func TestNativeDWMBackgroundSource(t *testing.T) {
	geometry := os.Getenv("FYNE_DWM_CHILD")
	if geometry != "full" && geometry != "partial" {
		t.Skip("run through an optical DWM comparison")
	}
	w := newVisualTestWindow("Granio WGL/DXGI background source")
	defer runOnMain(w.Close)
	width := float32(360)
	if geometry == "partial" {
		width = 402
	}
	fill := canvas.NewRectangle(color.NRGBA{R: 220, G: 40, B: 30, A: 255})
	var hwnd uintptr
	var available bool
	runOnMain(func() {
		w.SetPadded(false)
		w.SetContent(fill)
		w.Resize(granio.NewSize(width, 290))
		w.Show()
		if w.viewport != nil {
			w.viewport.SetAttrib(glfw.Floating, glfw.True)
			hwnd = uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
			available = w.presentation.dxgi != nil
		}
	})
	require.True(t, available, "the comparison requires a hardware WGL/DXGI source")
	fmt.Printf("HWND %d\n", hwnd)
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		command := input.Text()
		if command == "q" {
			return
		}
		checker := command == "k"
		var value uint64
		if !checker {
			var err error
			value, err = strconv.ParseUint(command, 16, 32)
			require.NoError(t, err)
		}
		var submitted bool
		runOnMain(func() {
			if checker {
				w.SetContent(visualChecker())
			} else {
				fill.FillColor = color.NRGBA{R: uint8(value), G: uint8(value >> 8), B: uint8(value >> 16), A: 255}
				if w.canvas.Content() != fill {
					w.SetContent(fill)
				}
			}
			w.canvas.CheckDirtyAndClear()
			p := w.presentation.dxgi
			if p != nil {
				before := p.frames
				w.RunWithContext(func() { d.repaintWindow(w) })
				submitted = w.presentation.dxgi == p && p.frames == before+1
			}
		})
		require.True(t, submitted, "expected one successful source Present")
		fmt.Println("FRAME 00000000")
	}
	require.NoError(t, input.Err())
}
