package software_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/driver/desktop"
	"github.com/Uzer007-admin/granio/v3/internal/painter/software"
	"github.com/Uzer007-admin/granio/v3/test"
	"github.com/stretchr/testify/require"
)

type scaledFrameCanvas struct {
	granio.Canvas
	factor float32
}

func (c scaledFrameCanvas) Scale() float32 { return c.factor }

func TestFramePainterClippingScaleAndCaptureOwnership(t *testing.T) {
	base := test.NewCanvas()
	base.Resize(granio.NewSize(40, 24))
	c := scaledFrameCanvas{base, 1.5}
	p := software.NewFramePainter(c)
	p.SetOutputSize(60, 36)
	p.Clear()
	background := p.Frame().NRGBAAt(0, 0)
	r := canvas.NewRectangle(color.NRGBA{R: 255, A: 255})
	r.Resize(granio.NewSize(20, 12))
	p.StartClipping(granio.NewPos(8, 6), granio.NewSize(6, 4))
	p.Paint(r, granio.NewPos(4, 2), c.Size(), nil)
	require.Equal(t, color.NRGBA{R: 255, A: 255}, p.Frame().NRGBAAt(13, 10))
	require.Equal(t, background, p.Frame().NRGBAAt(10, 10))
	require.Equal(t, background, p.Frame().NRGBAAt(22, 10))
	p.StopClipping()
	p.Paint(r, granio.NewPos(20, 12), c.Size(), nil)
	require.Equal(t, color.NRGBA{R: 255, A: 255}, p.Frame().NRGBAAt(32, 20))
	captured := p.Capture(c).(*image.NRGBA)
	frame := p.Frame()
	p.SetOutputSize(60, 36)
	require.Same(t, frame, p.Frame(), "unchanged dimensions must reuse the frame")
	p.Clear()
	require.Equal(t, color.NRGBA{R: 255, A: 255}, captured.NRGBAAt(13, 10), "capture must own its pixels")
	require.Equal(t, background, p.Frame().NRGBAAt(13, 10))
}

func TestFramePainterZeroSizeAndRestore(t *testing.T) {
	c := test.NewCanvas()
	c.Resize(granio.NewSize(20, 12))
	p := software.NewFramePainter(c)
	p.SetOutputSize(0, 12)
	_, err := p.CaptureCanvas(c, func() {})
	require.ErrorIs(t, err, desktop.ErrCaptureUnavailable)
	p.SetOutputSize(int(^uint(0)>>1), 12)
	require.Nil(t, p.Frame(), "invalid dimensions must not allocate")
	p.SetOutputSize(30, 18)
	img, err := p.CaptureCanvas(c, p.Clear)
	require.NoError(t, err)
	require.Equal(t, image.Rect(0, 0, 30, 18), img.Bounds())
	_, _, _, alpha := img.At(0, 0).RGBA()
	require.Equal(t, uint32(65535), alpha, "software window surface is opaque")
	p.SetOutputSize(45, 27)
	require.Equal(t, image.Rect(0, 0, 45, 27), p.Frame().Bounds())
}

func TestFramePainterRoundedShapeKeepsPositionInsideClip(t *testing.T) {
	c := test.NewCanvas()
	c.Resize(granio.NewSize(80, 60))
	p := software.NewFramePainter(c)
	p.SetOutputSize(80, 60)
	p.Clear()
	bg := p.Frame().NRGBAAt(0, 0)
	rect := canvas.NewRectangle(color.NRGBA{R: 242, G: 48, B: 80, A: 255})
	rect.CornerRadius = 17
	rect.Resize(granio.NewSize(54, 42))
	position := granio.NewPos(10, 8)
	p.Paint(rect, position, c.Size(), nil)
	unclipped := p.Capture(c).(*image.NRGBA)
	p.Clear()
	clip := image.Rect(25, 18, 60, 43)
	p.StartClipping(granio.NewPos(25, 18), granio.NewSize(35, 25))
	p.Paint(rect, position, c.Size(), nil)
	for y := 0; y < 60; y++ {
		for x := 0; x < 80; x++ {
			want := bg
			if image.Pt(x, y).In(clip) {
				want = unclipped.NRGBAAt(x, y)
			}
			require.Equal(t, want, p.Frame().NRGBAAt(x, y), "pixel %d,%d", x, y)
		}
	}
}

func TestFramePainterTranslucentMaterialFallback(t *testing.T) {
	c := test.NewCanvas()
	c.Resize(granio.NewSize(80, 60))
	p := software.NewFramePainter(c)
	p.SetOutputSize(80, 60)
	p.Clear()
	base := canvas.NewRectangle(color.Black)
	base.Resize(c.Size())
	p.Paint(base, granio.Position{}, c.Size(), nil)
	source := canvas.NewRectangle(color.White)
	source.Resize(granio.NewSize(20, 40))
	p.Paint(source, granio.NewPos(10, 10), c.Size(), nil)
	m := canvas.NewMaterial(color.NRGBA{R: 40, G: 40, B: 40, A: 230})
	m.TranslucentFallback = true
	m.Resize(granio.NewSize(60, 46))
	position := granio.NewPos(10, 7)
	p.Paint(m, position, c.Size(), nil)
	require.Greater(t, p.Frame().NRGBAAt(20, 28).R, p.Frame().NRGBAAt(50, 28).R, "tint must let the content behind it show through")
	m.Enabled = false
	p.Paint(m, position, c.Size(), nil)
	require.Equal(t, uint8(40), p.Frame().NRGBAAt(20, 28).R)
	require.Equal(t, uint8(40), p.Frame().NRGBAAt(50, 28).R)
}

