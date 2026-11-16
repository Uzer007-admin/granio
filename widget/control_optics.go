package widget

import (
	"image/color"
	"math"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	intTheme "github.com/Uzer007-admin/granio/v3/internal/theme"
)

// Keep the blur below the existing fill and vector focus outline. The fill
// can still animate independently of the expensive backdrop pass.
func controlOptics(th granio.Theme, kind granio.ComponentKind, variant granio.ThemeVariant, fill color.Color,
	shape *canvas.Rectangle, radii *granio.CornerRadii, filter **canvas.Material,
) (tint color.Color, created bool) {
	provider, ok := intTheme.Optional[granio.ControlOpticsTheme](th)
	if !ok || fill == nil {
		if m := *filter; m != nil && m.Visible() {
			m.Hide()
		}
		return fill, false
	}
	transparency, radius, explicit := provider.ControlOptics(kind, variant)
	effect := canvas.Material{Enabled: true, BlurRadius: radius}
	if !explicit || !(transparency > 0 && transparency <= 1) || !effect.EffectsAllowed() {
		if m := *filter; m != nil && m.Visible() {
			m.Hide()
		}
		return fill, false
	}
	c := color.NRGBAModel.Convert(fill).(color.NRGBA)
	if radius > 0 && c.A > 0 {
		if *filter == nil {
			m := canvas.NewMaterial(color.Transparent)
			m.TranslucentFallback, m.Live, m.Downsample = true, true, 4
			*filter, created = m, true
		}
		m := *filter
		m.CornerRadius = shape.CornerRadius
		m.SetCornerRadii(radii)
		_ = m.ApplyOptics(1, radius)
		m.Move(shape.Position())
		m.Resize(shape.Size())
		if !m.Visible() {
			m.Show()
		}
	} else if m := *filter; m != nil && m.Visible() {
		m.Hide()
	}
	c.A = uint8(math.Round(float64(float32(c.A) * (1 - transparency))))
	return c, created
}
