package main

import (
	"fmt"
	"image/color"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/container"
	"github.com/Uzer007-admin/granio/v3/dialog"
	"github.com/Uzer007-admin/granio/v3/driver/desktop"
	"github.com/Uzer007-admin/granio/v3/widget"
)

func (v *workspaceView) selectProject(index int) {
	if index < 0 || index >= len(v.threads) || index == v.project {
		return
	}
	if v.project >= 0 {
		v.threads[v.project].draft = v.entry.Text
	}
	v.stopMessageMotion()
	v.project = index
	v.title.SetText(workspaceThreadName(index))
	v.history.Select(index)
	v.feed.replace(v.conversationItems())
	v.composerGlass.InvalidateBackdrop()
	draft := v.threads[index].draft
	if draft != v.entry.Text {
		v.entry.SetText(draft)
	} else {
		v.draftChanged(draft)
	}
	v.relayoutConversation()
	v.feedScroll.RefreshContentSize()
	if len(v.threads[index].messages) > 0 {
		v.feedScroll.ScrollToBottom()
	} else {
		v.feedScroll.ScrollToTop()
	}
}

func workspaceThreadName(index int) string { return fmt.Sprintf("Диалог %d", index+1) }

func (v *workspaceView) newConversation() {
	if v.preview {
		return
	}
	v.threads = append(v.threads, workspaceThread{})
	v.history.Refresh()
	v.selectProject(len(v.threads) - 1)
	v.history.ScrollTo(len(v.threads) - 1)
	v.focusComposer()
}

func (v *workspaceView) sidebar() granio.CanvasObject {
	title := workspaceCanvasText("Granio", workspaceText, 17, true)
	var filter *widget.Button
	filter = workspaceButton("", "Навигация по панелям", "chevron", workspaceMuted, func() {
		v.menu(filter, granio.NewMenuItem("Скрыть навигацию", func() { v.setSidebar(false) }),
			granio.NewMenuItem("Открыть материалы", func() { v.setDocuments(true) }))
	})
	bell := workspaceButton("", "Журнал", "bell", workspaceSubtle, func() { v.selectDocument(0); v.setDocuments(true) })
	settings := workspaceButton("", "Оформление", "settings", workspaceSubtle, v.showAppearance)
	v.newDialog = workspaceButton("Новый диалог", "", "edit", workspaceText, v.newConversation)
	v.newDialog.Alignment = widget.ButtonAlignLeading
	v.history = widget.NewList(func() int { return len(v.threads) }, func() granio.CanvasObject {
		label := widget.NewLabel("")
		label.Truncation = granio.TextTruncateEllipsis
		return label
	}, func(id widget.ListItemID, object granio.CanvasObject) {
		object.(*widget.Label).SetText(workspaceThreadName(id))
	})
	v.history.HideSeparators = true
	v.history.OnSelected = v.selectProject
	bg := v.newMaterial(workspacePanels, workspaceSidebar, nil, false)
	divider, rule := canvas.NewRectangle(workspaceBorder), canvas.NewRectangle(workspaceBorder)
	grip := newWorkspaceGrip(func(delta float32) {
		v.sidebarPane.SetPanelWidth(max(172, min(370, v.sidebarPane.PanelWidth()+delta)))
	})
	return workspaceArrange(granio.NewSize(0, 240), func(s granio.Size) {
		workspacePlace(bg, 0, 0, s.Width, s.Height)
		workspacePlace(title, 17, 13, 67, 27)
		workspacePlace(filter, 66, 13, 31, 28)
		workspacePlace(bell, s.Width-74, 11, 29, 30)
		workspacePlace(settings, s.Width-42, 11, 29, 30)
		workspacePlace(v.newDialog, 12, 52, s.Width-24, 34)
		workspacePlace(rule, 0, 90, s.Width, 1)
		workspacePlace(v.history, 5, 93, s.Width-10, s.Height-100)
		workspacePlace(divider, s.Width-1, 0, 1, s.Height)
		workspacePlace(grip, s.Width-5, workspaceHeaderHeight, 5, s.Height-workspaceHeaderHeight)
	}, bg, title, filter, bell, settings, v.newDialog, rule, v.history, divider, grip)
}

