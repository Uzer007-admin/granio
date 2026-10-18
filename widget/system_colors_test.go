package widget

import (
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/stretchr/testify/require"
)

func TestSystemPalettePreservesSelectionAndButtonText(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	p := granio.SystemColors{Window: color.NRGBA{R: 255, G: 255, B: 255, A: 255}, WindowText: color.NRGBA{A: 255},
		ButtonFace: color.NRGBA{A: 255}, ButtonText: color.NRGBA{R: 255, G: 255, B: 255, A: 255},
		Highlight: color.NRGBA{R: 30, G: 70, B: 90, A: 255}, HighlightText: color.NRGBA{G: 255, A: 255},
		GrayText: color.NRGBA{R: 100, G: 100, B: 100, A: 255}, Hotlight: color.NRGBA{B: 190, A: 255}}
	modern, err := theme.NewModern(theme.ModernOptions{}) // accessibility overrides do not need FollowSystem
	require.NoError(t, err)
	adapted := modern.(granio.SystemAppearanceTheme).WithSystemAppearance(granio.SystemAppearance{Available: true,
		HighContrast: granio.SystemPreference[bool]{Value: true, Available: true}, Colors: granio.SystemPreference[granio.SystemColors]{Value: p, Available: true}})
	a.Settings().SetTheme(adapted)
	button := NewButton("Контраст", nil)
	r := cache.Renderer(button).(*buttonRenderer)
	button.FocusGained()
	r.Refresh()
	require.Equal(t, p.ButtonFace, r.background.FillColor)
	require.Equal(t, p.Highlight, r.background.StrokeColor)
	require.EqualValues(t, 2, r.background.StrokeWidth)
	require.Equal(t, theme.ColorNameForegroundOnButton, r.label.Segments[0].(*TextSegment).Style.ColorName)
	require.Equal(t, p.ButtonText, adapted.Color(theme.ColorNameForegroundOnButton, theme.VariantLight))
	selectBox := NewSelect([]string{"Пункт"}, nil)
	selectView := cache.Renderer(selectBox).(*selectRenderer)
	selectBox.FocusGained()
	selectView.Refresh()
	require.Equal(t, p.Window, selectView.background.FillColor)
	require.Equal(t, p.Highlight, selectView.background.StrokeColor)
	require.EqualValues(t, 2, selectView.background.StrokeWidth)
	check := NewCheck("Флаг", nil)
	checkView := cache.Renderer(check).(*checkRenderer)
	check.FocusGained()
	checkView.Refresh()
	require.Equal(t, color.Transparent, checkView.focusIndicator.FillColor)
	require.Equal(t, p.Highlight, checkView.focusIndicator.StrokeColor)
	require.EqualValues(t, 2, checkView.focusIndicator.StrokeWidth)
	require.Equal(t, p.GrayText, shortcutColor(adapted))
	selection := canvas.NewRectangle(color.Black)
	applySelectionStyle(selection, adapted, theme.VariantLight)
	require.Equal(t, color.Transparent, selection.FillColor, "opaque selection can make custom child text unreadable")
	require.Equal(t, p.Highlight, selection.StrokeColor)
	require.EqualValues(t, 2, selection.StrokeWidth)
	applySelectionStyle(selection, modern, theme.VariantLight)
	require.Equal(t, modern.Color(theme.ColorNameSelection, theme.VariantLight), selection.FillColor)
	require.Zero(t, selection.StrokeWidth, "pooled selection retained a stale native outline")
}
