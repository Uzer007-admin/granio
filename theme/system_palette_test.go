package theme

import (
	"github.com/Uzer007-admin/granio/v2"
	"image/color"
	"testing"
)

func TestModernSystemPaletteAndExplicitAccent(t *testing.T) {
	palette := granio.SystemColors{Window: color.NRGBA{R: 21, G: 22, B: 23, A: 255}, WindowText: color.NRGBA{R: 240, G: 241, B: 242, A: 255}, ButtonFace: color.NRGBA{R: 31, G: 32, B: 33, A: 255}, ButtonText: color.NRGBA{R: 250, G: 251, B: 252, A: 255}, Highlight: color.NRGBA{R: 40, G: 90, B: 180, A: 255}, HighlightText: color.NRGBA{R: 255, G: 255, B: 255, A: 255}, GrayText: color.NRGBA{R: 170, G: 171, B: 172, A: 255}}
	snapshot := granio.SystemAppearance{Available: true, Colors: granio.SystemPreference[granio.SystemColors]{Available: true, Value: palette}}
	base, err := NewModern(ModernOptions{FollowSystem: true, FollowSystemPalette: true})
	if err != nil {
		t.Fatal(err)
	}
	system := base.(granio.SystemAppearanceTheme).WithSystemAppearance(snapshot)
	if system.Color(ColorNameBackground, VariantDark) != palette.Window || system.Color(ColorNameForegroundOnButton, VariantDark) != palette.ButtonText {
		t.Fatal("palette not applied")
	}
	accent := color.NRGBA{R: 120, G: 30, B: 80, A: 255}
	base, _ = NewModern(ModernOptions{FollowSystemPalette: true, Accent: accent})
	system = base.(granio.SystemAppearanceTheme).WithSystemAppearance(snapshot)
	if system.Color(ColorNamePrimary, VariantDark) != accent {
		t.Fatal("explicit accent lost")
	}
	fallback := system.(granio.SystemAppearanceTheme).WithSystemAppearance(granio.SystemAppearance{})
	if fallback.Color(ColorNameBackground, VariantDark) == palette.Window {
		t.Fatal("stale palette")
	}
	if base.Color(ColorNameBackground, VariantDark) == palette.Window {
		t.Fatal("theme mutated")
	}
}
