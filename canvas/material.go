package canvas

import (
	"errors"
	"image/color"
	"math"
	"sync/atomic"

	"github.com/Uzer007-admin/granio/v3"
)

// Material filters the already painted lower canvas layers. It never captures
// other windows or the desktop. Native GL/GLES painters share a bounded pool of
// two material layers per window. Opted-in software blur uses a smaller budget.
// Unsupported or disabled effects paint Fallback. A material does not request
// animation frames.
type Material struct {
	baseObject
	Enabled                         bool
	BlurRadius, CornerRadius, Noise float32
	Tint                            color.NRGBA
	Fallback                        color.Color
	Downsample                      uint8 // 1, 2, 4 or 8; zero selects 2
	Linear                          bool
	// Live samples the backdrop on each painted frame. Use it above content
	// that scrolls or moves without invalidating its textures.
	Live bool
	// TranslucentFallback retains the alpha of Fallback when effects are
	// allowed but the painter cannot blur. Disabled effects stay opaque.
	TranslucentFallback bool
	// SoftwareBlur opts into a bounded, reduced resolution blur on CPU renderers.
	SoftwareBlur bool
	// Priority reserves room in the GL material pool for a primary surface.
	Priority bool
	// TrackBackdrop lets the owner explicitly invalidate lower-layer movement.
	// Live still takes precedence. Use InvalidateBackdrop after lower-layer edits.
	TrackBackdrop     bool
	backdropRevision  uint64
	effect            atomic.Uint32
	effectChanged     atomic.Pointer[func(MaterialEffect)]
	corners           granio.CornerRadii
	hasCorners        bool
	fallbackRectangle Rectangle
}

// MaterialEffect describes the last effect selected by the window painter.
type MaterialEffect uint32

const (
	MaterialPending MaterialEffect = iota
	MaterialTint
	MaterialOpaque
	MaterialHardwareBlur
	MaterialSoftwareBlur
)

// Effect reports the last painted effect.
func (m *Material) Effect() MaterialEffect { return MaterialEffect(m.effect.Load()) }

// SetEffect reports the effect selected by a painter.
func (m *Material) SetEffect(effect MaterialEffect) {
	if m.effect.Swap(uint32(effect)) == uint32(effect) {
		return
	}
	if callback := m.effectChanged.Load(); callback != nil && granio.CurrentApp() != nil {
		granio.Do(func() {
			if m.effectChanged.Load() == callback && m.Effect() == effect {
				(*callback)(effect)
			}
		})
	}
}

// SetOnEffectChanged installs a UI-thread callback for changes to Effect.
// Passing nil removes the callback.
func (m *Material) SetOnEffectChanged(callback func(MaterialEffect)) {
	if callback == nil {
		m.effectChanged.Store(nil)
		return
	}
	m.effectChanged.Store(&callback)
}

// ApplyOptics updates the tint transparency (0 through 1) and blur radius
// (0 through 50 logical units) together on the GUI thread.
func (m *Material) ApplyOptics(transparency, radius float32) error {
	if !(transparency >= 0 && transparency <= 1 && radius >= 0 && radius <= 50) {
		return errors.New("canvas: transparency must be in [0,1] and blur radius in [0,50]")
	}
	alpha := uint8(math.Round(float64(1-transparency) * 255))
	changed := m.Tint.A != alpha || m.BlurRadius != radius
	m.Tint.A, m.BlurRadius = alpha, radius
	if m.TranslucentFallback {
		fallback := m.Tint
		if m.Fallback != nil {
			fallback = color.NRGBAModel.Convert(m.Fallback).(color.NRGBA)
			changed = changed || fallback.A != alpha
		}
		fallback.A = alpha
		m.Fallback = fallback
	}
	if changed {
		repaint(m)
	}
	return nil
}

// SetCornerRadii overrides the material shape. A nil value restores CornerRadius.
func (m *Material) SetCornerRadii(corners *granio.CornerRadii) {
	if corners == nil {
		if !m.hasCorners {
			return
		}
		m.corners, m.hasCorners = granio.CornerRadii{}, false
	} else {
		if m.hasCorners && m.corners == *corners {
			return
		}
		m.corners, m.hasCorners = *corners, true
	}
	repaint(m)
}

