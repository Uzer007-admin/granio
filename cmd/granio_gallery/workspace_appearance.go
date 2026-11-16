package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"runtime"
	"time"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/container"
	"github.com/Uzer007-admin/granio/v3/dialog"
	"github.com/Uzer007-admin/granio/v3/driver/desktop"
	"github.com/Uzer007-admin/granio/v3/internal/app"
	"github.com/Uzer007-admin/granio/v3/widget"
)

const workspaceAppearanceKey = "workspace.appearance.v1"

const (
	workspaceEditor = iota
	workspacePanels
	workspaceCards
	workspaceControls
	workspaceGroupCount
)

var workspaceGroupNames = []string{"Редактор", "Боковые панели", "Карточки и меню", "Кнопки и поля"}
var workspacePresets = []string{"Стекло", "Мягкий", "Чёткий"}
var workspaceMotionNames = []string{"Плавно", "Быстро", "Выключено"}
var workspaceBackdropNames = []string{"Непрозрачный", "Прозрачный", "Системное размытие"}

type workspaceOptics struct {
	Transparency, Blur int
	Radii              [4]int
	Linked             bool
}

func (s workspaceOptics) corners() granio.CornerRadii {
	return granio.CornerRadii{TopLeft: float32(s.Radii[0]), TopRight: float32(s.Radii[1]), BottomRight: float32(s.Radii[2]), BottomLeft: float32(s.Radii[3])}
}

type workspaceAppearance struct {
	Groups               [workspaceGroupCount]workspaceOptics
	Motion, Backdrop     int
	BackdropTransparency int
}

func workspacePreset(preset int) workspaceAppearance {
	values := [3][workspaceGroupCount][3]int{
		{{12, 11, 21}, {7, 8, 14}, {9, 10, 12}, {8, 0, 9}},
		{{30, 17, 27}, {20, 12, 20}, {20, 15, 18}, {10, 3, 13}},
		{{0, 0, 6}, {0, 0, 3}, {0, 0, 5}, {0, 0, 4}},
	}
	var appearance workspaceAppearance
	for i, v := range values[preset] {
		appearance.Groups[i] = workspaceOptics{v[0], v[1], [4]int{v[2], v[2], v[2], v[2]}, true}
	}
	appearance.BackdropTransparency = 55
	if preset == 2 {
		appearance.Motion = 1
	}
	return appearance
}

func loadWorkspaceAppearance(preferences granio.Preferences) workspaceAppearance {
	defaults := workspacePreset(0)
	value := defaults
	if json.Unmarshal([]byte(preferences.String(workspaceAppearanceKey)), &value) != nil {
		return defaults
	}
	for i := range value.Groups {
		g := &value.Groups[i]
		g.Transparency = min(max(g.Transparency, 0), 100)
		g.Blur = min(max(g.Blur, 0), 50)
		for j, radius := range g.Radii {
			g.Radii[j] = min(max(radius, 0), 48)
		}
		if g.Linked {
			g.Radii = [4]int{g.Radii[0], g.Radii[0], g.Radii[0], g.Radii[0]}
		}
	}
	value.Motion = min(max(value.Motion, 0), len(workspaceMotionNames)-1)
	value.Backdrop = min(max(value.Backdrop, 0), len(workspaceBackdropNames)-1)
	value.BackdropTransparency = min(max(value.BackdropTransparency, 0), 100)
	return value
}

type workspaceMaterialTarget struct {
	group    int
	material *canvas.Material
	border   *canvas.Rectangle
}

func (v *workspaceView) makeMaterial(group int, tint color.NRGBA, priority bool) workspaceMaterialTarget {
	m := canvas.NewMaterial(tint)
	m.Tint, m.TranslucentFallback, m.Downsample = tint, true, 4
	m.Priority, m.SoftwareBlur = priority, priority
	m.TrackBackdrop = true
	target := workspaceMaterialTarget{group: group, material: m}
	target.apply(v.theme.appearance.Groups[group])
	return target
}

func (v *workspaceView) newMaterial(group int, tint color.NRGBA, border *canvas.Rectangle, priority bool) *canvas.Material {
	target := v.makeMaterial(group, tint, priority)
	target.border = border
	if border != nil {
		radii := v.theme.appearance.Groups[group].corners()
		border.SetCornerRadii(&radii)
	}
	v.surfaces = append(v.surfaces, target)
	return target.material
}

func (t workspaceMaterialTarget) apply(optics workspaceOptics) {
	radii := optics.corners()
	t.material.SetCornerRadii(&radii)
	_ = t.material.ApplyOptics(float32(optics.Transparency)/100, float32(optics.Blur))
	if t.border != nil {
		t.border.SetCornerRadii(&radii)
	}
}

