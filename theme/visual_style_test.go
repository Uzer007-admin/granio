package theme

import (
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/stretchr/testify/require"
)

type invalidTypography struct{ granio.Theme }

func (invalidTypography) Typography(granio.TextRole) granio.TypographyStyle {
	return granio.TypographyStyle{Size: float32(math.NaN()), LineSpacing: float32(math.Inf(1))}
}

func TestVisualTokensCompatibilityAndResolution(t *testing.T) {
	legacy := DefaultTheme()
	body := Typography(legacy, granio.TextRoleBody, granio.TextStyle{})
	require.Equal(t, legacy.Size(SizeNameText), body.Size)
	require.Equal(t, legacy.Font(granio.TextStyle{}), body.Font)
	invalid := Typography(invalidTypography{legacy}, granio.TextRoleTitle, granio.TextStyle{Bold: true})
	require.EqualValues(t, 28, invalid.Size)
	require.EqualValues(t, 4, invalid.LineSpacing)
	require.True(t, invalid.Style.Bold)
	require.NotNil(t, invalid.Font)
	_, err := NewModern(ModernOptions{TokenVersion: ModernTokenVersion + 1})
	require.Error(t, err)
	m, err := NewModern(ModernOptions{TokenVersion: ModernTokenVersion, TextScale: 2.25})
	require.NoError(t, err)
	roles := map[granio.TextRole]float32{granio.TextRoleBody: 14, granio.TextRoleCaption: 12, granio.TextRoleEmphasis: 16, granio.TextRoleSubtitle: 20, granio.TextRoleTitle: 28, granio.TextRoleCode: 14, granio.TextRole(255): 14}
	for role, size := range roles {
		style := Typography(m, role, granio.TextStyle{Italic: true})
		require.Equal(t, size*2.25, style.Size)
		require.True(t, style.Style.Italic)
		require.NotNil(t, style.Font)
		require.Equal(t, m.Font(style.Style), style.Font)
	}
	require.True(t, Typography(m, granio.TextRoleCode, granio.TextStyle{}).Style.Monospace)
	require.Equal(t, legacy.Size(SizeNameInlineIcon), m.Size(SizeNameInlineIcon))
	require.Greater(t, m.Size(granio.ThemeSizeName("future.token")), float32(0))
	require.NotNil(t, m.Color(granio.ThemeColorName("future.token"), VariantLight))
}

func TestVisualStatesAndMotionRestrictions(t *testing.T) {
	m, err := NewModern(ModernOptions{Density: DensityCompact})
	require.NoError(t, err)
	provider := m.(granio.ComponentTheme)
	for _, variant := range []granio.ThemeVariant{VariantLight, VariantDark} {
		state := granio.VisualReadOnly | granio.VisualError | granio.VisualFocused
		r := provider.ComponentStyle(granio.ComponentInput, state, variant)
		require.Equal(t, m.Color(ColorNameError, variant), r.Border)
		require.Equal(t, m.Color(ColorNameFocus, variant), r.Focus)
		require.Greater(t, r.FocusWidth, float32(0))
		require.EqualValues(t, 32, r.MinimumHeight)
		disabled := provider.ComponentStyle(granio.ComponentInput, state|granio.VisualDisabled|granio.VisualHovered|granio.VisualPressed, variant)
		require.Zero(t, disabled.FocusWidth)
		require.Equal(t, disabled.Background, provider.ComponentStyle(granio.ComponentInput, state|granio.VisualDisabled, variant).Background)
	}
	for role, duration := range map[granio.MotionRole]time.Duration{granio.MotionHover: 100 * time.Millisecond, granio.MotionPressed: 80 * time.Millisecond, granio.MotionSurface: 180 * time.Millisecond, granio.MotionNavigation: 220 * time.Millisecond} {
		require.Equal(t, duration, Motion(m, role).Duration)
	}
	adapted := m.(granio.SystemAppearanceTheme).WithSystemAppearance(granio.SystemAppearance{Available: true, HighContrast: granio.SystemPreference[bool]{Available: true, Value: true}})
	require.True(t, Motion(adapted, granio.MotionHover).Disabled)
	require.True(t, adapted.(granio.MaterialTheme).MaterialStyle(granio.SurfacePopup, VariantLight).Opaque)
	adapted = m.(granio.SystemAppearanceTheme).WithSystemAppearance(granio.SystemAppearance{Available: true, Transparency: granio.SystemPreference[bool]{Available: true, Value: false}})
	require.True(t, adapted.(granio.MaterialTheme).MaterialStyle(granio.SurfaceCard, VariantLight).Opaque)
	for _, style := range []granio.TextStyle{{}, {Bold: true}, {Italic: true}, {Bold: true, Italic: true}, {Monospace: true}, {Symbol: true}} {
		system, err := NewModern(ModernOptions{FontPolicy: FontSystem})
		require.NoError(t, err)
		require.NotNil(t, system.Font(style))
	}
}

func TestModernSemanticContrastOnSurfacesAndTintedMaterial(t *testing.T) {
	for _, variant := range []granio.ThemeVariant{VariantLight, VariantDark} {
		m, err := NewModern(ModernOptions{})
		require.NoError(t, err)
		contrast := func(a, b color.Color) float64 {
			x, y := luminance(a), luminance(b)
			return (math.Max(x, y) + .05) / (math.Min(x, y) + .05)
		}
		for _, foreground := range []granio.ThemeColorName{ColorNameForeground, ColorNamePlaceHolder, ColorNameError, ColorNameSuccess, ColorNameWarning} {
			for _, background := range []granio.ThemeColorName{ColorNameBackground, ColorNameInputBackground, ColorNameMenuBackground} {
				require.GreaterOrEqual(t, contrast(m.Color(foreground, variant), m.Color(background, variant)), 4.5, "%s on %s", foreground, background)
			}
		}
		for foreground, background := range map[granio.ThemeColorName]granio.ThemeColorName{ColorNameForegroundOnError: ColorNameError, ColorNameForegroundOnSuccess: ColorNameSuccess, ColorNameForegroundOnWarning: ColorNameWarning} {
			require.GreaterOrEqual(t, contrast(m.Color(foreground, variant), m.Color(background, variant)), 4.5)
		}
		style := m.(granio.MaterialTheme).MaterialStyle(granio.SurfacePopup, variant)
		tint := color.NRGBAModel.Convert(style.Tint).(color.NRGBA)
		tint.A = uint8(style.TintOpacity * 255)
		for _, source := range []color.Color{color.Black, color.White, color.NRGBA{R: 255, A: 255}, color.NRGBA{G: 255, A: 255}, color.NRGBA{B: 255, A: 255}} {
			composite := overlayVisualColor(source, tint)
			require.GreaterOrEqual(t, contrast(m.Color(ColorNameForeground, variant), composite), 4.5)
		}
	}
}
