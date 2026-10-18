package app

import (
	"context"
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/theme"
)

func TestSystemAppearancePreservesOverridesAndMotionPolicy(t *testing.T) {
	s := &settings{variant: theme.VariantLight, theme: theme.DarkTheme(), themeSpecified: true, schema: SettingsSchema{ThemeName: "light"}}
	themeObject := s.theme
	snapshot := granio.SystemAppearance{Generation: 1, Available: true, ColorScheme: granio.SystemPreference[granio.SystemColorScheme]{Value: granio.SystemColorDark, Available: true}, ReducedMotion: granio.SystemPreference[bool]{Value: true, Available: true}}
	s.applySystemAppearance(snapshot)
	if s.ThemeVariant() != theme.VariantLight || s.Theme() != themeObject {
		t.Fatal("system appearance replaced an explicit user preference")
	}
	if s.ShowAnimations() {
		t.Fatal("system reduced motion ignored")
	}
	s.schema.DisableAnimations = true
	snapshot.ReducedMotion.Value = false
	snapshot.Generation++
	s.applySystemAppearance(snapshot)
	if s.ShowAnimations() {
		t.Fatal("system appearance enabled explicitly disabled animations")
	}
	s.schema.ThemeName = ""
	s.applySystemAppearance(snapshot)
	if s.ThemeVariant() != theme.VariantDark || s.Theme() != themeObject {
		t.Fatal("automatic color scheme failed or replaced custom theme")
	}
	if s.SystemAppearance().TextScale.Available || s.SystemAppearance().Transparency.Available {
		t.Fatal("unavailable OS preferences invented")
	}
	s.stopSystemAppearance()
	s.stopSystemAppearance()
}

func TestSystemAppearanceModernThemeUpdates(t *testing.T) {
	modern, err := theme.NewModern(theme.ModernOptions{FollowSystem: true})
	if err != nil {
		t.Fatal(err)
	}
	s := &settings{theme: modern, variant: theme.VariantLight, themeSpecified: true, schema: SettingsSchema{ThemeName: "light"}}
	var changes int
	s.AddListener(func(granio.Settings) { changes++ })
	snapshot := granio.SystemAppearance{Available: true, Generation: 1,
		ColorScheme: granio.SystemPreference[granio.SystemColorScheme]{Value: granio.SystemColorDark, Available: true},
		TextScale:   granio.SystemPreference[float32]{Value: 1.5, Available: true},
		Accent:      granio.SystemPreference[color.NRGBA]{Value: color.NRGBA{R: 180, B: 80, A: 255}, Available: true}}
	s.applySystemAppearance(snapshot)
	if changes != 1 || s.ThemeVariant() != theme.VariantLight || s.Theme().Size(theme.SizeNameText) != 21 {
		t.Fatal("appearance was not applied coherently with the variant override")
	}
	if s.Theme().Color(theme.ColorNamePrimary, theme.VariantLight) != snapshot.Accent.Value {
		t.Fatal("system accent ignored")
	}
	// Installing another theme after startup applies the current snapshot too.
	s.SetTheme(modern)
	if s.Theme().Size(theme.SizeNameText) != 21 {
		t.Fatal("new theme lost the OS text scale")
	}
	s.applySystemAppearance(granio.SystemAppearance{Generation: 2})
	if s.Theme().Size(theme.SizeNameText) != 14 {
		t.Fatal("provider loss did not restore application defaults")
	}
}

func TestSystemAppearanceCoalescesAndCancelsUIWork(t *testing.T) {
	s := &settings{}
	var queued []func()
	enqueue := func(fn func()) { queued = append(queued, fn) }
	for i := range 1000 {
		snapshot := granio.SystemAppearance{Generation: uint64(i + 1), Available: true, ColorScheme: granio.SystemPreference[granio.SystemColorScheme]{Value: granio.SystemColorDark, Available: true}}
		s.appearance.Store(&snapshot)
		s.dispatchLatestAppearance(context.Background(), enqueue)
	}
	if len(queued) != 1 {
		t.Fatalf("queued %d UI callbacks", len(queued))
	}
	queued[0]()
	if s.ThemeVariant() != theme.VariantDark || s.SystemAppearance().Generation != 1000 {
		t.Fatal("latest update lost")
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.dispatchLatestAppearance(ctx, enqueue)
	cancel()
	s.variant = theme.VariantLight
	queued[1]()
	if s.ThemeVariant() != theme.VariantLight {
		t.Fatal("canceled callback touched UI")
	}
}
