//go:build !wasm && !test_web_driver

package glfw

import (
	"testing"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
)

func TestModifierRepeatRemainsPressed(t *testing.T) {
	for _, tc := range []struct {
		name string
		key  glfw.Key
		mask glfw.ModifierKey
		want granio.KeyModifier
	}{
		{"left control", glfw.KeyLeftControl, glfw.ModControl, granio.KeyModifierControl},
		{"right control", glfw.KeyRightControl, glfw.ModControl, granio.KeyModifierControl},
		{"left shift", glfw.KeyLeftShift, glfw.ModShift, granio.KeyModifierShift},
		{"right shift", glfw.KeyRightShift, glfw.ModShift, granio.KeyModifierShift},
		{"left alt", glfw.KeyLeftAlt, glfw.ModAlt, granio.KeyModifierAlt},
		{"left super", glfw.KeyLeftSuper, glfw.ModSuper, granio.KeyModifierSuper},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := desktopModifierCorrected(0, tc.key, glfw.Press); got != tc.want {
				t.Fatalf("Press: got %v, want %v", got, tc.want)
			}
			for _, mods := range []glfw.ModifierKey{0, tc.mask} {
				if got := desktopModifierCorrected(mods, tc.key, glfw.Repeat); got != tc.want {
					t.Fatalf("Repeat with mask %v: got %v, want %v", mods, got, tc.want)
				}
			}
			if got := desktopModifierCorrected(tc.mask, tc.key, glfw.Release); got != 0 {
				t.Fatalf("Release: got %v, want 0", got)
			}
		})
	}
}
