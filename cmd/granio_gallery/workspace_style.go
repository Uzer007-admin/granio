package main

import (
	"fmt"
	"image/color"
	"sync"
	"time"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/theme"
)

var (
	workspaceBase     = workspaceColor(0x181818)
	workspaceRail     = workspaceColor(0x1a2030)
	workspaceSidebar  = workspaceColor(0x191b21)
	workspaceTop      = workspaceColor(0x1e2023)
	workspaceBorder   = workspaceColor(0x2b2b2b)
	workspaceText     = workspaceColor(0xebebed)
	workspaceMuted    = workspaceColor(0xa2a3aa)
	workspaceSubtle   = workspaceColor(0x838389)
	workspaceSelected = workspaceColor(0x2b2d34)
	workspaceRaised   = workspaceColor(0x343434)
	workspaceWhite    = workspaceColor(0xfafafa)
)

func workspaceColor(rgb uint32) color.NRGBA {
	return color.NRGBA{R: uint8(rgb >> 16), G: uint8(rgb >> 8), B: uint8(rgb), A: 255}
}

type workspaceTheme struct {
	granio.Theme
	composer   bool
	round      bool
	appearance *workspaceAppearance
}

func (t workspaceTheme) Color(name granio.ThemeColorName, _ granio.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return workspaceBase
	case theme.ColorNameForeground, theme.ColorNameForegroundOnButton:
		return workspaceText
	case theme.ColorNameForegroundOnPrimary:
		return workspaceBase
	case theme.ColorNameDisabled, theme.ColorNamePlaceHolder:
		return workspaceSubtle
	case theme.ColorNameInputBackground:
		if t.composer {
			return color.Transparent
		}
		return workspaceRaised
	case theme.ColorNameInputBorder, theme.ColorNameFocus:
		if t.composer {
			return color.Transparent
		}
		return workspaceColor(0x72757e)
	case theme.ColorNameButton, theme.ColorNameMenuBackground:
		return workspaceTop
	case theme.ColorNameSelection:
		return workspaceSelected
	case theme.ColorNameHover:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 15}
	case theme.ColorNamePressed:
		return color.NRGBA{R: 255, G: 255, B: 255, A: 28}
	case theme.ColorNamePrimary:
		return workspaceWhite
	case theme.ColorNameScrollBar:
		return workspaceColor(0x414145)
	case theme.ColorNameShadow:
		return color.NRGBA{A: 72}
	case theme.ColorNameSeparator:
		return workspaceBorder
	}
	return t.Theme.Color(name, theme.VariantDark)
}

func (t workspaceTheme) Size(name granio.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNameText:
		return 13
	case theme.SizeNameCaptionText:
		return 12
	case theme.SizeNamePadding:
		return 4
	case theme.SizeNameInnerPadding:
		return 4
	case theme.SizeNameLineSpacing:
		return 6
	case theme.SizeNameInlineIcon:
		return 16
	case theme.SizeNameButtonRadius:
		if t.round {
			return 14
		}
		return 9
	case theme.SizeNameScrollBar:
		return 7
	case theme.SizeNameScrollBarSmall:
		return 5
	}
	return t.Theme.Size(name)
}

func (t workspaceTheme) Motion(role granio.MotionRole) granio.MotionSpec {
	spec := theme.Motion(t.Theme, role)
	spec.Curve = func(p float32) float32 {
		left := 1 - p
		return 1 - left*left*left
	}
	switch role {
	case granio.MotionHover:
		spec.Duration = 110 * time.Millisecond
	case granio.MotionPressed:
		spec.Duration = 75 * time.Millisecond
	case granio.MotionNavigation:
		spec.Duration = 210 * time.Millisecond
	}
	if t.appearance != nil {
		switch t.appearance.Motion {
		case 1:
			spec.Duration = spec.Duration * 5 / 9
		case 2:
			spec.Disabled = true
		}
	}
	return spec
}

func (t workspaceTheme) CornerRadii(kind granio.ComponentKind, _ granio.ThemeVariant) (granio.CornerRadii, bool) {
	if t.appearance == nil || t.composer {
		return granio.CornerRadii{}, false
	}
	switch kind {
	case granio.ComponentPopup, granio.ComponentCard:
		return t.appearance.Groups[workspaceCards].corners(), true
	case granio.ComponentInput, granio.ComponentButton, granio.ComponentChoice:
		return t.appearance.Groups[workspaceControls].corners(), true
	}
	return granio.CornerRadii{}, false
}

func (t workspaceTheme) ControlOptics(kind granio.ComponentKind, _ granio.ThemeVariant) (float32, float32, bool) {
	if t.appearance == nil || t.composer || (kind != granio.ComponentButton && kind != granio.ComponentInput) {
		return 0, 0, false
	}
	s := t.appearance.Groups[workspaceControls]
	return float32(s.Transparency) / 100, float32(s.Blur), true
}

