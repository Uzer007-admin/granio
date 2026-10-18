package canvas_test

import (
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/software"
	"github.com/Uzer007-admin/granio/v2/test"

	"github.com/stretchr/testify/assert"
)

func TestEllipse_MinSize(t *testing.T) {
	ellipse := canvas.NewEllipse(color.Black)
	minSize := ellipse.MinSize()

	assert.Positive(t, minSize.Width)
	assert.Positive(t, minSize.Height)
}

func TestEllipse_FillColor(t *testing.T) {
	c := color.White
	ellipse := canvas.NewEllipse(c)

	assert.Equal(t, c, ellipse.FillColor)
}

func TestEllipse_Resize(t *testing.T) {
	targetWidth := float32(50)
	targetHeight := float32(80)
	ellipse := canvas.NewEllipse(color.White)
	start := ellipse.Size()
	assert.True(t, start.Height == 0)
	assert.True(t, start.Width == 0)

	ellipse.Resize(granio.NewSize(targetWidth, targetHeight))
	target := ellipse.Size()
	assert.True(t, target.Height == targetHeight)
	assert.True(t, target.Width == targetWidth)
}

func TestEllipse_Move(t *testing.T) {
	ellipse := canvas.NewEllipse(color.White)
	ellipse.Resize(granio.NewSize(80, 50))

	start := granio.Position{X: 0, Y: 0}
	assert.Equal(t, ellipse.Position(), start)

	target := granio.Position{X: 10, Y: 75}
	ellipse.Move(target)
	assert.Equal(t, ellipse.Position(), target)
}

func TestEllipse_Stroke(t *testing.T) {
	ellipse := &canvas.Ellipse{
		FillColor:   color.NRGBA{R: 255, G: 200, B: 0, A: 180},
		StrokeColor: color.Black,
		StrokeWidth: 2,
	}

	ellipse.Resize(granio.NewSize(50, 90))
	test.AssertObjectRendersToMarkup(t, "ellipse.xml", ellipse)

	c := software.NewCanvas()
	c.SetContent(ellipse)
	c.Resize(granio.NewSize(60, 60))

	ellipse.Resize(granio.NewSize(50, 30))
	test.AssertRendersToImage(t, "ellipse_stroke.png", c)

	ellipse.Resize(granio.NewSize(30, 50))
	ellipse.StrokeWidth = 0
	test.AssertRendersToImage(t, "ellipse.png", c)
}
