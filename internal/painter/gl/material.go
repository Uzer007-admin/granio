package gl

import (
	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/internal"
)

type materialPipeline interface {
	Draw(*painter, *canvas.Material, granio.Position, granio.Size, *internal.ClipItem) bool
	ResetFrame()
	Close()
	Statistics() map[string]uint64
}

func (p *painter) drawMaterial(m *canvas.Material, pos granio.Position, frame granio.Size, clip *internal.ClipItem) {
	if m.EffectsAllowed() {
		if m.CanPaintTint() {
			p.drawRectangle(m.TintRectangle(), pos, frame)
			if m.Tint.A == 255 {
				m.SetEffect(canvas.MaterialOpaque)
			} else {
				m.SetEffect(canvas.MaterialTint)
			}
			return
		}
		if p.materials == nil && !p.materialUnavailable {
			p.materials = newMaterialPipeline(p)
			p.materialUnavailable = p.materials == nil
		}
		if p.materials != nil && p.materials.Draw(p, m, pos, frame, clip) {
			return
		}
	}
	fallback := m.FallbackRectangle()
	p.drawRectangle(fallback, pos, frame)
	_, _, _, alpha := fallback.FillColor.RGBA()
	mode := canvas.MaterialOpaque
	if alpha < 65535 {
		mode = canvas.MaterialTint
	}
	m.SetEffect(mode)
}

// ReleaseMaterials is called with the owning context current before Destroy.
func (p *painter) ReleaseMaterials() {
	if p.materials != nil {
		p.materials.Close()
		p.materials = nil
	}
}
func (p *painter) MaterialStatistics() map[string]uint64 {
	if p.materials == nil {
		return map[string]uint64{}
	}
	return p.materials.Statistics()
}