var workspaceGlyphs = map[string]string{
	"back":     `<path d="m14 6-6 6 6 6M8 12h12"/>`,
	"next":     `<path d="m10 6 6 6-6 6M16 12H4"/>`,
	"panels":   `<rect x="3" y="5" width="18" height="14" rx="3"/><path d="M10 5v14"/>`,
	"right":    `<rect x="3" y="5" width="18" height="14" rx="3"/><path d="M15 5v14"/>`,
	"home":     `<path fill="#fafafa" d="m3 11 9-8 9 8v8a2 2 0 0 1-2 2h-4v-6H9v6H5a2 2 0 0 1-2-2z"/>`,
	"layers":   `<path d="M8 4 20 6l-2 13-12-2zM5 7 3 20l11 2"/>`,
	"history":  `<circle cx="12" cy="12" r="9"/><path d="M12 6v6l-3 3"/>`,
	"terminal": `<rect x="3" y="4" width="18" height="16" rx="3"/><path d="m7 9 3 3-3 3m5 0h4"/>`,
	"grid":     `<rect x="3" y="3" width="7" height="7" rx="1.5"/><rect x="14" y="3" width="7" height="7" rx="1.5"/><rect x="3" y="14" width="7" height="7" rx="1.5"/><path d="M17.5 14v7m-3.5-3.5h7"/>`,
	"folder":   `<path d="M3 8V6a2 2 0 0 1 2-2h5l2 2h6a2 2 0 0 1 2 2v2M3 8h17a1.5 1.5 0 0 1 1.4 2l-2 7a2 2 0 0 1-1.9 1.5H6a2 2 0 0 1-2-1.6L3 10z"/>`,
	"edit":     `<path d="M11 5H6a2 2 0 0 0-2 2v11a2 2 0 0 0 2 2h11a2 2 0 0 0 2-2v-5M9 15l.5-3.5L17 4a2 2 0 1 1 3 3l-7.5 7.5z"/>`,
	"search":   `<circle cx="10.5" cy="10.5" r="6.5"/><path d="m15.5 15.5 4 4"/>`,
	"bell":     `<path d="M9.6 19a2.5 2.5 0 0 0 4.8 0M18.5 16.5h-13c-.6 0-.9-.7-.5-1.1l1.2-1.5v-3.7a5.8 5.8 0 1 1 11.6 0v3.7l1.2 1.5c.4.4.1 1.1-.5 1.1z"/>`,
	"document": `<path d="M7 3.5h6l4.5 4.6v11a1.4 1.4 0 0 1-1.4 1.4H7a1.5 1.5 0 0 1-1.5-1.5V5A1.5 1.5 0 0 1 7 3.5zM13 3.7v4a.8.8 0 0 0 .8.8h3.5"/>`,
	"chevron":  `<path d="m7 10 5 5 5-5"/>`,
	"overflow": `<path d="M5 12h.2m6.7 0h.2m6.7 0h.2" stroke-width="3"/>`,
	"plus":     `<path d="M12 5v14M5 12h14"/>`,
	"shield":   `<path d="M12 3.5 5 6v5.8c0 4.2 3 7.4 7 8.7 4-1.3 7-4.5 7-8.7V6l-7-2.5Z"/><path d="m9.5 12 1.7 1.8 3.4-3.5"/>`,
	"arrow":    `<path stroke-width="2.25" d="M12 18.5v-13m-5 5 5-5 5 5"/>`,
	"share":    `<path d="M7 11H5v9h14v-9h-2M12 15V3m-4 4 4-4 4 4"/>`,
	"settings": `<path d="M4 7h4m4 0h8M4 17h9m4 0h3"/><circle cx="10" cy="7" r="2"/><circle cx="15" cy="17" r="2"/>`,
	"mic":      `<rect x="9" y="3" width="6" height="11" rx="3"/><path d="M6 10.8v1.8a6 6 0 0 0 12 0v-1.8M12 18.6v2.9"/>`,
	"open":     `<path d="M13 4h7v7m0-7-9 9M10 5H4v15h15v-6"/>`,
	"copy":     `<rect x="8" y="8" width="12" height="12" rx="2"/><path d="M16 8V4H4v12h4"/>`,
	"minimize": `<path d="M5 12h14"/>`,
	"maximize": `<rect x="5" y="5" width="14" height="14" rx="1"/>`,
	"restore":  `<path d="M9 7V4h11v11h-3"/><rect x="4" y="8" width="12" height="12" rx="1"/>`,
	"close":    `<path d="m5 5 14 14M19 5 5 19"/>`,
}

type workspaceIconKey struct {
	name string
	tint color.NRGBA
}

var workspaceIcons sync.Map

func workspaceIcon(name string, tint color.NRGBA) granio.Resource {
	key := workspaceIconKey{name, tint}
	if resource, ok := workspaceIcons.Load(key); ok {
		return resource.(granio.Resource)
	}
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#%02x%02x%02x" stroke-width="1.55" stroke-linecap="round" stroke-linejoin="round">%s</svg>`,
		tint.R, tint.G, tint.B, workspaceGlyphs[name])
	filename := fmt.Sprintf("workspace-%s-%02x%02x%02x.svg", name, tint.R, tint.G, tint.B)
	resource, _ := workspaceIcons.LoadOrStore(key, granio.NewStaticResource(filename, []byte(svg)))
	return resource.(granio.Resource)
}
