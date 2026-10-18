//go:build graphics_probe

// Run under Xvfb: go run -tags gles,egl ./scripts/fyne-gles-probe/main.go
package main

import (
	"fmt"
	"image/color"
	"os"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/canvas"
)

func main() {
	a := app.NewWithID("native-stack-gles-probe")
	w := a.NewWindow("GLES2 consumer")
	w.SetMaster()
	w.SetContent(canvas.NewRectangle(color.NRGBA{G: 255, A: 255}))
	w.Resize(granio.NewSize(64, 64))
	go func() {
		time.Sleep(time.Second)
		granio.Do(func() {
			frame := w.Canvas().Capture()
			bounds := frame.Bounds()
			r, g, b, _ := frame.At(bounds.Min.X+bounds.Dx()/2, bounds.Min.Y+bounds.Dy()/2).RGBA()
			fmt.Printf("Fyne GLES capture: bounds=%v center=[%d,%d,%d]\n", bounds, r, g, b)
			if g < 0xff00 || r > 0xff || b > 0xff {
				fmt.Fprintf(os.Stderr, "unexpected green framebuffer: %d %d %d\n", r, g, b)
				os.Exit(1)
			}
			// This one-shot render probe deliberately exits after readback.
			// Native window lifecycle is exercised by separate integration tests.
			os.Exit(0)
		})
	}()
	w.ShowAndRun()
}
