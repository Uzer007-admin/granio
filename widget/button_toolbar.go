package widget

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/internal/cache"
	"github.com/Uzer007-admin/granio/v3/theme"
)

func (b *Button) buttonTextStyle() RichTextStyle {
	style := RichTextStyleStrong
	if b.Style == ButtonStyleToolbar {
		style = RichTextStyleInline
	}
	style.Alignment = granio.TextAlignCenter
	return style
}

func (b *Button) buttonTextInset(th granio.Theme) granio.Size {
	if b.Style == ButtonStyleToolbar {
		return granio.Size{}
	}
	return granio.NewSquareSize(th.Size(theme.SizeNameInnerPadding))
}

func (b *Button) refreshInteraction() {
	if b.Style == ButtonStyleToolbar {
		if renderer, ok := cache.CachedRenderer(b.super()); ok {
			if r, ok := renderer.(*buttonRenderer); ok {
				r.applyToolbarBackground()
				return
			}
		}
	}
	b.Refresh()
}

func (r *buttonRenderer) applyToolbarBackground() {
	b, bg := r.button, r.background
	th := b.Theme()
	variant := granio.CurrentApp().Settings().ThemeVariant()
	target := color.Color(color.Transparent)
	if b.Selected {
		target = th.Color(theme.ColorNameSelection, variant)
	}
	if b.Importance == HighImportance {
		target = th.Color(theme.ColorNamePrimary, variant)
		if b.Disabled() {
			target = th.Color(theme.ColorNameDisabledButton, variant)
		}
	}
	if !b.Disabled() {
		switch {
		case b.Importance == DangerImportance && (b.hovered || b.pressed):
			target = th.Color(theme.ColorNameError, variant)
		case b.pressed:
			target = blendColor(target, th.Color(theme.ColorNamePressed, variant))
		case b.hovered:
			target = blendColor(target, th.Color(theme.ColorNameHover, variant))
		}
	}
	bg.CornerRadius = th.Size(theme.SizeNameButtonRadius)
	corners := theme.Corners(th, granio.ComponentButton, variant)
	r.updateStateBackground(r.styleBackground(th, variant, target, corners), th)
	bg.SetCornerRadii(corners)
	r.tapBG.SetCornerRadii(corners)
	bg.StrokeColor, bg.StrokeWidth = nil, 0
	if colors, ok := activeSystemColors(th); ok {
		bg.StrokeColor, bg.StrokeWidth = colors.WindowText, 1
	}
	if b.focused && !b.Disabled() {
		bg.StrokeColor, bg.StrokeWidth = th.Color(theme.ColorNameFocus, variant), 1
	}
	canvas.Refresh(bg)
}