func (v *workspaceView) popupMaterial() *canvas.Material {
	target := v.makeMaterial(workspaceCards, workspaceTop, false)
	target.material.TrackBackdrop = false
	target.material.Live = true
	return target.material
}

func compositionMode(setting int) desktop.WindowBackdrop {
	switch setting {
	case 1:
		return desktop.BackdropTransparent
	case 2:
		if runtime.GOOS == "windows" {
			return desktop.BackdropDesktopAcrylic
		}
		return desktop.BackdropSystemBlur
	}
	return desktop.BackdropOpaque
}

func (v *workspaceView) configureBackdrop() {
	visuals, ok := v.w.(desktop.WindowVisuals)
	if !ok {
		return
	}
	cancel := visuals.ObserveVisualStatus(func(status desktop.VisualStatus) {
		v.visualStatus = status
		if v.appearanceUI != nil {
			v.appearanceUI()
		}
	})
	v.w.SetOnClosed(cancel)
	v.updateBackdrop()
}

func (v *workspaceView) updateBackdrop() {
	appearance := v.theme.appearance
	tint := workspaceBase
	tint.A = uint8(((100-appearance.BackdropTransparency)*255 + 50) / 100)
	_ = desktop.SetWindowBackdropTint(v.w, tint)
	visuals, ok := v.w.(desktop.WindowVisuals)
	if !ok {
		return
	}
	mode := compositionMode(appearance.Backdrop)
	if v.backdropConfigured && v.backdropMode == mode {
		return
	}
	v.backdropConfigured, v.backdropMode = true, mode
	_, v.visualError = visuals.ConfigureVisuals(desktop.VisualOptions{Backdrop: mode, PrepareAlpha: !v.preview})
}

func (v *workspaceView) backdropStatus() string {
	if v.visualError != nil {
		return "Непрозрачный фон. Альфа-буфер станет доступен при следующем запуске."
	}
	visuals, ok := v.w.(desktop.WindowVisuals)
	if !ok {
		return "Непрозрачный фон. Системные эффекты в этом режиме недоступны."
	}
	switch v.visualStatus.Phase {
	case desktop.VisualReady:
		switch v.visualStatus.Selected {
		case desktop.BackdropTransparent:
			return "Прозрачный фон"
		case desktop.BackdropDesktopAcrylic:
			return "Desktop Acrylic. Радиус задаёт композитор."
		case desktop.BackdropSystemBlur:
			return "Системное размытие. Радиус задаёт композитор."
		}
	case desktop.VisualFallback, desktop.VisualFailed:
		reason := v.visualStatus.Reason
		if v.theme.appearance.Backdrop == 2 && (reason == desktop.VisualBlurUnavailable || reason == desktop.VisualUnsupportedOS) && visuals.VisualCapabilities().TransparentSurface {
			return "Системное размытие недоступно. Можно выбрать прозрачный фон."
		}
		return "Непрозрачный фон. Системные эффекты в этом режиме недоступны."
	case desktop.VisualPending:
		return "Подготовка фона окна…"
	}
	return "Непрозрачный фон"
}

func (v *workspaceView) applyAppearance() {
	groups := v.theme.appearance.Groups
	if v.materialStyles == groups {
		return
	}
	for _, surface := range v.surfaces {
		if g := surface.group; v.materialStyles[g] != groups[g] {
			surface.apply(groups[g])
		}
	}
	if v.materialStyles[workspaceCards] != groups[workspaceCards] {
		for _, surface := range v.messageSurfaces {
			surface.apply(groups[workspaceCards])
		}
		if len(v.messageSurfaces) != 0 {
			v.composerGlass.InvalidateBackdrop()
		}
	}
	v.materialStyles = groups
}

func (v *workspaceView) scheduleAppearance() {
	if v.appearanceEdit != nil {
		return
	}
	a := granio.NewAnimation(time.Nanosecond, func(float32) {})
	a.Object = v.root
	a.OnFinished = func() {
		if v.appearanceEdit == a {
			v.appearanceEdit = nil
			v.flushAppearance()
		}
	}
	v.appearanceEdit = a
	a.Start()
}

func (v *workspaceView) flushAppearance() {
	if a := v.appearanceEdit; a != nil {
		v.appearanceEdit = nil
		a.Stop()
	}
	v.applyAppearance()
	v.updateBackdrop()
	if group := v.theme.appearance.Groups[workspaceControls]; v.controlStyle != group {
		v.controlStyle = group
		app.ApplyThemeTo(v.root, v.w.Canvas())
		for _, overlay := range v.w.Canvas().Overlays().List() {
			app.ApplyThemeTo(overlay, v.w.Canvas())
		}
	}
	if v.appearanceUI != nil {
		v.appearanceUI()
	}
}

