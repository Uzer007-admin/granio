package granio

import (
	"image/color"
	"math"
	"time"
)

// TextRole is a semantic size/style. Default preserves explicit SizeName and
// TextStyle; it does not reinterpret a low-level canvas.Text's explicit size.
type TextRole uint8

const (
	TextRoleDefault TextRole = iota
	TextRoleBody
	TextRoleCaption
	TextRoleEmphasis
	TextRoleSubtitle
	TextRoleTitle
	TextRoleCode
)

type TypographyStyle struct {
	Size, LineSpacing float32
	Style             TextStyle
	Font              Resource
}
type TypographyTheme interface {
	Theme
	Typography(TextRole) TypographyStyle
}

type MotionRole uint8

const (
	MotionHover MotionRole = iota
	MotionPressed
	MotionSurface
	MotionNavigation
)

type MotionSpec struct {
	Duration time.Duration
	Curve    AnimationCurve
	Disabled bool
}
type MotionTheme interface {
	Theme
	Motion(MotionRole) MotionSpec
}

// VisualState is composable: readonly, error and focus may coexist. Disabled
// suppresses hover/pressed; error and keyboard focus retain separate outlines.
type VisualState uint16

const (
	VisualHovered VisualState = 1 << iota
	VisualPressed
	VisualFocused
	VisualDisabled
	VisualReadOnly
	VisualSelected
	VisualError
	VisualInactive
)

type ComponentKind uint8

const (
	ComponentButton ComponentKind = iota
	ComponentInput
	ComponentChoice
	ComponentContent
	ComponentCard
	ComponentPopup
	ComponentSelection
)

type ComponentStyle struct {
	Background, Foreground, Border, Focus          color.Color
	Radius, BorderWidth, FocusWidth, MinimumHeight float32
}
type ComponentTheme interface {
	Theme
	ComponentStyle(ComponentKind, VisualState, ThemeVariant) ComponentStyle
}

// CornerRadii specifies four independent radii in logical units.
// Zero makes that corner square.
type CornerRadii struct {
	TopLeft, TopRight, BottomRight, BottomLeft float32
}

// Normalized fits the radii to the drawable rectangle.
func (r CornerRadii) Normalized(size Size) CornerRadii {
	limit := max(0, min(size.Width, size.Height)*0.5)
	clamp := func(radius float32) float32 {
		if radius < 0 || math.IsNaN(float64(radius)) || math.IsInf(float64(radius), 0) {
			return 0
		}
		return min(radius, limit)
	}
	return CornerRadii{clamp(r.TopLeft), clamp(r.TopRight), clamp(r.BottomRight), clamp(r.BottomLeft)}
}

// CornerTheme optionally overrides the shape of individual widget groups.
type CornerTheme interface {
	Theme
	CornerRadii(ComponentKind, ThemeVariant) (CornerRadii, bool)
}

// ControlOpticsTheme can supply independent tint transparency [0,1] and blur
// radius [0,50] for the backgrounds of buttons and entries.
type ControlOpticsTheme interface {
	Theme
	ControlOptics(ComponentKind, ThemeVariant) (transparency, blurRadius float32, explicit bool)
}

type SurfaceRole uint8

const (
	SurfaceWindow SurfaceRole = iota
	SurfaceContent
	SurfaceCard
	SurfacePopup
)

// MaterialStyle describes in-canvas materials only. A native window backdrop
// continues to use desktop.WindowVisuals, independently of these tokens.
type MaterialStyle struct {
	BlurRadius, Noise, TintOpacity, CornerRadius float32
	Tint, Fallback                               color.Color
	Opaque                                       bool
}
type MaterialTheme interface {
	Theme
	MaterialStyle(SurfaceRole, ThemeVariant) MaterialStyle
}

// ThemeChangeAware is optional for widgets that preserve transient state while
// Fyne reflows their owned renderer objects. The returned completion runs after
// the child traversal, layout and Refresh, on the same UI goroutine.
type ThemeChangeAware interface {
	BeginThemeChange() func()
}
