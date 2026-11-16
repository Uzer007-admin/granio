package painter_test

import (
	"image"
	"image/color"
	"testing"

	"golang.org/x/image/draw"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/internal/painter"
	"github.com/Uzer007-admin/granio/v3/internal/painter/software"
	"github.com/Uzer007-admin/granio/v3/test"
	"github.com/stretchr/testify/assert"
)

func TestPaintImage_MinSize(t *testing.T) {
	img := canvas.NewImageFromFile("testdata/svg-stroke-default.png")
	img.FillMode = canvas.ImageFillOriginal
	c := test.NewCanvasWithPainter(software.NewPainter())
	c.SetScale(1.0)
	c.SetContent(img)

	// check fallback min
	assert.Equal(t, float32(480), img.MinSize().Width)
	assert.Equal(t, float32(240), img.MinSize().Height)

	// render it with original will set min
	painter.PaintImage(img, c, 100, 100)
	assert.Equal(t, float32(480), img.MinSize().Width)
	assert.Equal(t, float32(240), img.MinSize().Height)
}

func TestPaintImageWithBadSVGElement(t *testing.T) {
	img := canvas.NewImageFromFile("testdata/stroke-bad-element-data.svg")
	c := test.NewCanvasWithPainter(software.NewPainter())
	c.SetContent(img)
	c.Resize(granio.NewSize(480, 240))
	img.Refresh()

	test.AssertImageMatches(t, "svg-stroke-default.png", c.Capture())
}

func TestPaintImage_SVG(t *testing.T) {
	test.NewTempApp(t)

	for name, tt := range map[string]struct {
		width     float32
		height    float32
		fillMode  canvas.ImageFill
		wantImage string
	}{
		"default": {
			width:  480,
			height: 240,
		},
		"stretchx": {
			width:    640,
			height:   240,
			fillMode: canvas.ImageFillStretch,
		},
		"stretchy": {
			width:    480,
			height:   480,
			fillMode: canvas.ImageFillStretch,
		},
		"containx": {
			width:    640,
			height:   240,
			fillMode: canvas.ImageFillContain,
		},
		"containy": {
			width:    480,
			height:   480,
			fillMode: canvas.ImageFillContain,
		},
	} {
		t.Run(name, func(t *testing.T) {
			img := canvas.NewImageFromFile("testdata/stroke.svg")
			c := test.NewCanvasWithPainter(software.NewPainter())
			c.SetContent(img)
			c.Resize(granio.NewSize(tt.width, tt.height))
			img.Refresh()
			img.FillMode = tt.fillMode

			test.AssertImageMatches(t, "svg-stroke-"+name+".png", c.Capture())
		})
	}
}

func TestPaintImage_Direct(t *testing.T) {
	img := canvas.NewImageFromFile("testdata/stroke.svg")
	pix := painter.PaintImage(img, nil, 480, 240)

	test.AssertImageMatches(t, "svg-stroke-direct.png", pix)
}

func TestScaleImageSmoothPreservesSubimagePixels(t *testing.T) {
	src := image.NewNRGBA(image.Rect(-9, -5, 130, 100))
	for y := src.Rect.Min.Y; y < src.Rect.Max.Y; y++ {
		for x := src.Rect.Min.X; x < src.Rect.Max.X; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 23), G: uint8(y * 19), B: 201, A: uint8(x + 3*y)})
		}
	}
	src = src.SubImage(image.Rect(3, 4, 104, 71)).(*image.NRGBA)
	for _, offset := range []image.Point{{X: 7, Y: 9}, {X: -2, Y: 13}} {
		pixels := image.NewNRGBA(image.Rectangle{Min: offset.Sub(image.Pt(3, 3)), Max: offset.Add(image.Pt(58, 40))})
		bounds := image.Rectangle{Min: offset, Max: offset.Add(image.Pt(53, 36))}
		dst := pixels.SubImage(bounds).(*image.NRGBA)
		want := image.NewNRGBA(bounds)
		draw.CatmullRom.Scale(want, bounds, src, src.Bounds(), draw.Over, nil)
		painter.ScaleImageSmooth(dst, src)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			row := dst.PixOffset(bounds.Min.X, y)
			expected := want.PixOffset(bounds.Min.X, y)
			assert.Equal(t, want.Pix[expected:expected+4*bounds.Dx()], dst.Pix[row:row+4*bounds.Dx()])
		}
	}
}
