package theme

import (
	"github.com/stretchr/testify/require"
	"image/color"
	"math"
	"testing"
)

func TestModernContrastAndTextScaling(t *testing.T) {
	for _, variant := range []uint{0, 1} {
		for _, high := range []bool{false, true} {
			for _, accent := range []color.Color{color.Black, color.White, color.NRGBA{R: 255, G: 255, A: 255}, color.NRGBA{B: 255, A: 255}} {
				m, err := NewModern(ModernOptions{TextScale: 2.25, Accent: accent, HighContrast: high})
				require.NoError(t, err)
				th := m.(*modernTheme)
				v := VariantDark
				if variant == 1 {
					v = VariantLight
				}
				fg, bg := luminance(th.Color(ColorNameForeground, v)), luminance(th.Color(ColorNameBackground, v))
				ratio := (math.Max(fg, bg) + 0.05) / (math.Min(fg, bg) + 0.05)
				require.GreaterOrEqual(t, ratio, 4.5)
				fg, bg = luminance(th.Color(ColorNameForegroundOnPrimary, v)), luminance(th.Color(ColorNamePrimary, v))
				require.GreaterOrEqual(t, (math.Max(fg, bg)+0.05)/(math.Min(fg, bg)+0.05), 4.5)
				require.EqualValues(t, 31.5, th.Size(SizeNameText))
				require.Equal(t, DefaultTheme().Size(SizeNameInlineIcon), th.Size(SizeNameInlineIcon))
			}
		}
	}
	_, err := NewModern(ModernOptions{TextScale: float32(math.NaN())})
	require.Error(t, err)
}
