package theme

import (
	"image/color"
	"math"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/stretchr/testify/require"
)

func TestModernSystemAppearanceOverridesAndRestore(t *testing.T) {
	snapshot := granio.SystemAppearance{Available: true,
		TextScale:    granio.SystemPreference[float32]{Value: 2.25, Available: true},
		Accent:       granio.SystemPreference[color.NRGBA]{Value: color.NRGBA{R: 200, G: 100, A: 255}, Available: true},
		HighContrast: granio.SystemPreference[bool]{Available: true},
	}
	original, err := NewModern(ModernOptions{FollowSystem: true})
	require.NoError(t, err)
	adapted := original.(granio.SystemAppearanceTheme).WithSystemAppearance(snapshot)
	require.EqualValues(t, 31.5, adapted.Size(SizeNameText))
	require.Equal(t, original.Size(SizeNameInlineIcon), adapted.Size(SizeNameInlineIcon))
	require.EqualValues(t, 14, original.Size(SizeNameText), "shared theme mutated")
	require.Equal(t, snapshot.Accent.Value, adapted.Color(ColorNamePrimary, VariantLight))
	explicit, err := NewModern(ModernOptions{FollowSystem: true, TextScale: 1.5, Accent: color.NRGBA{B: 200, A: 255}})
	require.NoError(t, err)
	explicit = explicit.(granio.SystemAppearanceTheme).WithSystemAppearance(snapshot)
	require.EqualValues(t, 21, explicit.Size(SizeNameText))
	require.Equal(t, color.NRGBA{B: 200, A: 255}, explicit.Color(ColorNamePrimary, VariantLight))
	legacy, _ := NewModern(ModernOptions{})
	legacyAdapted := legacy.(granio.SystemAppearanceTheme).WithSystemAppearance(snapshot)
	require.Equal(t, legacy.Color(ColorNamePrimary, VariantLight), legacyAdapted.Color(ColorNamePrimary, VariantLight))
	require.Equal(t, legacy.Size(SizeNameText), legacyAdapted.Size(SizeNameText))
	for _, scale := range []float32{0, -1, 5, float32(math.NaN()), float32(math.Inf(1))} {
		snapshot.TextScale.Value = scale
		next := original.(granio.SystemAppearanceTheme).WithSystemAppearance(snapshot)
		require.EqualValues(t, 14, next.Size(SizeNameText))
	}
	restored := adapted.(granio.SystemAppearanceTheme).WithSystemAppearance(granio.SystemAppearance{})
	require.Equal(t, original.Color(ColorNamePrimary, VariantLight), restored.Color(ColorNamePrimary, VariantLight))
	require.Equal(t, original.Size(SizeNameText), restored.Size(SizeNameText))
}

func TestModernSystemHighContrastPalette(t *testing.T) {
	p := granio.SystemColors{Window: color.NRGBA{G: 24, A: 255}, WindowText: color.NRGBA{R: 240, G: 240, A: 255},
		ButtonFace: color.NRGBA{B: 30, A: 255}, ButtonText: color.NRGBA{R: 240, G: 240, A: 255},
		Highlight: color.NRGBA{R: 50, G: 60, B: 70, A: 255}, HighlightText: color.NRGBA{R: 255, G: 255, A: 255},
		GrayText: color.NRGBA{R: 180, G: 180, B: 180, A: 255}, Hotlight: color.NRGBA{G: 255, A: 255}}
	original, _ := NewModern(ModernOptions{FollowSystem: true, Accent: color.NRGBA{R: 255, A: 255}})
	snapshot := granio.SystemAppearance{Available: true, HighContrast: granio.SystemPreference[bool]{Value: true, Available: true}, Colors: granio.SystemPreference[granio.SystemColors]{Value: p, Available: true}}
	adapted := original.(granio.SystemAppearanceTheme).WithSystemAppearance(snapshot)
	require.True(t, adapted.(interface{ HighContrast() bool }).HighContrast())
	for role, want := range map[granio.ThemeColorName]color.Color{ColorNameBackground: p.Window, ColorNameForeground: p.WindowText,
		ColorNameButton: p.ButtonFace, ColorNameInputBackground: p.Window, ColorNamePrimary: p.Highlight, ColorNameForegroundOnPrimary: p.HighlightText,
		ColorNameSelection: p.Highlight, ColorNameDisabled: p.GrayText, ColorNameHyperlink: p.Hotlight, ColorNameShadow: color.Transparent,
		ColorNameHover: color.Transparent, ColorNamePressed: color.Transparent} {
		for _, variant := range []granio.ThemeVariant{VariantLight, VariantDark} {
			require.Equal(t, want, adapted.Color(role, variant), role)
		}
	}
	snapshot.HighContrast.Value = false
	restored := adapted.(granio.SystemAppearanceTheme).WithSystemAppearance(snapshot)
	require.False(t, restored.(interface{ HighContrast() bool }).HighContrast())
	require.Equal(t, original.Color(ColorNamePrimary, VariantLight), restored.Color(ColorNamePrimary, VariantLight))
}
