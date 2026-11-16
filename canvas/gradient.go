package canvas

import (
	"image"
	"image/color"
	"math"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/internal/painter/geom"
)

// LinearGradient defines a Gradient travelling straight at a given angle.
// The only supported values for the angle are `0.0` (vertical) and `90.0` (horizontal), currently.
type LinearGradient struct {
	baseObject

	StartColor color.Color // The beginning color of the gradient
	EndColor   color.Color // The end color of the gradient
	Angle      float64     // The angle of the gradient (0/180 for vertical; 90/270 for horizontal)
}

// Generate calculates an image of the gradient with the specified width and height.
func (g *LinearGradient) Generate(iw, ih int) image.Image {
	w, h := float64(iw), float64(ih)
	var generator func(x, y float64) float64
	repeatRows := false
	switch g.Angle {
	case geom.AngleQuarter, -geom.AngleThreeQuarter: // horizontal flipped
		repeatRows = true
		generator = func(x, _ float64) float64 {
			return (w - x) / w
		}
	case geom.AngleThreeQuarter, -geom.AngleQuarter: // horizontal
		repeatRows = true
		generator = func(x, _ float64) float64 {
			return x / w
		}
	case geom.AngleEighth, -geom.AngleSevenEighth: // diagonal negative flipped
		generator = func(x, y float64) float64 {
			return math.Abs((w - x + y) / (w + h)) // ((w+h)-(x+h-y)) / (w+h)
		}
	case geom.AngleFiveEighth, -geom.AngleThreeEighth: // diagonal negative
		generator = func(x, y float64) float64 {
			return math.Abs((x + h - y) / (w + h))
		}
	case geom.AngleThreeEighth, -geom.AngleFiveEighth: // diagonal positive flipped
		generator = func(x, y float64) float64 {
			return math.Abs((w + h - (x + y)) / (w + h))
		}
	case geom.AngleSevenEighth, -geom.AngleEighth: // diagonal positive
		generator = func(x, y float64) float64 {
			return math.Abs((x + y) / (w + h))
		}
	case geom.AngleHalf, -geom.AngleHalf: // vertical flipped
		generator = func(_, y float64) float64 {
			return (h - y) / h
		}
	default: // vertical
		generator = func(_, y float64) float64 {
			return y / h
		}
	}
	return computeGradient(generator, iw, ih, g.StartColor, g.EndColor, repeatRows)
}

// Hide will set this gradient to not be visible
func (g *LinearGradient) Hide() {
	g.baseObject.Hide()

	repaint(g)
}

// Move the gradient to a new position, relative to its parent / canvas
func (g *LinearGradient) Move(pos granio.Position) {
	if g.Position() == pos {
		return
	}

	g.baseObject.Move(pos)

	repaint(g)
}

// Resize resizes the gradient to a new size.
func (g *LinearGradient) Resize(size granio.Size) {
	if size == g.Size() {
		return
	}
	g.baseObject.Resize(size)

	// refresh needed to invalidate cached textures
	g.Refresh()
}

// Refresh causes this gradient to be redrawn with its configured state.
func (g *LinearGradient) Refresh() {
	Refresh(g)
}

// RadialGradient defines a Gradient travelling radially from a center point outward.
type RadialGradient struct {
	baseObject

	StartColor color.Color // The beginning color of the gradient
	EndColor   color.Color // The end color of the gradient
	// The offset of the center for generation of the gradient.
	// This is not a DP measure but relates to the width/height.
	// A value of 0.5 would move the center by the half width/height.
	CenterOffsetX, CenterOffsetY float64
}

// Generate calculates an image of the gradient with the specified width and height.
func (g *RadialGradient) Generate(iw, ih int) image.Image {
	w, h := float64(iw), float64(ih)
	// define center plus offset
	centerX := w/2 + w*g.CenterOffsetX
	centerY := h/2 + h*g.CenterOffsetY

	// handle negative offsets
	var a, b float64
	if g.CenterOffsetX < 0 {
		a = w - centerX
	} else {
		a = centerX
	}
	if g.CenterOffsetY < 0 {
		b = h - centerY
	} else {
		b = centerY
	}

	generator := func(x, y float64) float64 {
		// calculate distance from center for gradient multiplier
		dx, dy := centerX-x, centerY-y
		da := math.Sqrt(dx*dx + dy*dy*a*a/b/b)
		if da > a {
			return 1
		}
		return da / a
	}
	return computeGradient(generator, iw, ih, g.StartColor, g.EndColor, false)
}

// Hide will set this gradient to not be visible
func (g *RadialGradient) Hide() {
	g.baseObject.Hide()

	repaint(g)
}

// Move the gradient to a new position, relative to its parent / canvas
func (g *RadialGradient) Move(pos granio.Position) {
	g.baseObject.Move(pos)

	repaint(g)
}

// Resize resizes the gradient to a new size.
func (g *RadialGradient) Resize(size granio.Size) {
	if size == g.Size() {
		return
	}
	g.baseObject.Resize(size)

	// refresh needed to invalidate cached textures
	g.Refresh()
}

// Refresh causes this gradient to be redrawn with its configured state.
func (g *RadialGradient) Refresh() {
	Refresh(g)
}

func computeGradient(generator func(x, y float64) float64, w, h int, startColor, endColor color.Color, repeatRows bool) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))

	if startColor == nil && endColor == nil {
		return img
	} else if startColor == nil {
		startColor = color.Transparent
	} else if endColor == nil {
		endColor = color.Transparent
	}

	r0, g0, b0, a0 := startColor.RGBA()
	r1, g1, b1, a1 := endColor.RGBA()
	dr := float64(r1) - float64(r0)
	dg := float64(g1) - float64(g0)
	db := float64(b1) - float64(b0)
	da := float64(a1) - float64(a0)
	for y := 0; y < h; y++ {
		if repeatRows && y != 0 {
			offset := y * img.Stride
			copy(img.Pix[offset:offset+img.Stride], img.Pix[:img.Stride])
			continue
		}
		for x := 0; x < w; x++ {
			distance := generator(float64(x)+0.5, float64(y)+0.5)
			a := uint16(float64(a0) + distance*da)
			if a == 0 {
				continue
			}
			img.SetRGBA64(x, y, color.RGBA64{
				R: uint16(float64(r0) + distance*dr),
				G: uint16(float64(g0) + distance*dg),
				B: uint16(float64(b0) + distance*db),
				A: a,
			})
		}
	}
	return img
}

// NewHorizontalGradient creates a new horizontally travelling linear gradient.
// The start color will be at the left of the gradient and the end color will be at the right.
func NewHorizontalGradient(start, end color.Color) *LinearGradient {
	g := &LinearGradient{StartColor: start, EndColor: end}
	g.Angle = geom.AngleThreeQuarter
	return g
}

// NewLinearGradient creates a linear gradient at the specified angle.
// The angle parameter is the degree angle along which the gradient is calculated.
// A NewHorizontalGradient uses 270 degrees and NewVerticalGradient is 0 degrees.
func NewLinearGradient(start, end color.Color, angle float64) *LinearGradient {
	g := &LinearGradient{StartColor: start, EndColor: end}
	g.Angle = angle
	return g
}

// NewRadialGradient creates a new radial gradient.
func NewRadialGradient(start, end color.Color) *RadialGradient {
	return &RadialGradient{StartColor: start, EndColor: end}
}

// NewVerticalGradient creates a new vertically travelling linear gradient.
// The start color will be at the top of the gradient and the end color will be at the bottom.
func NewVerticalGradient(start color.Color, end color.Color) *LinearGradient {
	return &LinearGradient{StartColor: start, EndColor: end}
}
