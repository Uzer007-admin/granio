package main

import (
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"time"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/container"
	"github.com/Uzer007-admin/granio/v3/driver"
	"github.com/Uzer007-admin/granio/v3/driver/desktop"
	"github.com/Uzer007-admin/granio/v3/theme"
	"github.com/Uzer007-admin/granio/v3/widget"
)

const (
	workspaceTitleBarHeight = 40
	workspaceHeaderHeight   = 52
	workspaceRailWidth      = 48
)

func runWorkspace(a granio.App, fps int, smoke bool, capturePath string) error {
	base, err := theme.NewModern(theme.ModernOptions{Density: theme.DensityCompact, FontPolicy: theme.FontSystem})
	if err != nil {
		return err
	}
	if driver, ok := a.Driver().(desktop.FrameRateDriver); ok {
		if err := driver.SetFrameRate(fps); err != nil {
			return err
		}
	}
	palette := workspaceTheme{Theme: base}
	appearance := workspacePreset(0)
	if !smoke {
		appearance = loadWorkspaceAppearance(a.Preferences())
	}
	palette.appearance = &appearance
	a.Settings().SetTheme(palette)
	w := a.NewWindow("Рабочая область Granio")
	w.SetPadded(false)
	titleBar, _ := w.(desktop.ClientTitleBarWindow)
	if titleBar != nil {
		titleErr := titleBar.SetClientTitleBar(desktop.ClientTitleBarOptions{
			Height: workspaceTitleBarHeight, LeadingReserved: 358, TrailingReserved: 139,
		})
		if titleErr != nil {
			titleBar = nil
		}
	}
	view := newWorkspace(w, palette, titleBar)
	view.preview = smoke
	view.configureBackdrop()
	if smoke {
		view.entry.ReadOnly = true
	}
	view.rendererButton.SetText(workspaceRendererName(workspaceRenderMode(a, w)))
	w.SetContent(view.root)
	w.Resize(granio.NewSize(1240, 848))
	w.CenterOnScreen()
	if smoke {
		a.Lifecycle().SetOnStarted(func() {
			time.AfterFunc(350*time.Millisecond, func() {
				granio.Do(func() {
					err = captureWorkspace(w, capturePath)
					w.Close()
				})
			})
		})
	}
	w.Show()
	if mode := workspaceRenderMode(a, w); mode != driver.RenderAuto {
		name := workspaceRendererName(mode)
		if view.rendererButton.Text != name {
			view.rendererButton.SetText(name)
		}
		fmt.Println("Рендерер", name)
	}
	if titleBar != nil && !titleBar.ClientTitleBarActive() {
		view.windowControls.Hide()
	}
	a.Run()
	return err
}

