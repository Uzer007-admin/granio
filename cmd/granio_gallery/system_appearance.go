package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func galleryAppearance(settings granio.Settings) granio.SystemAppearance {
	if provider, ok := settings.(granio.SystemAppearanceProvider); ok {
		return provider.SystemAppearance()
	}
	return granio.SystemAppearance{}
}

func gallerySystemAppearance(settings granio.Settings) granio.CanvasObject {
	label := widget.NewLabel("")
	label.Wrapping = granio.TextWrapWord
	update := func(granio.Settings) {
		snapshot := galleryAppearance(settings)
		if !snapshot.Available {
			label.SetText("Системные настройки недоступны; используются настройки приложения")
			return
		}
		text := "недоступен"
		if snapshot.TextScale.Available {
			text = fmt.Sprintf("%.0f%%", snapshot.TextScale.Value*100)
		}
		label.SetText(fmt.Sprintf("Система: текст %s · контраст %t · прозрачность %t · анимации %t",
			text, snapshot.HighContrast.Available && snapshot.HighContrast.Value,
			snapshot.Transparency.Available && snapshot.Transparency.Value, settings.ShowAnimations()))
	}
	settings.AddListener(update)
	update(settings)
	return label
}

func writeGalleryAppearanceReport(path string, beforeLoop granio.SystemAppearance, settings granio.Settings, window granio.Window) error {
	var display desktop.DisplayInfo
	if provider, ok := granio.CurrentApp().Driver().(desktop.DisplayDriver); ok {
		display = provider.DisplayInfo()
	}
	var visuals desktop.VisualStatus
	if provider, ok := window.(desktop.WindowVisuals); ok {
		visuals = provider.VisualStatus()
	}
	data, err := json.MarshalIndent(map[string]any{
		"display":            display,
		"before_driver_loop": beforeLoop, "at_capture": galleryAppearance(settings),
		"theme_text_size": settings.Theme().Size(theme.SizeNameText), "theme_icon_size": settings.Theme().Size(theme.SizeNameInlineIcon),
		"theme_primary": settings.Theme().Color(theme.ColorNamePrimary, settings.ThemeVariant()),
		"visuals":       visuals,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