func TestFramePainterMaterialOpticsAndMixedCorners(t *testing.T) {
	c := test.NewCanvas()
	c.Resize(granio.NewSize(160, 96))
	p := software.NewFramePainter(c)
	p.SetOutputSize(160, 96)
	base := canvas.NewRectangle(color.Black)
	base.Resize(c.Size())
	white := canvas.NewRectangle(color.White)
	white.Resize(granio.NewSize(80, 96))
	m := canvas.NewMaterial(color.NRGBA{G: 120, A: 255})
	m.Tint = color.NRGBA{G: 120, A: 255}
	m.TrackBackdrop, m.SoftwareBlur, m.TranslucentFallback = true, true, true
	m.Resize(granio.NewSize(144, 80))
	m.SetCornerRadii(&granio.CornerRadii{TopLeft: 20, TopRight: 0, BottomRight: 26, BottomLeft: 0})
	draw := func(transparency, radius float32, left float32) {
		require.NoError(t, m.ApplyOptics(transparency, radius))
		p.Clear()
		p.Paint(base, granio.Position{}, c.Size(), nil)
		p.Paint(white, granio.NewPos(left, 0), c.Size(), nil)
		p.Paint(m, granio.NewPos(8, 8), c.Size(), nil)
	}
	draw(0, 50, 0)
	require.Equal(t, color.NRGBA{R: 255, G: 255, B: 255, A: 255}, p.Frame().NRGBAAt(8, 8))
	require.Equal(t, color.NRGBA{G: 120, A: 255}, p.Frame().NRGBAAt(151, 8))
	require.Equal(t, color.NRGBA{A: 255}, p.Frame().NRGBAAt(151, 87))
	require.Equal(t, color.NRGBA{G: 120, A: 255}, p.Frame().NRGBAAt(8, 87))
	draw(1, 0, 0)
	require.Equal(t, uint8(255), p.Frame().NRGBAAt(79, 48).R)
	require.Zero(t, p.Frame().NRGBAAt(80, 48).R)
	width := func() int {
		n := 0
		for x := 12; x < 145; x++ {
			if r := p.Frame().NRGBAAt(x, 48).R; r > 16 && r < 239 {
				n++
			}
		}
		return n
	}
	draw(1, 8, 0)
	short := width()
	require.Positive(t, short)
	draw(1, 45, 0)
	require.Equal(t, canvas.MaterialSoftwareBlur, m.Effect())
	require.Greater(t, width(), short)
	draw(.5, 45, 0)
	require.InDelta(t, 60, p.Frame().NRGBAAt(142, 48).G, 1)
	require.Zero(t, p.Frame().NRGBAAt(142, 48).R)
	m.InvalidateBackdrop()
	draw(1, 45, 66)
	require.Greater(t, p.Frame().NRGBAAt(123, 48).R, uint8(230))
}

func TestFramePainterFramebufferScale(t *testing.T) {
	c := test.NewCanvas()
	c.Resize(granio.NewSize(20, 12))
	p := software.NewFramePainter(c)
	p.SetFrameBufferScale(2)
	p.SetOutputSize(40, 24)
	p.Clear()
	background := p.Frame().NRGBAAt(0, 0)
	r := canvas.NewRectangle(color.NRGBA{G: 255, A: 255})
	r.Resize(granio.NewSize(4, 4))
	p.Paint(r, granio.NewPos(6, 4), c.Size(), nil)
	require.Equal(t, background, p.Frame().NRGBAAt(7, 5))
	require.Equal(t, color.NRGBA{G: 255, A: 255}, p.Frame().NRGBAAt(14, 10))
}

func TestFramePainterTextCacheTracksChanges(t *testing.T) {
	c := test.NewCanvas()
	size := granio.NewSize(160, 44)
	c.Resize(size)
	p := software.NewFramePainter(c)
	label := canvas.NewText("Нажать", color.NRGBA{R: 255, A: 255})
	label.Resize(size)
	label.Alignment = granio.TextAlignCenter

	draw := func(p *software.FramePainter, scale int, x float32) []byte {
		p.SetFrameBufferScale(float32(scale))
		p.SetOutputSize(160*scale, 44*scale)
		p.Clear()
		p.StartClipping(granio.NewPos(20, 0), granio.NewSize(110, 44))
		p.Paint(label, granio.NewPos(x, 0), size, nil)
		return p.Capture(c).(*image.NRGBA).Pix
	}
	original := draw(p, 1, 0)
	require.Equal(t, original, draw(p, 1, 0))
	require.Equal(t, draw(software.NewFramePainter(c), 1, 5), draw(p, 1, 5))

	label.Text = "Готово"
	label.Color = color.NRGBA{G: 255, A: 160}
	drawn := draw(p, 1, 0)
	require.NotEqual(t, original, drawn)
	require.Equal(t, draw(software.NewFramePainter(c), 1, 0), drawn)
	require.Equal(t, draw(software.NewFramePainter(c), 2, 0), draw(p, 2, 0))
}
