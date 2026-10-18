package software_test

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/software"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func Test_canvas_Capture(t *testing.T) {
	c := software.NewCanvas()
	c.Resize(granio.NewSquareSize(100))

	img := c.Capture()
	assert.Positive(t, img.Bounds().Size().X)
	assert.Positive(t, img.Bounds().Size().Y)

	r1, g1, b1, a1 := theme.Color(theme.ColorNameBackground).RGBA()
	r2, g2, b2, a2 := img.At(1, 1).RGBA()
	assert.Equal(t, r1, r2)
	assert.Equal(t, g1, g2)
	assert.Equal(t, b1, b2)
	assert.Equal(t, a1, a2)
}

func Test_canvas_InteractiveArea(t *testing.T) {
	c := software.NewCanvas()
	// TODO a custom interactive area that does not break other tests (overlays restrict to the space)
	c.Resize(granio.NewSize(600, 400))
	pos, size := c.InteractiveArea()
	assert.Equal(t, granio.NewPos(0, 0), pos)
	assert.Equal(t, granio.NewSize(600, 400), size)
}

func Test_canvas_PixelCoordinateAtPosition(t *testing.T) {
	c := software.NewCanvas()

	pos := granio.NewPos(4, 4)
	c.SetScale(2.5)
	x, y := c.PixelCoordinateForPosition(pos)
	assert.Equal(t, 10, x)
	assert.Equal(t, 10, y)
}

func Test_canvas_TransparentCapture(t *testing.T) {
	c := software.NewTransparentCanvasWithPainter(nil)
	c.Resize(granio.NewSquareSize(10))

	img := c.Capture()
	assert.Positive(t, img.Bounds().Size().X)
	assert.Positive(t, img.Bounds().Size().Y)

	r1, g1, b1, a1 := color.Transparent.RGBA()
	r2, g2, b2, a2 := img.At(1, 1).RGBA()
	assert.Equal(t, r1, r2)
	assert.Equal(t, g1, g2)
	assert.Equal(t, b1, b2)
	assert.Equal(t, a1, a2)
}

func Test_canvas_Resize(t *testing.T) {
	c := software.NewCanvas()
	assert.Equal(t, float32(100), c.Size().Width) // backwards compatible

	smallSize := granio.NewSize(10, 5)
	c.Resize(smallSize)
	c.SetContent(widget.NewLabel("Bigger"))
	assert.NotEqual(t, float32(100), c.Size().Width) // backwards compatible
	assert.Greater(t, c.Size().Width, smallSize.Width)
	assert.Greater(t, c.Size().Height, smallSize.Height)

	c = software.NewCanvas()
	largeSize := granio.NewSize(100, 75)
	c.Resize(largeSize)
	c.SetContent(widget.NewLabel("Smaller"))
	assert.Equal(t, largeSize, c.Size())
}
