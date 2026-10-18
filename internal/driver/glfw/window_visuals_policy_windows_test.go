//go:build windows && !no_glfw && !wasm && !test_web_driver && !mobile

package glfw

import (
	"errors"
	"testing"

	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/stretchr/testify/require"
)

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
