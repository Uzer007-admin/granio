//go:build !ci && !wasm && !test_web_driver && !android && !ios && !mobile && (linux || openbsd || freebsd || netbsd) && !tinygo && !noos && !tamago

package app

import (
	"image/color"
	"testing"

	"github.com/Uzer007-admin/dbus/v5/portal"
)

func TestPortalAppearanceAvailability(t *testing.T) {
	parsed, err := portal.ParseAppearance(nil)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := portalAppearance(portal.AppearanceSnapshot{Generation: 5, Available: true, Appearance: parsed})
	if !snapshot.Available || snapshot.Generation != 5 || snapshot.ColorScheme.Available || snapshot.Accent.Available || snapshot.TextScale.Available || snapshot.Transparency.Available {
		t.Fatal(snapshot)
	}
	appearance := portal.Appearance{HighContrast: portal.Preference[bool]{Value: true, Available: true}, Accent: portal.Preference[portal.RGB]{Value: portal.RGB{Red: 1, Green: 0.5, Blue: 0}, Available: true}}
	snapshot = portalAppearance(portal.AppearanceSnapshot{Available: true, Appearance: appearance})
	if !snapshot.HighContrast.Value || snapshot.Accent.Value.A != 255 || snapshot.Accent.Value.R != 255 || snapshot.Accent.Value.G != 128 || snapshot.Accent.Value.B != 0 {
		t.Fatal(snapshot)
	}
}

func TestPortalDesktopThemeAndPaletteAvailability(t *testing.T) {
	a := portal.Appearance{Desktop: portal.DesktopAppearance{
		Name:        portal.Preference[string]{Value: "BreezeDark", Available: true},
		WidgetStyle: portal.Preference[string]{Value: "Union", Available: true},
		Colors:      portal.Preference[portal.DesktopPalette]{Available: true, Value: portal.DesktopPalette{Window: color.NRGBA{R: 31, G: 32, B: 33, A: 255}}},
	}}
	snapshot := portalAppearance(portal.AppearanceSnapshot{Available: true, Appearance: a})
	if !snapshot.DesktopTheme.Available || snapshot.DesktopTheme.Value.Name != "BreezeDark" || snapshot.DesktopTheme.Value.WidgetStyle != "Union" || !snapshot.Colors.Available || snapshot.Colors.Value.Window.R != 31 || snapshot.TextScale.Available || snapshot.Transparency.Available {
		t.Fatal(snapshot)
	}
	snapshot = portalAppearance(portal.AppearanceSnapshot{Available: false, Appearance: a})
	if snapshot.DesktopTheme.Available || snapshot.Colors.Available {
		t.Fatal(snapshot)
	}
}
