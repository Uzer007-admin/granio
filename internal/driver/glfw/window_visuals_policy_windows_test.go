//go:build windows && !no_glfw && !wasm && !test_web_driver && !mobile

package glfw

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v3/driver/desktop"
	"github.com/stretchr/testify/require"
)

func TestWindowsBackdropMaterialByBuild(t *testing.T) {
	for _, build := range []uint32{22000, 22621, 22631, 26100, 26200, 26300, 28000} {
		t.Run(fmt.Sprint(build), func(t *testing.T) {
			policy := windowVisualPolicy{build: build, accentPowerAllowed: true}
			for _, mode := range []struct {
				backdrop desktop.WindowBackdrop
				dwm      glfw.SystemBackdrop
			}{
				{desktop.BackdropOpaque, glfw.BackdropNone},
				{desktop.BackdropTransparent, glfw.BackdropNone},
				{desktop.BackdropMica, glfw.BackdropMainWindow},
				{desktop.BackdropMicaAlt, glfw.BackdropTabbedWindow},
				{desktop.BackdropDesktopAcrylic, glfw.BackdropTransientWindow},
				{desktop.BackdropSystemBlur, glfw.BackdropTransientWindow},
			} {
				native, accent := windowsBackdropMaterial(mode.backdrop, policy, true)
				require.Equal(t, mode.dwm, native, mode.backdrop)
				require.Equal(t, build == 26300 && mode.backdrop == desktop.BackdropDesktopAcrylic, accent, mode.backdrop)
				native, accent = windowsBackdropMaterial(mode.backdrop, policy, false)
				require.Equal(t, mode.dwm, native, mode.backdrop)
				require.False(t, accent, mode.backdrop)
				savingPower := policy
				savingPower.accentPowerAllowed = false
				native, accent = windowsBackdropMaterial(mode.backdrop, savingPower, true)
				require.Equal(t, mode.dwm, native, mode.backdrop)
				require.False(t, accent, mode.backdrop)
			}
		})
	}
}

func TestWindowVisualsPolicyPreflight(t *testing.T) {
	o := desktop.VisualOptions{Backdrop: desktop.BackdropMica}
	baseCaps := desktop.VisualCapabilities{OSBackdrops: 0xf, RendererBackdrops: 0xf, TransparentSurface: true}
	basePolicy := windowVisualPolicy{composition: true, transparent: true, transparencyKnown: true}
	for _, test := range []struct {
		name   string
		change func(*desktop.VisualCapabilities, *windowVisualPolicy)
		err    error
		want   desktop.VisualFallbackReason
	}{
		{"supported", nil, nil, desktop.VisualReasonNone},
		{"query_failure", nil, errors.New("native query failed"), desktop.VisualUnsupportedOS},
		{"old_OS", func(c *desktop.VisualCapabilities, _ *windowVisualPolicy) { c.OSBackdrops = 1 }, nil, desktop.VisualUnsupportedOS},
		{"high_contrast", func(_ *desktop.VisualCapabilities, p *windowVisualPolicy) { p.highContrast = true }, nil, desktop.VisualHighContrast},
		{"transparency_disabled", func(_ *desktop.VisualCapabilities, p *windowVisualPolicy) { p.transparent = false }, nil, desktop.VisualTransparencyDisabled},
		{"remote", func(_ *desktop.VisualCapabilities, p *windowVisualPolicy) { p.remote = true }, nil, desktop.VisualRemoteSession},
		{"composition_disabled", func(_ *desktop.VisualCapabilities, p *windowVisualPolicy) { p.composition = false }, nil, desktop.VisualUnsupportedBackend},
		{"renderer_unavailable", func(c *desktop.VisualCapabilities, _ *windowVisualPolicy) { c.RendererBackdrops = 1 }, nil, desktop.VisualUnsupportedBackend},
		{"opaque_surface", func(c *desktop.VisualCapabilities, _ *windowVisualPolicy) { c.TransparentSurface = false }, nil, desktop.VisualSurfaceNotTransparent},
		{"transparency_unknown", func(_ *desktop.VisualCapabilities, p *windowVisualPolicy) {
			p.transparent = false
			p.transparencyKnown = false
		}, nil, desktop.VisualReasonNone},
	} {
		t.Run(test.name, func(t *testing.T) {
			c, p := baseCaps, basePolicy
			if test.change != nil {
				test.change(&c, &p)
			}
			require.Equal(t, test.want, windowBackdropRejection(o, p, &c, test.err))
			require.Equal(t, desktop.VisualReasonNone, windowBackdropRejection(desktop.VisualOptions{}, p, &c, test.err))
		})
	}
}
