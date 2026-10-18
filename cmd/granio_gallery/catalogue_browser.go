package main

import (
	"fmt"
	"strings"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

type gallerySample struct {
	id, title, category, description, keywords, code string
	experimental                                     bool
	build                                            func() granio.CanvasObject
}

func gallerySamples(w granio.Window) []gallerySample {
	return []gallerySample{
		{id: "widgets", title: "Элементы и состояния", category: "Основы", description: "Кнопки, поля, формы, меню и доступность.", keywords: "button entry form checkbox radio select slider progress dialog context menu wcag focus keyboard readonly", code: "button := widget.NewButton(\"Сохранить\", save)\nbutton.Importance = widget.HighImportance\nentry := widget.NewEntry()\nentry.SetReadOnly(true)", build: func() granio.CanvasObject { return galleryWidgetCatalogue(w) }},
		{id: "data-grid", title: "Таблица данных", category: "Данные", description: "10 000 строк: сортировка, фильтр, группы и проверка редактирования.", keywords: "table datagrid sort filter group edit validation виртуализация таблица фильтрация сортировка группировка", experimental: true, code: "grid, err := widget.NewDataGrid(columns, source)\nif err != nil { panic(err) }\ngrid.Table.StickyColumnCount = 1\n_ = grid.SetSort(\"size\", true)\n_ = grid.SetGroupColumn(\"status\")\n_ = grid.BeginEdit(widget.DataGridCell{RowID: \"row-00000\", ColumnID: \"name\"})", build: galleryDataGrid},
		{id: "collections", title: "Списки и деревья", category: "Данные", description: "Виртуализированные коллекции и множественный выбор.", keywords: "list tree selection virtualization список дерево 10000", code: "list := widget.NewList(length, create, update)\nlist.SelectionMode = widget.SelectionMultiple\nlist.Select(2)", build: galleryCollections},
		{id: "typography", title: "Типографика", category: "Оформление", description: "Семантические роли, русский текст, пути и числа.", keywords: "font typography baseline text scale caption title body code шрифт текст масштаб", code: "label := widget.NewLabel(\"Русский · English\")\nlabel.TextRole = granio.TextRoleTitle", build: func() granio.CanvasObject {
			objects := []granio.CanvasObject{}
			for _, item := range []struct {
				role granio.TextRole
				text string
			}{{granio.TextRoleTitle, "Заголовок · Title"}, {granio.TextRoleSubtitle, "Подзаголовок · Subtitle"}, {granio.TextRoleBody, "Русский · English · 0123456789"}, {granio.TextRoleCaption, "Подпись · Caption"}, {granio.TextRoleCode, "/данные/документы/report_2026.csv"}} {
				objects = append(objects, roleLabel(item.text, item.role))
			}
			return container.NewVScroll(container.NewVBox(objects...))
		}},
		{id: "system", title: "Системная тема", category: "Платформа", description: "Палитра, акцент, ограничения и фактический оконный backend.", keywords: "kde plasma union portal windows accent contrast theme wayland xwayland x11 система тема палитра", code: "th, err := theme.NewModern(theme.ModernOptions{\n    FollowSystem: true, FollowSystemPalette: true,\n})\nif err != nil { panic(err) }\napp.Settings().SetTheme(th)", build: func() granio.CanvasObject { return gallerySystemPage() }},
		{id: "materials", title: "Материалы внутри окна", category: "Оформление", description: "Две панели и popup поверх содержимого приложения.", keywords: "material acrylic blur popup alpha прозрачность размытие", code: "material := canvas.NewThemedMaterial(\n    theme.Current(), granio.SurfaceCard, app.Settings().ThemeVariant(),\n)\ncontent := container.NewStack(material, sharpContent)", build: func() granio.CanvasObject { return galleryInAppMaterials(w) }},
	}
}

func matchingGallerySamples(samples []gallerySample, query string) []int {
	tokens := strings.Fields(strings.ToLower(query))
	result := make([]int, 0, len(samples))
	for i, sample := range samples {
		text := strings.ToLower(sample.title + " " + sample.category + " " + sample.description + " " + sample.keywords + " " + sample.id)
		matched := true
		for _, token := range tokens {
			if !strings.Contains(text, token) {
				matched = false
				break
			}
		}
		if matched {
			result = append(result, i)
		}
	}
	return result
}

type galleryCatalogueBrowser struct {
	samples                           []gallerySample
	filtered                          []int
	pages                             map[int]granio.CanvasObject
	history                           []int
	position, current                 int
	synchronizing                     bool
	list                              *widget.List
	search                            *widget.Entry
	status, title, description, badge *widget.Label
	code                              *widget.Entry
	body                              *granio.Container
	back, forward                     *widget.Button
	root                              granio.CanvasObject
}

func newGalleryCatalogue(w granio.Window, requested string) *galleryCatalogueBrowser {
	b := &galleryCatalogueBrowser{samples: gallerySamples(w), pages: map[int]granio.CanvasObject{}, current: -1, position: -1}
	b.search = widget.NewEntry()
	b.search.PlaceHolder = "Поиск: таблица, WCAG, Wayland…"
	b.status = widget.NewLabel("")
	b.title = roleLabel("", granio.TextRoleTitle)
	b.description = widget.NewLabel("")
	b.description.Wrapping = granio.TextWrapWord
	b.badge = widget.NewLabel("Экспериментальный API: контракт может уточняться")
	b.code = widget.NewMultiLineEntry()
	b.code.TextStyle.Monospace = true
	b.code.SetReadOnly(true)
	b.body = container.NewStack(widget.NewLabel("Выберите пример"))
	b.list = widget.NewList(func() int { return len(b.filtered) }, func() granio.CanvasObject { return widget.NewLabel("") }, func(id widget.ListItemID, o granio.CanvasObject) {
		s := b.samples[b.filtered[id]]
		o.(*widget.Label).SetText(s.category + " · " + s.title)
	})
	b.list.OnSelected = func(id widget.ListItemID) {
		if !b.synchronizing && id >= 0 && id < len(b.filtered) {
			b.navigate(b.filtered[id], true)
		}
	}
	b.back = widget.NewButtonWithIcon("Назад", theme.NavigateBackIcon(), func() { b.move(-1) })
	b.forward = widget.NewButtonWithIcon("Вперёд", theme.NavigateNextIcon(), func() { b.move(1) })
	clear := widget.NewButton("Все примеры", func() { b.search.SetText("") })
	side := container.NewBorder(container.NewVBox(b.search, b.status, clear), nil, nil, nil, b.list)
	source := container.NewBorder(widget.NewLabel("Пример кода · фрагмент для копирования"), nil, nil, nil, b.code)
	detail := container.NewVSplit(b.body, source)
	detail.Offset = .72
	right := container.NewBorder(container.NewVBox(container.NewHBox(b.back, b.forward), b.title, b.description, b.badge), nil, nil, nil, detail)
	split := container.NewHSplit(side, right)
	split.Offset = .28
	b.root = split
	b.search.OnChanged = b.filter
	b.filter("")
	for index, sample := range b.samples {
		if sample.id == requested {
			b.navigate(index, true)
			break
		}
	}
	return b
}

func (b *galleryCatalogueBrowser) filter(query string) {
	b.filtered = matchingGallerySamples(b.samples, query)
	b.status.SetText(fmt.Sprintf("Найдено примеров: %d", len(b.filtered)))
	b.list.UnselectAll()
	b.list.Refresh()
	if len(b.filtered) == 0 {
		b.title.SetText("Ничего не найдено")
		b.description.SetText("Запрос: " + query + ". Попробуйте название элемента или его назначение.")
		b.badge.Hide()
		b.code.SetText("")
		b.body.Objects = []granio.CanvasObject{widget.NewLabel("Измените запрос или нажмите «Все примеры»")}
		b.body.Refresh()
		return
	}
	for _, index := range b.filtered {
		if index == b.current {
			b.show(index)
			return
		}
	}
	b.navigate(b.filtered[0], true)
}
func (b *galleryCatalogueBrowser) navigate(index int, record bool) {
	if index < 0 || index >= len(b.samples) {
		return
	}
	if record && (b.position < 0 || b.history[b.position] != index) {
		b.history = append(b.history[:b.position+1], index)
		if len(b.history) > 128 {
			b.history = b.history[len(b.history)-128:]
		}
		b.position = len(b.history) - 1
	}
	b.show(index)
}
func (b *galleryCatalogueBrowser) show(index int) {
	sample := b.samples[index]
	b.current = index
	page := b.pages[index]
	if page == nil {
		page = sample.build()
		b.pages[index] = page
	}
	b.title.SetText(sample.title)
	b.description.SetText(sample.description)
	if sample.experimental {
		b.badge.Show()
	} else {
		b.badge.Hide()
	}
	b.body.Objects = []granio.CanvasObject{page}
	b.body.Refresh()
	b.code.SetText(sample.code)
	b.synchronizing = true
	b.list.UnselectAll()
	for position, filtered := range b.filtered {
		if filtered == index {
			b.list.Select(position)
			break
		}
	}
	b.synchronizing = false
	if b.position > 0 {
		b.back.Enable()
	} else {
		b.back.Disable()
	}
	if b.position+1 < len(b.history) {
		b.forward.Enable()
	} else {
		b.forward.Disable()
	}
}
func (b *galleryCatalogueBrowser) move(delta int) {
	next := b.position + delta
	if next < 0 || next >= len(b.history) {
		return
	}
	b.position = next
	visible := false
	for _, index := range b.filtered {
		if index == b.history[next] {
			visible = true
			break
		}
	}
	if !visible {
		b.search.SetText("")
	}
	b.show(b.history[next])
}
func galleryCatalogue(w granio.Window, requested ...string) granio.CanvasObject {
	id := ""
	if len(requested) > 0 {
		id = requested[0]
	}
	return newGalleryCatalogue(w, id).root
}
