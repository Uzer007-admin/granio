package main

import (
	"fmt"
	"image/color"
	"strings"
	"unicode/utf8"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/container"
	"github.com/Uzer007-admin/granio/v3/dialog"
	"github.com/Uzer007-admin/granio/v3/theme"
	"github.com/Uzer007-admin/granio/v3/widget"
)

const (
	workspaceChatLimit     = 16000
	workspaceComposeMin    = 78
	workspaceComposeMax    = 226
	workspaceComposeTools  = 48
	workspaceComposeBottom = 17
	workspaceUsageSize     = 34
	workspaceUsageGap      = 12
)

type workspaceThread struct {
	draft    string
	messages []string
	used     int
}

func (v *workspaceView) composer() granio.CanvasObject {
	v.entry = widget.NewMultiLineEntry()
	v.entry.Wrapping = granio.TextWrapWord
	v.entry.SubmitOnEnter = true
	v.entry.Borderless = true
	v.entry.SetMinRowsVisible(1)
	v.entry.SetPlaceHolder("Напишите сообщение")
	v.entry.OnChanged = v.draftChanged
	v.entry.OnSubmitted = func(string) { v.addNote() }
	plain := v.theme
	plain.composer = true
	entry := container.NewThemeOverride(v.entry, plain)
	edge := workspaceSurface(color.Transparent, 21)
	edge.StrokeColor, edge.StrokeWidth = color.NRGBA{R: 255, G: 255, B: 255, A: 10}, 1
	glass := v.newMaterial(workspaceEditor, workspaceColor(0x383838), edge, true)
	glass.TrackBackdrop = true
	v.composerGlass = glass
	var plus *widget.Button
	plus = workspaceButton("", "Материалы", "plus", workspaceText, func() {
		v.menu(plus, granio.NewMenuItem("Показать документы", func() { v.setDocuments(true) }))
	})
	delivery := workspaceButton("Режим отправки", "Режим отправки", "shield", workspaceMuted, nil)
	delivery.Alignment = widget.ButtonAlignLeading
	delivery.OnTapped = func() {
		selectConfirmation := func(value bool) {
			v.confirmSend = value
			delivery.SetSelected(value)
		}
		v.menu(delivery,
			granio.NewMenuItem("Отправлять сразу", func() { selectConfirmation(false) }),
			granio.NewMenuItem("Подтверждать отправку", func() { selectConfirmation(true) }))
	}
	v.rendererButton = workspaceButton("Система", "Режим графики", "chevron", workspaceMuted, v.rendererMenu)
	v.rendererButton.IconPlacement = widget.ButtonIconTrailingText
	focus := workspaceButton("", "Фокус на поле ввода", "mic", workspaceText, v.focusComposer)
	v.send = workspaceButton("", "Отправить сообщение", "arrow", workspaceBase, v.addNote)
	v.send.Importance = widget.HighImportance
	round := v.theme
	round.round = true
	sendControl := container.NewThemeOverride(v.send, round)
	v.usage = newWorkspaceUsage(v.w)
	v.usage.material = v.popupMaterial
	v.draftChanged(v.entry.Text)
	return workspaceArrange(granio.NewSize(260, workspaceComposeMin), func(s granio.Size) {
		workspacePlace(glass, 0, 0, s.Width, s.Height)
		workspacePlace(edge, 0, 0, s.Width, s.Height)
		workspacePlace(entry, 8, 9, s.Width-16, s.Height-workspaceComposeTools)
		workspacePlace(plus, 7, s.Height-36, 29, 29)
		deliveryWidth := min(164, max(0, s.Width-213))
		caption := "Режим отправки"
		if deliveryWidth < 122 {
			caption = ""
		}
		if caption != delivery.Text {
			delivery.SetText(caption)
		}
		workspacePlace(delivery, 37, s.Height-36, deliveryWidth, 29)
		workspacePlace(v.rendererButton, s.Width-166, s.Height-36, 88, 29)
		workspacePlace(focus, s.Width-72, s.Height-36, 28, 29)
		workspacePlace(sendControl, s.Width-36, s.Height-35, 28, 28)
	}, glass, edge, entry, plus, delivery, v.rendererButton, focus, sendControl)
}