func captureWorkspace(w granio.Window, path string) error {
	frame := w.Canvas().Capture()
	if frame == nil || frame.Bounds().Empty() {
		return errors.New("рабочая область вернула пустой кадр")
	}
	fmt.Printf("Рабочая область: %d×%d\n", frame.Bounds().Dx(), frame.Bounds().Dy())
	if path == "" {
		return nil
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	return errors.Join(png.Encode(f, frame), f.Close())
}

type workspaceView struct {
	w                  granio.Window
	theme              workspaceTheme
	preview            bool
	appearanceEdit     *granio.Animation
	appearanceUI       func()
	surfaces           []workspaceMaterialTarget
	messageSurfaces    []workspaceMaterialTarget
	composerGlass      *canvas.Material
	visualStatus       desktop.VisualStatus
	visualError        error
	backdropConfigured bool
	backdropMode       desktop.WindowBackdrop
	controlStyle       workspaceOptics
	materialStyles     [workspaceGroupCount]workspaceOptics
	root               granio.CanvasObject
	windowControls     *granio.Container
	sidebarPane        *container.SidePanel
	documentPane       *container.SidePanel
	documentContent    *granio.Container
	toolbarPane        *widget.Button
	leftToggle         *widget.Button
	title              *widget.Label
	entry              *widget.Entry
	rendererButton     *widget.Button
	send               *widget.Button
	usage              *workspaceUsage
	confirmSend        bool
	conversationView   *granio.Container
	composeHeight      float32
	composeTarget      float32
	composeMaxHeight   float32
	composeAnimation   *granio.Animation
	messageAnimation   *granio.Animation
	movingMessage      *granio.Container
	feed               *workspaceFlow
	feedScroll         *container.Scroll
	threads            []workspaceThread
	docScroll          *container.Scroll
	docName            *canvas.Text
	history            *widget.List
	newDialog          *widget.Button
	project            int
	docTabs            []*workspaceDocTab
	docOrder           []int
	tabStrip           *granio.Container
	draggedTab         int
	slidingTab         int
	tabDragX           float32
	tabSlide           float32
	tabAnimation       *granio.Animation
	tabHighlight       *canvas.Rectangle
	tabHighlightAt     float32
	highlightMotion    *granio.Animation
	doc                int
}

func newWorkspace(w granio.Window, palette workspaceTheme, titleBar desktop.ClientTitleBarWindow) *workspaceView {
	if palette.appearance == nil {
		appearance := workspacePreset(0)
		palette.appearance = &appearance
	}
	v := &workspaceView{
		w: w, theme: palette, project: -1,
		docOrder: []int{0, 1, 2}, draggedTab: -1, slidingTab: -1,
	}
	central := v.conversation()
	documents := v.documents()
	v.documentPane = container.NewSidePanel(central, documents, container.PanelTrailing, 413)
	sidebar := v.sidebar()
	v.sidebarPane = container.NewSidePanel(v.documentPane, sidebar, container.PanelLeading, 239)
	rail, top := v.rail(), v.topBar(titleBar)
	railSurface := canvas.NewRectangle(workspaceRail)
	topSurface := canvas.NewHorizontalGradient(workspaceRail, workspaceTop)
	v.root = workspaceArrange(granio.NewSize(560, 480), func(s granio.Size) {
		width, height := s.Width, s.Height
		workspacePlace(railSurface, 0, workspaceTitleBarHeight, workspaceRailWidth, height-workspaceTitleBarHeight)
		workspacePlace(topSurface, 0, 0, width, workspaceTitleBarHeight)
		workspacePlace(top, 0, 0, width, workspaceTitleBarHeight)
		workspacePlace(rail, 0, workspaceTitleBarHeight, workspaceRailWidth, height-workspaceTitleBarHeight)
		workspacePlace(v.sidebarPane, workspaceRailWidth, workspaceTitleBarHeight, width-workspaceRailWidth-1, height-workspaceTitleBarHeight-1)
	}, railSurface, topSurface, v.sidebarPane, rail, top)
	v.selectDocument(0)
	v.title.SetText("Рабочая область")
	v.materialStyles = v.theme.appearance.Groups
	v.controlStyle = v.materialStyles[workspaceControls]
	return v
}

func (v *workspaceView) setSidebar(open bool) {
	v.sidebarPane.SetPanelOpen(open)
	v.leftToggle.SetSelected(!open)
}

func (v *workspaceView) setDocuments(open bool) {
	v.documentPane.SetPanelOpen(open)
	v.toolbarPane.SetSelected(!open)
}

func (v *workspaceView) focusComposer() { v.w.Canvas().Focus(v.entry) }

func (v *workspaceView) newDraft() {
	v.entry.SetText("")
	v.focusComposer()
}

func (v *workspaceView) menu(button *widget.Button, items ...*granio.MenuItem) {
	popup := widget.NewPopUpMenu(granio.NewMenu("", items...), v.w.Canvas())
	popup.Material = v.popupMaterial()
	pos := granio.CurrentApp().Driver().AbsolutePositionForObject(button)
	popup.ShowAtPosition(pos.AddXY(0, button.Size().Height))
}

func (v *workspaceView) topBar(titleBar desktop.ClientTitleBarWindow) granio.CanvasObject {
	back := workspaceButton("", "Назад", "back", workspaceSubtle, func() { v.selectProject(v.project - 1) })
	next := workspaceButton("", "Вперёд", "next", workspaceSubtle, func() { v.selectProject(v.project + 1) })
	v.leftToggle = workspaceButton("", "Переключить боковую панель", "panels", workspaceMuted,
		func() { v.setSidebar(!v.sidebarPane.PanelOpen()) })
	var file, edit, view, help *widget.Button
	file = workspaceButton("Файл", "", "", workspaceMuted, func() {
		v.menu(file, granio.NewMenuItem("Новый диалог", v.newConversation), granio.NewMenuItem("Закрыть окно", v.w.Close))
	})
	edit = workspaceButton("Правка", "", "", workspaceMuted, func() {
		v.menu(edit, granio.NewMenuItem("Вернуться к вводу", v.focusComposer), granio.NewMenuItem("Очистить поле", v.newDraft))
	})
	view = workspaceButton("Вид", "", "", workspaceMuted, func() {
		v.menu(view, granio.NewMenuItem("Оформление", v.showAppearance),
			granio.NewMenuItem("Боковая панель", func() { v.setSidebar(!v.sidebarPane.PanelOpen()) }),
			granio.NewMenuItem("Документы", func() { v.setDocuments(!v.documentPane.PanelOpen()) }),
			granio.NewMenuItem("Перенести панель документов", v.moveDocuments))
	})
	help = workspaceButton("Справка", "", "", workspaceMuted, func() {
		v.menu(help, granio.NewMenuItem("Открыть README", func() { v.selectDocument(2); v.setDocuments(true) }))
	})
	v.windowControls = container.NewWithoutLayout()
	if titleBar != nil {
		minimize := workspaceButton("", "Свернуть", "minimize", workspaceWhite, titleBar.Minimize)
		maximize := workspaceButton("", "Развернуть или восстановить", "maximize", workspaceWhite, titleBar.ToggleMaximize)
		close := workspaceButton("", "Закрыть", "close", workspaceWhite, v.w.Close)
		close.Importance = widget.DangerImportance
		restoreIcon, maximizeIcon := workspaceIcon("restore", workspaceWhite), maximize.Icon
		v.windowControls = workspaceArrange(granio.NewSize(138, 32), func(s granio.Size) {
			icon := maximizeIcon
			if titleBar.Maximized() {
				icon = restoreIcon
			}
			if maximize.Icon != icon {
				maximize.SetIcon(icon)
			}
			for i, b := range []*widget.Button{minimize, maximize, close} {
				workspacePlace(b, float32(i)*46, 0, 46, 32)
			}
		}, minimize, maximize, close)
	}
	return workspaceArrange(granio.Size{}, func(s granio.Size) {
		workspacePlace(back, 2, 1, 30, 28)
		workspacePlace(next, 35, 1, 30, 28)
		workspacePlace(v.leftToggle, 69, 1, 30, 28)
		workspacePlace(file, 109, 1, 54, 28)
		workspacePlace(edit, 166, 1, 64, 28)
		workspacePlace(view, 234, 1, 48, 28)
		workspacePlace(help, 286, 1, 68, 28)
		workspacePlace(v.windowControls, s.Width-138, 0, 138, 32)
	}, back, next, v.leftToggle, file, edit, view, help, v.windowControls)
}

func (v *workspaceView) rail() granio.CanvasObject {
	home := workspaceButton("", "Рабочая область", "home", workspaceWhite, func() { v.setSidebar(true) })
	home.Selected = true
	layers := workspaceButton("", "Документы", "layers", workspaceMuted, func() { v.setDocuments(!v.documentPane.PanelOpen()) })
	history := workspaceButton("", "История", "history", workspaceMuted, func() { v.setSidebar(!v.sidebarPane.PanelOpen()) })
	terminal := workspaceButton("", "Журнал проверок", "terminal", workspaceMuted, func() { v.selectDocument(0); v.setDocuments(true) })
	grid := workspaceButton("", "Отчёт о компонентах", "grid", workspaceMuted, func() { v.selectDocument(1); v.setDocuments(true) })
	var more *widget.Button
	more = workspaceButton("", "Параметры рабочей области", "overflow", workspaceMuted, func() {
		v.menu(more, granio.NewMenuItem("Показать обе панели", func() { v.setSidebar(true); v.setDocuments(true) }),
			granio.NewMenuItem("Режим чтения", func() { v.setSidebar(false); v.setDocuments(false) }))
	})
	rule := canvas.NewRectangle(workspaceColor(0x2b3140))
	objects := []granio.CanvasObject{home, layers, history, terminal, more, rule, grid}
	return workspaceArrange(granio.Size{}, func(granio.Size) {
		for i, b := range []*widget.Button{home, layers, history, terminal, more} {
			workspacePlace(b, 5, 9+float32(i)*44, 36, 36)
		}
		workspacePlace(rule, 11, 229, 24, 1)
		workspacePlace(grid, 5, 236, 36, 36)
	}, objects...)
}

func (v *workspaceView) centerHeader() granio.CanvasObject {
	folder := canvas.NewImageFromResource(workspaceIcon("folder", workspaceMuted))
	folder.FillMode = canvas.ImageFillContain
	v.title = widget.NewLabelWithStyle("", granio.TextAlignLeading, granio.TextStyle{Bold: true})
	v.title.Truncation = granio.TextTruncateEllipsis
	chevron := canvas.NewImageFromResource(workspaceIcon("chevron", workspaceMuted))
	chevron.FillMode = canvas.ImageFillContain
	var share *widget.Button
	share = workspaceButton("", "Действия с материалами", "share", workspaceText, func() {
		v.menu(share, granio.NewMenuItem("Открыть материалы", func() { v.setDocuments(true) }))
	})
	v.toolbarPane = workspaceButton("", "Переключить панель документов", "settings", workspaceText,
		func() { v.setDocuments(!v.documentPane.PanelOpen()) })
	rule := canvas.NewRectangle(workspaceBorder)
	return workspaceArrange(granio.NewSize(280, workspaceHeaderHeight), func(s granio.Size) {
		workspacePlace(folder, 20, 19, 16, 16)
		workspacePlace(v.title, 37, 11, min(237, s.Width-145), 34)
		workspacePlace(chevron, min(278, s.Width-103), 21, 12, 12)
		workspacePlace(share, s.Width-72, 10, 30, 32)
		workspacePlace(v.toolbarPane, s.Width-38, 10, 30, 32)
		workspacePlace(rule, 0, s.Height-1, s.Width, 1)
	}, folder, v.title, chevron, share, v.toolbarPane, rule)
}

type workspaceLayout struct {
	minimum granio.Size
	arrange func(granio.Size)
}

func (l workspaceLayout) Layout(_ []granio.CanvasObject, size granio.Size) { l.arrange(size) }
func (l workspaceLayout) MinSize([]granio.CanvasObject) granio.Size        { return l.minimum }

func workspaceArrange(minimum granio.Size, arrange func(granio.Size), objects ...granio.CanvasObject) *granio.Container {
	return container.New(workspaceLayout{minimum: minimum, arrange: arrange}, objects...)
}

func workspacePlace(o granio.CanvasObject, x, y, width, height float32) {
	o.Move(granio.NewPos(x, y))
	o.Resize(granio.NewSize(max(0, width), max(0, height)))
}

func workspaceShow(o granio.CanvasObject, visible bool) {
	if visible {
		o.Show()
	} else {
		o.Hide()
	}
}

func workspaceCanvasText(text string, c color.Color, size float32, bold bool) *canvas.Text {
	t := canvas.NewText(text, c)
	t.TextSize, t.TextStyle.Bold = size, bold
	return t
}

func workspaceSurface(c color.Color, radius float32) *canvas.Rectangle {
	r := canvas.NewRectangle(c)
	r.CornerRadius = radius
	return r
}

func workspaceButton(text, accessible, icon string, tint color.NRGBA, tapped func()) *widget.Button {
	b := widget.NewButton(text, tapped)
	b.Style = widget.ButtonStyleToolbar
	b.AccessibleName = accessible
	if icon != "" {
		b.Icon = workspaceIcon(icon, tint)
	}
	return b
}