func (v *workspaceView) saveAppearance() {
	v.flushAppearance()
	if !v.preview {
		data, err := json.Marshal(v.theme.appearance)
		if err == nil {
			granio.CurrentApp().Preferences().SetString(workspaceAppearanceKey, string(data))
		}
	}
}

func (v *workspaceView) showAppearance() {
	if v.appearanceUI != nil {
		return
	}
	appearance := v.theme.appearance
	group, binding := workspaceEditor, false
	preview := v.makeMaterial(group, workspaceColor(0x383838), true)
	status := widget.NewLabel("")
	windowStatus := widget.NewLabel("")
	windowStatus.Wrapping = granio.TextWrapWord
	const (
		previewWidth  = 480
		previewHeight = 92
	)
	colors := [...]color.NRGBA{workspaceColor(0x154760), workspaceColor(0xb1d8c0), workspaceColor(0x294456), workspaceColor(0xb78365)}
	stripes := make([]*canvas.Rectangle, 16)
	objects := make([]granio.CanvasObject, 0, len(stripes)+1)
	for i := range stripes {
		stripes[i] = canvas.NewRectangle(colors[i%len(colors)])
		objects = append(objects, stripes[i])
	}
	objects = append(objects, preview.material)
	sample := workspaceArrange(granio.NewSize(previewWidth, previewHeight), func(s granio.Size) {
		step := s.Width / float32(len(stripes))
		for i, stripe := range stripes {
			workspacePlace(stripe, float32(i)*step, 0, step, s.Height)
		}
		workspacePlace(preview.material, 70, 13, s.Width-140, s.Height-26)
	}, objects...)
	effectText := func(effect canvas.MaterialEffect) {
		text := [...]string{"Подготовка образца…", "Полупрозрачная заливка", "Непрозрачная заливка", "Аппаратное размытие", "Размытие в программном режиме"}[effect]
		if effect == canvas.MaterialTint && appearance.Groups[group].Transparency == 100 {
			text = "Прозрачная область"
		}
		if status.Text != text {
			status.SetText(text)
		}
	}
	preview.material.SetOnEffectChanged(effectText)
	setLabel := func(w *widget.Label, text string) {
		if w.Text != text {
			w.SetText(text)
		}
	}
	newSlider := func(name, unit string, limit int, update func(int)) (*widget.Slider, *widget.Label, granio.CanvasObject) {
		slider := widget.NewSlider(0, float64(limit))
		label := widget.NewLabel(name)
		value := widget.NewLabel("")
		value.Alignment = granio.TextAlignTrailing
		slider.OnChanged = func(input float64) {
			number := int(input + .5)
			value.SetText(fmt.Sprintf("%d%s", number, unit))
			if !binding {
				update(number)
				v.scheduleAppearance()
			}
		}
		slider.OnChangeEnded = func(float64) {
			if !binding {
				v.saveAppearance()
			}
		}
		slider.OnChanged(0)
		row := container.NewBorder(nil, nil, label, container.NewGridWrap(granio.NewSize(55, 31), value), slider)
		return slider, label, row
	}
	binding = true
	transparency, _, transparencyRow := newSlider("Прозрачность", "%", 100, func(value int) { appearance.Groups[group].Transparency = value })
	radius, _, radiusRow := newSlider("Радиус размытия", " dp", 50, func(value int) { appearance.Groups[group].Blur = value })
	var cornerSliders [4]*widget.Slider
	var firstCornerLabel *widget.Label
	var cornerRows []granio.CanvasObject
	for i, name := range []string{"Верхний левый", "Верхний правый", "Нижний правый", "Нижний левый"} {
		slider, label, row := newSlider(name, " dp", 48, func(value int) {
			g := &appearance.Groups[group]
			g.Radii[i] = value
			if g.Linked {
				g.Radii = [4]int{value, value, value, value}
			}
		})
		if i == 0 {
			firstCornerLabel = label
		}
		cornerSliders[i] = slider
		cornerRows = append(cornerRows, row)
	}
	linked := widget.NewCheck("Связанные углы", func(value bool) {
		if binding {
			return
		}
		g := &appearance.Groups[group]
		g.Linked = value
		if value {
			g.Radii = [4]int{g.Radii[0], g.Radii[0], g.Radii[0], g.Radii[0]}
		}
		v.saveAppearance()
	})
	groupSelect := widget.NewSelect(workspaceGroupNames, nil)
	groupSelect.OnChanged = func(string) {
		if index := groupSelect.SelectedIndex(); index >= 0 {
			group = index
			if v.appearanceUI != nil {
				v.appearanceUI()
			}
		}
	}
	groupSelect.SetSelectedIndex(group)
	motion := widget.NewSelect(workspaceMotionNames, nil)
	motion.OnChanged = func(string) {
		if !binding {
			appearance.Motion = motion.SelectedIndex()
			v.saveAppearance()
		}
	}
	stylesheet := widget.NewSelect(workspacePresets, nil)
	stylesheet.PlaceHolder = "Пресеты"
	stylesheet.OnChanged = func(string) {
		if index := stylesheet.SelectedIndex(); index >= 0 && !binding {
			*appearance = workspacePreset(index)
			v.saveAppearance()
		}
	}
	resetGroup := widget.NewButton("Сбросить группу", func() {
		appearance.Groups[group] = workspacePreset(0).Groups[group]
		v.saveAppearance()
	})
	reset := widget.NewButton("Сбросить всё", func() {
		*appearance = workspacePreset(0)
		stylesheet.ClearSelected()
		v.saveAppearance()
	})
	backdrops := workspaceBackdropNames
	if runtime.GOOS == "windows" {
		backdrops = []string{"Непрозрачный", "Прозрачный", "Desktop Acrylic"}
	}
	backdrop := widget.NewSelect(backdrops, nil)
	backdrop.OnChanged = func(string) {
		if !binding {
			appearance.Backdrop = backdrop.SelectedIndex()
			v.saveAppearance()
		}
	}
	behindAlpha, _, behindRow := newSlider("Прозрачность подложки", "%", 100, func(value int) { appearance.BackdropTransparency = value })
	v.appearanceUI = func() {
		binding = true
		defer func() { binding = false }()
		g := appearance.Groups[group]
		preview.apply(g)
		effectText(preview.material.Effect())
		transparency.SetValue(float64(g.Transparency))
		radius.SetValue(float64(g.Blur))
		linked.SetChecked(g.Linked)
		if g.Linked {
			setLabel(firstCornerLabel, "Все углы")
		} else {
			setLabel(firstCornerLabel, "Верхний левый")
		}
		for i, slider := range cornerSliders {
			slider.SetValue(float64(g.Radii[i]))
			workspaceShow(cornerRows[i], i == 0 || !g.Linked)
		}
		motion.SetSelectedIndex(appearance.Motion)
		backdrop.SetSelectedIndex(appearance.Backdrop)
		behindAlpha.SetValue(float64(appearance.BackdropTransparency))
		adjustable := v.visualStatus.Phase == desktop.VisualReady && v.visualStatus.Selected != desktop.BackdropOpaque
		if adjustable && behindAlpha.Disabled() {
			behindAlpha.Enable()
		} else if !adjustable && !behindAlpha.Disabled() {
			behindAlpha.Disable()
		}
		setLabel(windowStatus, v.backdropStatus())
	}
	controls := container.NewVBox(
		container.NewGridWithColumns(2, groupSelect, stylesheet),
		transparencyRow, radiusRow, linked, container.NewVBox(cornerRows...),
		container.NewGridWithColumns(2, widget.NewLabel("Анимации"), motion),
		container.NewGridWithColumns(2, resetGroup, reset),
	)
	window := container.NewVBox(backdrop, behindRow, windowStatus)
	addTab := func(name string, content granio.CanvasObject, footer granio.CanvasObject) *container.TabItem {
		scroll := container.NewVScroll(content)
		scroll.SetMinSize(granio.NewSize(previewWidth, 265))
		return container.NewTabItem(name, container.NewBorder(nil, footer, nil, nil, scroll))
	}
	tabs := container.NewAppTabs(addTab("Элементы", controls, container.NewVBox(sample, status)), addTab("Фон за окном", window, nil))
	// The preview is the only software blur while its settings are visible.
	v.composerGlass.SoftwareBlur = false
	v.composerGlass.Priority = false
	tabs.OnSelected = func(tab *container.TabItem) {
		sampleVisible := tab == tabs.Items[0]
		v.composerGlass.SoftwareBlur, v.composerGlass.Priority = !sampleVisible, !sampleVisible
	}
	v.appearanceUI()
	d := dialog.NewCustom("Оформление", "Закрыть", tabs, v.w)
	d.SetOnClosed(func() {
		preview.material.SetOnEffectChanged(nil)
		v.appearanceUI = nil
		v.composerGlass.SoftwareBlur, v.composerGlass.Priority = true, true
		v.saveAppearance()
	})
	d.Show()
}
