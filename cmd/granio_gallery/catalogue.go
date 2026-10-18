package main

import (
	"errors"
	"fmt"
	"image/color"
	"net/url"
	"strconv"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/dialog"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func roleLabel(text string, role granio.TextRole) *widget.Label {
	l := widget.NewLabel(text)
	l.TextRole, l.Wrapping = role, granio.TextWrapWord
	return l
}

func galleryWidgetCatalogue(w granio.Window) granio.CanvasObject {
	buttons := make([]granio.CanvasObject, 0, 6)
	for _, importance := range []widget.Importance{widget.MediumImportance, widget.HighImportance, widget.LowImportance, widget.DangerImportance, widget.SuccessImportance, widget.WarningImportance} {
		button := widget.NewButton("Действие", nil)
		button.Importance = importance
		buttons = append(buttons, button)
	}
	search := widget.NewToolbarAction(theme.SearchIcon(), func() {})
	search.AccessibleName = "Поиск"
	save := widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {})
	save.AccessibleName = "Сохранить"
	toolbar := widget.NewToolbar(search, save, widget.NewToolbarSeparator(), widget.NewToolbarSpacer())
	entry := widget.NewEntry()
	entry.SetText("Readonly: можно выделить и скопировать")
	entry.SetReadOnly(true)
	invalid := widget.NewEntry()
	invalid.SetText("Ошибка")
	invalid.Validator = func(string) error {
		return errors.New("Введите корректные данные: длинное пояснение не должно перекрывать следующую строку формы.")
	}
	invalid.AlwaysShowValidationError = true
	password := widget.NewPasswordEntry()
	password.SetText("password")
	form := widget.NewForm(widget.NewFormItem("Длинная подпись для проверки переноса и увеличения текста", invalid), widget.NewFormItem("Только чтение", entry), widget.NewFormItem("Пароль", password))
	form.Items[0].HintText = "Проверьте ошибку после редактирования и потери фокуса."
	choices := widget.NewSelect([]string{"Первый вариант", "Второй вариант"}, nil)
	choices.SetSelectedIndex(0)
	check := widget.NewCheck("Выбранное состояние", nil)
	check.SetChecked(true)
	radio := widget.NewRadioGroup([]string{"Вариант A", "Вариант Б"}, nil)
	radio.SetSelected("Вариант A")
	slider := widget.NewSlider(0, 100)
	slider.SetValue(65)
	progress := widget.NewProgressBar()
	progress.SetValue(.65)
	link, _ := url.Parse("https://fyne.io/")
	numbers := widget.NewLabelWithStyle("1 234 567,89", granio.TextAlignTrailing, granio.TextStyle{Monospace: true})
	numbers.TextRole = granio.TextRoleCode
	rich := widget.NewRichText(&widget.TextSegment{Text: "Русский + English 012345 ", Style: widget.RichTextStyle{TextRole: granio.TextRoleBody, Inline: true}}, &widget.TextSegment{Text: "Emphasis", Style: widget.RichTextStyle{TextRole: granio.TextRoleEmphasis}})
	accordion := widget.NewAccordion(widget.NewAccordionItem("Разворачиваемая панель", roleLabel("Содержимое панели", granio.TextRoleBody)))
	accordion.Open(0)
	tabs := container.NewAppTabs(container.NewTabItem("Первая", roleLabel("Содержимое вкладки", granio.TextRoleBody)), container.NewTabItem("Вторая", widget.NewLabel("Вторая вкладка")))
	menus := container.NewHBox(widget.NewButton("Меню", func() {
		menu := widget.NewPopUpMenu(granio.NewMenu("", granio.NewMenuItem("Команда", func() {}), granio.NewMenuItem("Закрыть", func() {})), w.Canvas())
		menu.ShowAtPosition(granio.NewPos(30, 40))
	}), widget.NewButton("Диалог", func() {
		dialog.ShowInformation("Диалог Fyne", "Escape закрывает диалог и возвращает фокус.", w)
	}), widget.NewButton("FileDialog", func() { dialog.ShowFileOpen(func(granio.URIReadCloser, error) {}, w) }))
	page := container.NewVBox(roleLabel("Каталог штатных виджетов", granio.TextRoleTitle), toolbar, container.NewGridWithColumns(3, buttons...), form,
		container.NewHBox(check, choices), radio, slider, progress, numbers, rich, widget.NewHyperlink("Ссылка Fyne", link), accordion, tabs, menus,
		widget.NewCard("Карточка", "Обычные поверхности и скругления", roleLabel("Все элементы используют штатные рендереры Fyne.", granio.TextRoleBody)))
	return container.NewVScroll(page)
}

