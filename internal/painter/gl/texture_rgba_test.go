package gl

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTextureRGBAOpaqueAndAlpha(t *testing.T) {
	original := image.NewNRGBA(image.Rect(2, 3, 5, 6))
	for y := 3; y < 6; y++ {
		for x := 2; x < 5; x++ {
			original.SetNRGBA(x, y, color.NRGBA{R: 200, G: uint8(x), B: uint8(y), A: 255})
		}
	}
	sub := original.SubImage(image.Rect(3, 4, 5, 6)).(*image.NRGBA)
	opaque := textureRGBA(sub)
	require.Same(t, &sub.Pix[0], &opaque.Pix[0])
	require.Equal(t, []byte{200, 3, 4, 255, 200, 4, 4, 255, 200, 3, 5, 255, 200, 4, 5, 255},
		packedRGBAPixels(&opaque))

	sub.SetNRGBA(3, 4, color.NRGBA{R: 200, G: 100, B: 240, A: 128})
	converted := textureRGBA(sub)
	require.Equal(t, image.Rect(0, 0, 2, 2), converted.Rect)
	want := color.RGBAModel.Convert(sub.NRGBAAt(3, 4)).(color.RGBA)
	require.Equal(t, want, converted.RGBAAt(0, 0))
	require.Equal(t, color.RGBA{R: 200, G: 4, B: 5, A: 255}, converted.RGBAAt(1, 1))
}
