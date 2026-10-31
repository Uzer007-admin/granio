package software_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal/painter/software"
	"github.com/Uzer007-admin/granio/v2/test"
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