func (v *workspaceView) conversation() granio.CanvasObject {
	header := v.centerHeader()
	editor := v.composer()
	v.feed = newWorkspaceFlow(21, 10)
	v.feedScroll = container.NewVScroll(v.feed)
	v.feedScroll.OnScrolled = func(granio.Position) { v.composerGlass.InvalidateBackdrop() }
	v.conversationView = workspaceArrange(granio.NewSize(320, 365), func(s granio.Size) {
		readingWidth := min(770, s.Width)
		readingLeft := (s.Width - readingWidth) / 2
		editorWidth := readingWidth - 34
		limit := min(workspaceComposeMax, max(workspaceComposeMin, (s.Height-workspaceHeaderHeight-100)*0.6))
		if width := editorWidth - 16; v.entry.Size().Width != width || v.composeMaxHeight != limit {
			v.entry.Resize(granio.NewSize(width, v.entry.Size().Height))
			v.composeMaxHeight = limit
			v.setComposerHeight(v.desiredComposerHeight(), false)
		}
		editorBottom := s.Height - workspaceComposeBottom
		editorY := editorBottom - v.composeHeight
		usageY := editorY - workspaceUsageGap - workspaceUsageSize
		workspacePlace(header, 0, 0, s.Width, workspaceHeaderHeight)
		oldViewport := v.feedScroll.Size()
		follow := v.feedScroll.Offset.Y > 0 && v.feedScroll.Offset.Y+oldViewport.Height >= v.feed.Size().Height-1
		// Let content scroll under the glass and leave enough room to read the
		// last message above both controls when the scroll reaches the end.
		inset := editorBottom - usageY + 4
		insetChanged := v.feed.bottomInset != inset
		v.feed.bottomInset = inset
		workspacePlace(v.feedScroll, readingLeft, workspaceHeaderHeight, readingWidth, editorBottom-workspaceHeaderHeight)
		sizeChanged := oldViewport != v.feedScroll.Size()
		if insetChanged && !sizeChanged {
			v.feedScroll.RefreshContentSize()
		}
		if follow && (insetChanged || sizeChanged) {
			v.feedScroll.ScrollToBottom()
		}
		workspacePlace(v.usage.control, (s.Width-workspaceUsageSize)/2, usageY, workspaceUsageSize, workspaceUsageSize)
		workspacePlace(editor, readingLeft+17, editorY, editorWidth, v.composeHeight)
	}, header, v.feedScroll, v.usage.control, editor)
	return v.conversationView
}

func (v *workspaceView) conversationItems() []granio.CanvasObject {
	messages := v.threads[v.project].messages
	items := make([]granio.CanvasObject, 0, len(messages))
	clear(v.messageSurfaces)
	v.messageSurfaces = v.messageSurfaces[:0]
	for _, text := range messages {
		message, _ := v.workspaceMessage(text)
		items = append(items, message)
	}
	return items
}

func (v *workspaceView) documents() granio.CanvasObject {
	bg := v.newMaterial(workspacePanels, workspaceBase, nil, false)
	tabs := v.documentTabs()
	add := workspaceButton("", "Открыть README", "plus", workspaceMuted, func() { v.selectDocument(2) })
	toggle := workspaceButton("", "Свернуть документы", "right", workspaceText, func() { v.setDocuments(false) })
	v.docName = workspaceCanvasText("", workspaceText, 13, true)
	v.docName.Alignment = granio.TextAlignCenter
	fileChip := workspaceSurface(workspaceColor(0x2b2b2b), 16)
	copyIcon := workspaceButton("", "Показать отчёт", "copy", workspaceText, func() { v.selectDocument(1) })
	open := workspaceButton("Открыть", "", "open", workspaceText, func() {
		preview := dialog.NewCustom(workspaceDocNames[v.doc], "Закрыть", container.NewVScroll(workspaceDocument(v.doc)), v.w)
		preview.Resize(granio.NewSize(min(540, v.w.Canvas().Size().Width-48), min(575, v.w.Canvas().Size().Height-48)))
		preview.Show()
	})
	var more *widget.Button
	more = workspaceButton("", "Действия с документом", "overflow", workspaceText, func() {
		v.menu(more, granio.NewMenuItem("К началу", v.docScroll.ScrollToTop),
			granio.NewMenuItem("Перенести панель на другую сторону", v.moveDocuments),
			granio.NewMenuItem("Свернуть документы", func() { v.setDocuments(false) }))
	})
	round := v.theme
	round.round = true
	copyIcon.Selected, more.Selected, open.Selected = true, true, true
	copyControl := container.NewThemeOverride(copyIcon, round)
	moreControl := container.NewThemeOverride(more, round)
	openControl := container.NewThemeOverride(open, round)
	v.docScroll = container.NewVScroll(workspaceDocument(0))
	firstRule, secondRule, divider := canvas.NewRectangle(workspaceBorder), canvas.NewRectangle(workspaceBorder), canvas.NewRectangle(workspaceBorder)
	grip := newWorkspaceGrip(func(delta float32) {
		if v.documentPane.Side() == container.PanelLeading {
			delta = -delta
		}
		v.documentPane.SetPanelWidth(max(210, min(610, v.documentPane.PanelWidth()-delta)))
	})
	v.documentContent = workspaceArrange(granio.NewSize(0, 240), func(s granio.Size) {
		workspacePlace(bg, 0, 0, s.Width, s.Height)
		workspacePlace(tabs, 9, 10, max(0, s.Width-83), 32)
		workspacePlace(add, s.Width-70, 10, 28, 32)
		workspacePlace(toggle, s.Width-38, 10, 29, 32)
		workspacePlace(firstRule, 0, 51, s.Width, 1)
		workspacePlace(fileChip, 9, 60, 70, 34)
		workspacePlace(v.docName, 9, 66, 70, 23)
		openWidth := float32(34)
		caption := ""
		if s.Width >= 270 {
			openWidth, caption = 104, "Открыть"
		}
		if open.Text != caption {
			open.SetText(caption)
		}
		openX := s.Width - 9 - openWidth
		workspacePlace(openControl, openX, 59, openWidth, 34)
		workspacePlace(moreControl, openX-39, 59, 34, 34)
		workspacePlace(copyControl, openX-78, 59, 34, 34)
		workspaceShow(moreControl, openX-39 >= 84)
		workspaceShow(copyControl, openX-78 >= 84)
		workspacePlace(secondRule, 0, 100, s.Width, 1)
		workspacePlace(v.docScroll, 7, 101, s.Width-10, s.Height-103)
		edge, handle := float32(0), float32(0)
		if v.documentPane != nil && v.documentPane.Side() == container.PanelLeading {
			edge, handle = s.Width-1, s.Width-5
		}
		workspacePlace(divider, edge, 0, 1, s.Height)
		workspacePlace(grip, handle, workspaceHeaderHeight, 5, s.Height-workspaceHeaderHeight)
	}, bg, tabs, add, toggle, fileChip, v.docName, copyControl, moreControl, openControl, v.docScroll, firstRule, secondRule, divider, grip)
	return v.documentContent
}

