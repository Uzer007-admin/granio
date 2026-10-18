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

func TestBezierCurve_MinSizeAndStrokeColor(t *testing.T) {
	curve := canvas.NewQuadraticBezierCurve(granio.NewPos(0, 0), granio.NewPos(20, -5), granio.NewPos(30, 0), color.Black)
	minSize := curve.MinSize()

	assert.Positive(t, minSize.Width)
	assert.Positive(t, minSize.Height)

	assert.Equal(t, color.Black, curve.StrokeColor)
}

func TestBezierCurve_Linear(t *testing.T) {
	curve := canvas.NewLinearBezierCurve(granio.NewPos(10, 0), granio.NewPos(50, 10), color.Black)
	curve.StrokeWidth = 1
	curve.Resize(granio.NewSize(50, 50))
	test.AssertObjectRendersToMarkup(t, "bezier_curve_linear.xml", curve)

	c := software.NewCanvas()
	c.SetContent(curve)
	c.Resize(granio.NewSize(70, 60))
	test.AssertRendersToImage(t, "bezier_curve_linear.png", c)

	curve.StartPoint = granio.NewPos(10, -5)
	test.AssertRendersToImage(t, "bezier_curve_linear.png", c)
}

func TestBezierCurve_Quadratic(t *testing.T) {
	curve := canvas.NewQuadraticBezierCurve(granio.NewPos(0, 0), granio.NewPos(30, 40), granio.NewPos(50, 10), color.Black)
	curve.StrokeWidth = 2
	curve.Resize(granio.NewSize(50, 50))
	test.AssertObjectRendersToMarkup(t, "bezier_curve_quadratic.xml", curve)

	c := software.NewCanvas()
	c.SetContent(curve)
	c.Resize(granio.NewSize(70, 60))
	test.AssertRendersToImage(t, "bezier_curve_quadratic.png", c)

	curve.StartPoint = granio.NewPos(-10, -10)
	test.AssertRendersToImage(t, "bezier_curve_quadratic.png", c)
}

func TestBezierCurve_Cubic(t *testing.T) {
	curve := canvas.NewCubicBezierCurve(granio.NewPos(0, 60), granio.NewPos(20, 30), granio.NewPos(40, 60), granio.NewPos(50, 10), color.RGBA{R: 0, G: 0, B: 255, A: 255})
	curve.Resize(granio.NewSize(50, 50))
	test.AssertObjectRendersToMarkup(t, "bezier_curve_cubic.xml", curve)

	c := software.NewCanvas()
	c.SetContent(curve)
	c.Resize(granio.NewSize(70, 60))
	test.AssertRendersToImage(t, "bezier_curve_cubic.png", c)

	curve.ControlPoints = []granio.Position{granio.NewPos(20, 30), granio.NewPos(40, 500), granio.NewPos(60, 60)}
	test.AssertRendersToImage(t, "bezier_curve_cubic.png", c)
}