func (v *workspaceView) draftChanged(text string) {
	count := utf8.RuneCountInString(strings.TrimSpace(text))
	used := v.activeUsage() + count
	v.usage.update(used)
	canSend := count != 0 && used <= workspaceChatLimit
	if canSend && v.send.Disabled() {
		v.send.Enable()
	} else if !canSend && !v.send.Disabled() {
		v.send.Disable()
	}
	if v.conversationView != nil {
		v.setComposerHeight(v.desiredComposerHeight(), true)
	}
}

func (v *workspaceView) activeUsage() int {
	if v.project < 0 {
		return 0
	}
	return v.threads[v.project].used
}

func (v *workspaceView) desiredComposerHeight() float32 {
	return min(v.composeMaxHeight, max(workspaceComposeMin, v.entry.TextContentHeight()+workspaceComposeTools))
}

func (v *workspaceView) setComposerHeight(height float32, animated bool) {
	if animated && height == v.composeTarget {
		return
	}
	if v.composeAnimation != nil {
		v.composeAnimation.Stop()
		v.composeAnimation = nil
	}
	v.composeTarget = height
	if !animated || height == v.composeHeight {
		v.composeHeight = height
		return
	}
	from := v.composeHeight
	a := workspaceAnimation(v.conversationView, granio.MotionNavigation, func(p float32) {
		v.composeHeight = from + (height-from)*p
		v.relayoutConversation()
	})
	a.OnFinished = func() {
		if v.composeAnimation == a {
			v.composeAnimation = nil
		}
	}
	v.composeAnimation = a
	a.Start()
}

func (v *workspaceView) relayoutConversation() {
	c := v.conversationView
	c.Layout.Layout(c.Objects, c.Size())
}

func (v *workspaceView) addNote() {
	if v.preview || v.entry.CompositionActive() {
		return
	}
	text := strings.TrimSpace(v.entry.Text)
	count := utf8.RuneCountInString(text)
	if count == 0 || v.activeUsage()+count > workspaceChatLimit {
		return
	}
	if v.confirmSend {
		project, original := v.project, v.entry.Text
		dialog.ShowConfirm("Отправить сообщение", "Сообщение появится в текущей переписке.", func(confirmed bool) {
			if confirmed && v.project == project && v.entry.Text == original && v.activeUsage()+count <= workspaceChatLimit {
				v.sendNote(text, count)
			}
		}, v.w)
		return
	}
	v.sendNote(text, count)
}

func (v *workspaceView) sendNote(text string, count int) {
	if v.project < 0 {
		v.newConversation()
	}
	thread := &v.threads[v.project]
	thread.messages = append(thread.messages, text)
	thread.used += count
	thread.draft = ""
	message, slide := v.workspaceMessage(text)
	v.feed.add(message)
	v.composerGlass.InvalidateBackdrop()
	if v.entry.Text == "" {
		v.draftChanged("")
	} else {
		v.entry.SetText("")
	}
	v.relayoutConversation()
	v.feedScroll.RefreshContentSize()
	v.feedScroll.ScrollToBottom()
	v.focusComposer()
	v.stopMessageMotion()
	a := workspaceAnimation(message, granio.MotionSurface, func(p float32) {
		slide.Move(granio.NewPos(0, (1-p)*7))
		v.composerGlass.InvalidateBackdrop()
	})
	a.OnFinished = func() {
		if v.messageAnimation == a {
			v.messageAnimation, v.movingMessage = nil, nil
		}
	}
	v.messageAnimation, v.movingMessage = a, slide
	a.Start()
}

func (v *workspaceView) stopMessageMotion() {
	if v.messageAnimation != nil {
		v.messageAnimation.Stop()
		v.movingMessage.Move(granio.Position{})
		v.messageAnimation, v.movingMessage = nil, nil
		v.composerGlass.InvalidateBackdrop()
	}
}

