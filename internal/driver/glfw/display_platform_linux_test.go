//go:build linux && !wasm && !test_web_driver

package glfw

import "testing"

func TestDisplaySelectionDoesNotProbeCompositor(t *testing.T) {
	t.Setenv("WAYLAND_DISPLAY", "unavailable-display")
	t.Setenv("DISPLAY", ":999")
	for _, value := range []string{"", "wayland", "x11", "invalid"} {
		t.Setenv("FYNE_PLATFORM", value)
		want := value
		if value == "invalid" {
			want = platformAuto
		}
		if got := forcePlatform(); got != want {
			t.Fatalf("%q: got %q, want %q", value, got, want)
		}
	}
}