func galleryCollections() granio.CanvasObject {
	list := widget.NewList(func() int { return 10000 }, func() granio.CanvasObject { return widget.NewLabel("") }, func(id widget.ListItemID, object granio.CanvasObject) {
		object.(*widget.Label).SetText(fmt.Sprintf("Строка %05d · Русский", id))
	})
	list.SelectionMode = widget.SelectionMultiple
	list.Select(2)
	tree := widget.NewTree(func(id widget.TreeNodeID) []widget.TreeNodeID {
		if id != "" {
			return nil
		}
		children := make([]widget.TreeNodeID, 10000)
		for i := range children {
			children[i] = strconv.Itoa(i)
		}
		return children
	}, func(id widget.TreeNodeID) bool { return id == "" }, func(bool) granio.CanvasObject { return widget.NewLabel("") }, func(id widget.TreeNodeID, _ bool, object granio.CanvasObject) {
		object.(*widget.Label).SetText("Узел " + id)
	})
	tree.SelectionMode = widget.SelectionMultiple
	tree.Select("3")
	split := container.NewHSplit(container.NewBorder(roleLabel("List · 10 000", granio.TextRoleSubtitle), nil, nil, nil, list), container.NewBorder(roleLabel("Tree · 10 000", granio.TextRoleSubtitle), nil, nil, nil, tree))
	return split
}

func galleryInAppMaterials(w granio.Window) granio.CanvasObject {
	source := canvas.NewRasterWithPixels(func(x, y, width, height int) color.Color {
		if (x/24+y/24)%2 == 0 {
			return color.NRGBA{R: 40, G: 110, B: 210, A: 255}
		}
		return color.NRGBA{R: 235, G: 175, B: 65, A: 255}
	})
	th, variant := granio.CurrentApp().Settings().Theme(), granio.CurrentApp().Settings().ThemeVariant()
	left := canvas.NewThemedMaterial(th, granio.SurfaceCard, variant)
	right := canvas.NewThemedMaterial(th, granio.SurfaceContent, variant)
	granio.CurrentApp().Settings().AddListener(func(settings granio.Settings) {
		for _, surface := range []struct {
			material *canvas.Material
			role     granio.SurfaceRole
		}{{left, granio.SurfaceCard}, {right, granio.SurfaceContent}} {
			resolved := canvas.NewThemedMaterial(settings.Theme(), surface.role, settings.ThemeVariant())
			m := surface.material
			m.Enabled, m.BlurRadius, m.CornerRadius, m.Noise, m.Tint, m.Fallback = resolved.Enabled, resolved.BlurRadius, resolved.CornerRadius, resolved.Noise, resolved.Tint, resolved.Fallback
			m.Refresh()
		}
	})
	left.Move(granio.NewPos(30, 90))
	left.Resize(granio.NewSize(300, 230))
	right.Move(granio.NewPos(330, 180))
	right.Resize(granio.NewSize(280, 240))
	foreground := container.NewVBox(roleLabel("Резкий текст поверх InAppAcrylic", granio.TextRoleSubtitle), widget.NewEntry(), widget.NewButton("Popup с материалом", func() {
		popup := widget.NewPopUp(container.NewVBox(widget.NewLabel("Резкий текст popup"), widget.NewButton("Готово", func() {})), w.Canvas())
		popup.Material = canvas.NewThemedMaterial(granio.CurrentApp().Settings().Theme(), granio.SurfacePopup, granio.CurrentApp().Settings().ThemeVariant())
		popup.ShowAtPosition(granio.NewPos(200, 130))
	}))
	foreground.Move(granio.NewPos(45, 105))
	foreground.Resize(granio.NewSize(255, 190))
	panels := container.NewWithoutLayout(left, right, foreground)
	return container.NewStack(source, panels)
}