// EffectiveCornerRadii returns the drawable mask geometry in logical units.
func (m *Material) EffectiveCornerRadii() granio.CornerRadii {
	if m.hasCorners {
		return m.corners.Normalized(m.Size())
	}
	r := m.CornerRadius
	return (granio.CornerRadii{TopLeft: r, TopRight: r, BottomRight: r, BottomLeft: r}).Normalized(m.Size())
}

// InvalidateBackdrop schedules a repaint after the lower layers have changed.
func (m *Material) InvalidateBackdrop() {
	m.backdropRevision++
	repaint(m)
}

// BackdropRevision identifies the explicitly tracked lower-layer contents.
func (m *Material) BackdropRevision() uint64 { return m.backdropRevision }

// CanPaintTint reports whether the result can be drawn without backdrop sampling.
func (m *Material) CanPaintTint() bool { return m.Noise == 0 && (m.BlurRadius == 0 || m.Tint.A == 255) }

// TintRectangle returns the borrowed geometry for a tint-only material pass.
func (m *Material) TintRectangle() *Rectangle { return m.rectangle(m.Tint) }

func NewMaterial(fallback color.Color) *Material {
	return &Material{Enabled: true, BlurRadius: 12, CornerRadius: 8, Tint: color.NRGBA{R: 255, G: 255, B: 255, A: 160}, Fallback: fallback, Downsample: 2}
}
func (m *Material) Refresh() { Refresh(m) }
func (m *Material) Hide()    { m.baseObject.Hide(); repaint(m) }
func (m *Material) Move(pos granio.Position) {
	if pos != m.Position() {
		m.baseObject.Move(pos)
		repaint(m)
	}
}
func (m *Material) Resize(size granio.Size) {
	if size != m.Size() {
		m.baseObject.Resize(size)
		m.Refresh()
	}
}

// OpaqueFallback is defined even for nil or translucent caller colors.
func (m *Material) OpaqueFallback() color.Color {
	if m.Fallback == nil {
		return color.NRGBA{R: 240, G: 240, B: 240, A: 255}
	}
	c := color.NRGBAModel.Convert(m.Fallback).(color.NRGBA)
	c.A = 255
	return c
}
func (m *Material) Valid() bool {
	finite := func(v float32) bool { return !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) }
	validCorners := m.hasCorners || finite(m.CornerRadius) && m.CornerRadius >= 0
	return validCorners && finite(m.BlurRadius) && finite(m.Noise) && m.BlurRadius >= 0 && m.BlurRadius <= 50 && m.Noise >= 0 && m.Noise <= 1 && (m.Downsample == 0 || m.Downsample == 1 || m.Downsample == 2 || m.Downsample == 4 || m.Downsample == 8)
}

// EffectsAllowed reports whether the material can sample or show its backdrop
// with the current theme and system appearance.
func (m *Material) EffectsAllowed() bool {
	if !m.Enabled || !m.Valid() {
		return false
	}
	app := granio.CurrentApp()
	if app == nil {
		return true
	}
	settings := app.Settings()
	if theme, ok := settings.Theme().(interface{ HighContrast() bool }); ok && theme.HighContrast() {
		return false
	}
	if system, ok := settings.(granio.SystemAppearanceProvider); ok {
		appearance := system.SystemAppearance()
		if appearance.HighContrast.Available && appearance.HighContrast.Value ||
			appearance.Transparency.Available && !appearance.Transparency.Value {
			return false
		}
	}
	return true
}

// FallbackRectangle returns a borrowed primitive with matching geometry.
// Painters use it without registering another canvas object or refreshing a tree.
func (m *Material) FallbackRectangle() *Rectangle {
	fill := m.OpaqueFallback()
	if m.TranslucentFallback && m.Fallback != nil && m.EffectsAllowed() {
		fill = m.Fallback
	}
	return m.rectangle(fill)
}

func (m *Material) rectangle(fill color.Color) *Rectangle {
	r := &m.fallbackRectangle
	r.FillColor = fill
	r.CornerRadius = m.CornerRadius
	if math.IsNaN(float64(r.CornerRadius)) || math.IsInf(float64(r.CornerRadius), 0) || r.CornerRadius < 0 {
		r.CornerRadius = 0
	}
	r.corners, r.hasCorners = m.corners, m.hasCorners
	r.baseObject.Resize(m.Size())
	return r
}
