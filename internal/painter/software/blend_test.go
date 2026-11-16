package software

import (
	"image"
	"image/color"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/image/draw"
)

func TestBlendSolidRectMatchesDraw(t *testing.T) {
	bounds := image.Rect(-5, 14, 150, 105)
	sub := image.Rect(2, 15, 130, 90)
	random := rand.New(rand.NewPCG(43, 97))
	for _, fill := range []color.Color{
		color.Transparent,
		color.NRGBA{R: 255, G: 22, B: 81, A: 1},
		color.NRGBA{R: 255, G: 22, B: 81, A: 137},
		color.NRGBA{R: 255, G: 22, B: 81, A: 254},
		color.NRGBA64{R: 0xfffe, G: 0x1234, B: 0x4321, A: 0x93ea},
		color.RGBA{R: 180, G: 77, B: 123, A: 190},
	} {
		want, got := image.NewNRGBA(bounds), image.NewNRGBA(bounds)
		for i := 0; i < len(want.Pix); i += 4 {
			want.Pix[i], want.Pix[i+1], want.Pix[i+2] = uint8(random.Uint32()), uint8(random.Uint32()), uint8(random.Uint32())
			want.Pix[i+3] = uint8(random.Uint32())
			switch i / 4 % 4 {
			case 0:
				want.Pix[i+3] = 0
			case 1:
				want.Pix[i+3] = 255
			}
		}
		for _, clip := range []image.Rectangle{image.Rect(-2, 19, 41, 33), image.Rect(-2, 19, 145, 83)} {
			copy(got.Pix, want.Pix)
			draw.Draw(want.SubImage(sub).(*image.NRGBA), clip, image.NewUniform(fill), image.Point{}, draw.Over)
			blendSolidRect(got.SubImage(sub).(*image.NRGBA), clip, fill)
			require.Equal(t, want.Pix, got.Pix, "source %v, clip %v", fill, clip)
		}
	}
}