func (v *workspaceView) moveDocuments() {
	side := container.PanelLeading
	if v.documentPane.Side() == container.PanelLeading {
		side = container.PanelTrailing
	}
	v.documentPane.SetSide(side)
	v.documentContent.Layout.Layout(v.documentContent.Objects, v.documentContent.Size())
}

func (v *workspaceView) selectDocument(index int) {
	if index < 0 || index >= len(workspaceDocNames) {
		return
	}
	if v.doc != index {
		v.doc = index
		v.docScroll.Content = workspaceDocument(index)
		v.docScroll.Refresh()
		v.docScroll.ScrollToTop()
		v.moveTabHighlight(index)
	}
	v.docName.Text = workspaceDocNames[index]
	v.docName.Refresh()
	for i, b := range v.docTabs {
		b.SetSelected(i == index)
	}
}

func workspaceDocument(index int) *workspaceFlow {
	switch index {
	case 1:
		return newWorkspaceFlow(12, 12,
			workspaceParagraph("Отчёт о компонентах"),
			workspaceDocumentRow("Навигация", "Панель с историей можно свернуть и снова открыть. Текущая заметка сохраняет выделение."),
			workspaceDocumentRow("Документы", "Перетащите вкладку, чтобы изменить порядок. Ширина панели задаётся разделителем."),
			workspaceDocumentRow("Клавиатура", "Enter отправляет сообщение. Shift+Enter добавляет новую строку."))
	case 2:
		return newWorkspaceFlow(14, 12,
			workspaceParagraph("Рабочая область Granio"),
			workspaceParagraph("В панели навигации можно открыть отдельную переписку. При переключении сохраняются сообщения и черновики."),
			workspaceParagraph("Потяните разделители, чтобы изменить ширину панелей. Поле ввода растёт до предельной высоты, затем прокручивается внутри."))
	default:
		return newWorkspaceFlow(10, 10,
			workspaceDocumentRow("Сборка", "Первый кадр отображает панели, шрифты и контурные иконки в тёмной теме."),
			workspaceDocumentRow("Масштабирование", "При изменении размеров окна заголовок и поле ввода остаются на месте. Боковые панели сдвигаются, а внутри сохраняются отступы и высота строк. Вкладки остаются доступны после восстановления размера."),
			workspaceParagraph("Обновления размера применяются по кадрам. Контрольный снимок включает полную рабочую область, навигацию и открытые вкладки. Для повторной проверки доступны журнал и короткий отчёт."),
			workspaceParagraph("Проверки ввода\nФокус возвращается в черновик. Кнопки окна поддерживают сворачивание, разворачивание и восстановление."),
			workspaceParagraph("Система учитывает настройку сокращённого движения. Прокрутка доступна для журнала и списка заметок."))
	}
}

type workspaceGrip struct {
	widget.BaseWidget
	line   *canvas.Rectangle
	resize func(float32)
}

func newWorkspaceGrip(resize func(float32)) *workspaceGrip {
	g := &workspaceGrip{line: canvas.NewRectangle(color.Transparent), resize: resize}
	g.ExtendBaseWidget(g)
	return g
}

func (g *workspaceGrip) CreateRenderer() granio.WidgetRenderer {
	return widget.NewSimpleRenderer(g.line)
}
func (g *workspaceGrip) Cursor() desktop.Cursor       { return desktop.HResizeCursor }
func (g *workspaceGrip) Dragged(e *granio.DragEvent)  { g.resize(e.Dragged.DX) }
func (*workspaceGrip) DragEnd()                       {}
func (*workspaceGrip) MouseMoved(*desktop.MouseEvent) {}
func (g *workspaceGrip) MouseIn(*desktop.MouseEvent) {
	g.line.FillColor = color.NRGBA{R: 154, G: 166, B: 198, A: 65}
	g.line.Refresh()
}
func (g *workspaceGrip) MouseOut() {
	g.line.FillColor = color.Transparent
	g.line.Refresh()
}
