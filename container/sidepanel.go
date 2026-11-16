package container

import (
	"math"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/internal/cache"
	"github.com/Uzer007-admin/granio/v3/theme"
	"github.com/Uzer007-admin/granio/v3/widget"
)

// PanelSide defines the horizontal edge where a side panel opens.
type PanelSide uint8

const (
	// PanelLeading places the panel to the left of the main content.
	PanelLeading PanelSide = iota
	// PanelTrailing places the panel to the right of the main content.
	PanelTrailing
)

// SidePanel arranges two canvas objects and animates the side panel's reveal.
// The panel is clipped during motion. Its children retain their layout width
// throughout the transition. The main content receives the available space.
type SidePanel struct {
	widget.BaseWidget
	content, panel granio.CanvasObject
	side           PanelSide
	width, reveal  float32
	open           bool
	transition     *granio.Animation
}

// NewSidePanel creates an open panel with a preferred width in logical units.
func NewSidePanel(content, panel granio.CanvasObject, side PanelSide, width float32) *SidePanel {
	p := &SidePanel{content: content, panel: panel, side: side, open: true, reveal: 1}
	p.ExtendBaseWidget(p)
	p.SetPanelWidth(width)
	return p
}

// PanelWidth returns the preferred panel width before constraints from the container.
func (p *SidePanel) PanelWidth() float32 { return p.width }

// Side returns the edge on which the panel is placed.
func (p *SidePanel) Side() PanelSide { return p.side }

// SetSide moves the panel to another edge while retaining its current reveal.
func (p *SidePanel) SetSide(side PanelSide) {
	if side == p.side || (side != PanelLeading && side != PanelTrailing) {
		return
	}
	p.side = side
	p.updatePanelLayout()
}

// SetPanelWidth changes the preferred width. It can be used by a drag handle.
func (p *SidePanel) SetPanelWidth(width float32) {
	if width < 0 || math.IsNaN(float64(width)) || math.IsInf(float64(width), 0) || p.width == width {
		return
	}
	p.width = width
	p.updatePanelLayout()
}

// PanelOpen reports the requested state, including while a transition runs.
func (p *SidePanel) PanelOpen() bool { return p.open }

// SetPanelOpen opens or closes the panel. Reversing during a transition starts
// at the visible width. Reduced-motion settings apply to the transition.
func (p *SidePanel) SetPanelOpen(open bool) {
	if p.open == open {
		return
	}
	p.open = open
	p.stopPanelTransition()
	target := float32(0)
	if open {
		target = 1
	}
	app := granio.CurrentApp()
	if app == nil || !p.Visible() || !cache.IsRendered(p) || !app.Settings().ShowAnimations() || target == p.reveal {
		p.reveal = target
		p.updatePanelLayout()
		return
	}
	spec := theme.Motion(p.Theme(), granio.MotionNavigation)
	from := p.reveal
	a := granio.NewAnimation(spec.Duration, func(value float32) {
		p.reveal = from + (target-from)*value
		p.updatePanelLayout()
	})
	a.Object, a.Disabled, a.Curve = p, spec.Disabled, spec.Curve
	a.OnFinished = func() {
		if p.transition == a {
			p.transition = nil
		}
	}
	p.transition = a
	a.Start()
}

func (p *SidePanel) stopPanelTransition() {
	if p.transition == nil {
		return
	}
	if granio.CurrentApp() != nil {
		p.transition.Stop()
	}
	p.transition = nil
}

func (p *SidePanel) updatePanelLayout() {
	if r, ok := cache.CachedRenderer(p); ok {
		r.Layout(p.Size())
	}
}

// CreateRenderer implements granio.Widget.
func (p *SidePanel) CreateRenderer() granio.WidgetRenderer {
	clip := NewClip(NewWithoutLayout(p.panel))
	return &sidePanelRenderer{owner: p, clip: clip, objects: []granio.CanvasObject{p.content, clip}}
}

type sidePanelRenderer struct {
	owner   *SidePanel
	clip    *Clip
	objects []granio.CanvasObject
}

func (r *sidePanelRenderer) MinSize() granio.Size           { return r.owner.content.MinSize() }
func (r *sidePanelRenderer) Objects() []granio.CanvasObject { return r.objects }
func (r *sidePanelRenderer) Refresh()                       { r.Layout(r.owner.Size()) }

func (r *sidePanelRenderer) Destroy() {
	p := r.owner
	p.stopPanelTransition()
	p.reveal = 0
	if p.open {
		p.reveal = 1
	}
}

func (r *sidePanelRenderer) Layout(size granio.Size) {
	p := r.owner
	width := min(p.width, max(0, size.Width-p.content.MinSize().Width))
	visible := max(0, min(width, width*p.reveal))
	contentWidth := max(0, size.Width-visible)
	contentX, clipX, panelX := float32(0), contentWidth, float32(0)
	if p.side == PanelLeading {
		contentX, clipX, panelX = visible, 0, visible-width
	}
	p.content.Move(granio.NewPos(contentX, 0))
	p.content.Resize(granio.NewSize(contentWidth, size.Height))
	if visible <= 0 {
		r.clip.Hide()
		return
	}
	p.panel.Move(granio.NewPos(panelX, 0))
	p.panel.Resize(granio.NewSize(width, size.Height))
	r.clip.Move(granio.NewPos(clipX, 0))
	r.clip.Resize(granio.NewSize(visible, size.Height))
	r.clip.Show()
}
