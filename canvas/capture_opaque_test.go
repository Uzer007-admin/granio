package canvas_test

import (
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/internal/painter/software"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/stretchr/testify/require"
)

func TestOpaqueCaptureCompositesAlphaWithoutChangingCanvas(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	c := test.NewTransparentCanvasWithPainter(software.NewPainter())
	c.SetPadded(false)
	c.SetContent(canvas.NewRectangle(color.NRGBA{R: 255, A: 128}))
	c.Resize(granio.NewSize(20, 20))
	before := c.Capture().At(10, 10)
	opaque := canvas.CaptureOpaque(c, color.NRGBA{B: 255, A: 20})
	pixel := opaque.NRGBAAt(10, 10)
	require.EqualValues(t, 255, pixel.A)
	require.InDelta(t, 128, pixel.R, 1)
	require.InDelta(t, 127, pixel.B, 1)
	require.Equal(t, before, c.Capture().At(10, 10))
	require.Nil(t, canvas.CaptureOpaque(nil, color.Black))
	m, err := theme.NewModern(theme.ModernOptions{HighContrast: true})
	require.NoError(t, err)
	material := canvas.NewThemedMaterial(m, granio.SurfacePopup, theme.VariantLight)
	require.False(t, material.Enabled)
	_, _, _, alpha := material.OpaqueFallback().RGBA()
	require.EqualValues(t, 65535, alpha)
}
