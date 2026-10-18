package main

import (
	"fmt"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func gallerySystemPage() granio.CanvasObject {
	a := granio.CurrentApp()
	settings := a.Settings()
	palette := widget.NewLabel("")
	palette.Wrapping = granio.TextWrapWord
	update := func(granio.Settings) {
		s := galleryAppearance(settings)
		name := "не предоставлена"
		style := "не предоставлен"
		if s.DesktopTheme.Available {
			name = s.DesktopTheme.Value.Name
			style = s.DesktopTheme.Value.WidgetStyle
		}
		palette.SetText(fmt.Sprintf("Тема рабочего стола: %s\nСтиль: %s\nПолная палитра доступна: %t\nАкцент доступен: %t\nШрифт: %s\nИконки: %s", name, style, s.Colors.Available, s.Accent.Available, s.DesktopTheme.Value.Font, s.DesktopTheme.Value.IconTheme))
	}
	settings.AddListener(update)
	update(settings)
	backend := widget.NewLabel("")
	backend.Wrapping = granio.TextWrapWord
	refresh := func() {
		if d, ok := a.Driver().(desktop.DisplayDriver); ok {
			info := d.DisplayInfo()
			backend.SetText(fmt.Sprintf("Фактический backend: %s\nЗапрошен: %s\nТип сеанса из окружения: %s", info.Backend, info.Requested, info.SessionType))
		} else {
			backend.SetText("Backend: программный или иной драйвер")
		}
	}
	refresh()
	return container.NewVScroll(container.NewVBox(gallerySystemAppearance(settings), palette, backend, widget.NewButton("Обновить диагностику", refresh), widget.NewLabel("Стиль Union передаётся как метаданные. Палитра применяется через семантические роли Fyne.")))
}