func (v *workspaceView) workspaceMessage(text string) (*workspaceBlock, *granio.Container) {
	paragraph := workspaceParagraph(text)
	target := v.makeMaterial(workspaceCards, workspaceRaised, false)
	v.messageSurfaces = append(v.messageSurfaces, target)
	bg := target.material
	content := workspaceArrange(granio.Size{}, func(s granio.Size) {
		workspacePlace(bg, 0, 0, s.Width, s.Height)
		workspacePlace(paragraph, 8, 5, s.Width-16, s.Height-10)
	}, bg, paragraph)
	clip := container.NewClip(content)
	return newWorkspaceBlock(func(width float32) float32 {
		bubbleWidth := width * 0.81
		height := workspaceMeasureHeight(paragraph, max(0, bubbleWidth-16)) + 10
		workspacePlace(clip, width-bubbleWidth, 0, bubbleWidth, height)
		return height
	}, clip), content
}

func workspaceAnimation(o granio.CanvasObject, role granio.MotionRole, update func(float32)) *granio.Animation {
	spec := theme.Motion(theme.CurrentForWidget(o), role)
	a := granio.NewAnimation(spec.Duration, update)
	a.Object, a.Curve, a.Disabled = o, spec.Curve, spec.Disabled
	return a
}

type workspaceUsage struct {
	window   granio.Window
	control  *granio.Container
	button   *widget.Button
	arc      *canvas.Arc
	used     int
	material func() *canvas.Material
}

func newWorkspaceUsage(w granio.Window) *workspaceUsage {
	u := &workspaceUsage{window: w, used: -1}
	u.button = workspaceButton("", "Лимит переписки", "overflow", workspaceText, u.show)
	disc := workspaceSurface(color.NRGBA{R: 29, G: 29, B: 29, A: 243}, workspaceUsageSize/2)
	disc.StrokeColor, disc.StrokeWidth = workspaceColor(0x343434), 1
	u.arc = canvas.NewArc(0, 0, 0.91, workspaceMuted)
	u.control = workspaceArrange(granio.Size{}, func(s granio.Size) {
		workspacePlace(disc, 0, 0, s.Width, s.Height)
		workspacePlace(u.arc, 2, 2, s.Width-4, s.Height-4)
		workspacePlace(u.button, 0, 0, s.Width, s.Height)
	}, disc, u.arc, u.button)
	return u
}

func (u *workspaceUsage) update(used int) {
	if u.used == used {
		return
	}
	u.used = used
	angle := float32(min(used, workspaceChatLimit) * 360 / workspaceChatLimit)
	tint := workspaceMuted
	if used > workspaceChatLimit {
		tint = workspaceColor(0xf58e89)
	}
	if u.arc.EndAngle != angle || u.arc.FillColor != tint {
		u.arc.EndAngle, u.arc.FillColor = angle, tint
		u.arc.Refresh()
	}
	workspaceShow(u.arc, angle > 0)
}

func (u *workspaceUsage) show() {
	title := widget.NewLabelWithStyle("Лимит переписки", granio.TextAlignLeading, granio.TextStyle{Bold: true})
	count := widget.NewLabel(fmt.Sprintf("%d из %d символов", u.used, workspaceChatLimit))
	progress := widget.NewProgressBar()
	progress.TextFormatter = func() string { return fmt.Sprintf("%.0f%%", 100*progress.Value) }
	progress.SetValue(float64(min(u.used, workspaceChatLimit)) / workspaceChatLimit)
	remaining := fmt.Sprintf("Осталось %d символов", workspaceChatLimit-u.used)
	if u.used > workspaceChatLimit {
		remaining = fmt.Sprintf("Уберите %d символов из черновика", u.used-workspaceChatLimit)
	}
	panel := container.NewPadded(container.NewVBox(title, count, progress, widget.NewLabel(remaining)))
	popup := widget.NewPopUp(panel, u.window.Canvas())
	if u.material != nil {
		popup.Material = u.material()
	}
	pos := granio.CurrentApp().Driver().AbsolutePositionForObject(u.button)
	minimum := panel.MinSize()
	size := u.window.Canvas().Size()
	left := pos.X + (u.button.Size().Width-minimum.Width)/2
	popup.ShowAtPosition(granio.NewPos(max(0, min(left, size.Width-minimum.Width-12)), max(0, pos.Y-minimum.Height-12)))
}
