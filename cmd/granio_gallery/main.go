// Галерея Granio проверяет оформление, ввод и виртуализированные компоненты
// теми же средствами, которые используются в приложениях.
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func main() {
	fps := flag.Int("fps", 0, "частота кадров (0 — частота дисплея)")
	backdrop := flag.String("backdrop", "opaque", "материал окна: opaque, transparent, system-blur, mica, mica-alt, acrylic")
	smoke := flag.Bool("smoke", false, "снять кадр и штатно завершить работу")
	capturePath := flag.String("capture-path", "", "сохранить кадр PNG при проверке -smoke")
	appearanceReport := flag.String("appearance-report", "", "сохранить диагностику темы и оконного backend при -smoke")
	profile := flag.String("profile", "modern", "профиль оформления: legacy, modern, materials")
	textScaleFlag := flag.Float64("text-scale", 0, "независимый масштаб текста (0 — системный)")
	compact := flag.Bool("compact", false, "компактные отступы")
	fontPolicy := flag.String("font-policy", "bundled", "политика шрифтов: bundled, system")
	systemPalette := flag.Bool("system-palette", false, "использовать полную системную палитру при её доступности")
	sample := flag.String("sample", "", "открыть пример: widgets, data-grid, collections, typography, system, materials")
	references := flag.String("references-dir", "", "экспортировать воспроизводимые непрозрачные эталоны v1 и завершить работу")
	flag.Parse()
	if *sample != "" {
		found := false
		for _, item := range gallerySamples(nil) {
			if item.id == *sample {
				found = true
				break
			}
		}
		if !found {
			fmt.Fprintln(os.Stderr, "Неизвестный пример; допустимо: widgets, data-grid, collections, typography, system, materials")
			os.Exit(2)
		}
	}
	initialBackdrop, err := galleryBackdrop(*backdrop)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	a := app.NewWithID("io.granio.visual-gallery")
	if *references != "" {
		if err := exportGalleryReferences(a, *references); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *profile != "legacy" && *profile != "modern" && *profile != "materials" {
		fmt.Fprintln(os.Stderr, "профиль должен быть legacy, modern или materials")
		os.Exit(2)
	}
	var beforeLoop granio.SystemAppearance
	a.Lifecycle().SetOnStarted(func() { beforeLoop = galleryAppearance(a.Settings()) })
	w := a.NewWindow("Галерея компонентов Granio")
	if scalable, ok := w.(desktop.ScalableWindow); ok {
		if err := scalable.SetScalePolicy(desktop.ScaleExact); err != nil {
			panic(err)
		}
	}
	if driver, ok := a.Driver().(desktop.FrameRateDriver); ok {
		if err := driver.SetFrameRate(*fps); err != nil {
			panic(err)
		}
	}
	options := theme.ModernOptions{FollowSystem: true, FollowSystemPalette: *systemPalette, TextScale: float32(*textScaleFlag)}
	if *compact {
		options.Density = theme.DensityCompact
	}
	if *fontPolicy == "system" {
		options.FontPolicy = theme.FontSystem
	} else if *fontPolicy != "bundled" {
		fmt.Fprintln(os.Stderr, "политика шрифтов должна быть bundled или system")
		os.Exit(2)
	}
	selectedProfile := *profile
	apply := func() {
		if selectedProfile == "legacy" {
			a.Settings().SetTheme(theme.DefaultTheme())
			return
		}
		if err := desktop.ApplyModernProfile(a, w, options); err != nil {
			panic(err)
		}
	}
	apply()
	entry := widget.NewMultiLineEntry()
	entry.SetText("Русский текст · English · 日本語 · 한글 · 😀\nПроверьте ввод, выделение, композицию и Undo.")
	entry.SetMinRowsVisible(3)
	password := widget.NewPasswordEntry()
	password.SetPlaceHolder("Пароль: текст скрыт от native surrounding text")
	status := widget.NewLabel("Готово")
	entry.OnChanged = func(text string) {
		status.SetText(fmt.Sprintf("Подтверждённый текст: %d символов", len([]rune(text))))
	}
	contrast := widget.NewCheck("Повышенная контрастность", func(value bool) { options.HighContrast = value; apply() })
	palette := widget.NewCheck("Палитра системы", func(value bool) { options.FollowSystemPalette = value; apply() })
	palette.SetChecked(options.FollowSystemPalette)
	density := widget.NewCheck("Компактные отступы", func(value bool) {
		options.Density = theme.DensityComfortable
		if value {
			options.Density = theme.DensityCompact
		}
		apply()
	})
	density.SetChecked(*compact)
	textScale := widget.NewSelect([]string{"Системный", "100%", "150%", "200%", "225%"}, func(value string) {
		switch value {
		case "Системный":
			options.TextScale = 0
		case "100%":
			options.TextScale = 1
		case "150%":
			options.TextScale = 1.5
		case "200%":
			options.TextScale = 2
		case "225%":
			options.TextScale = 2.25
		}
		apply()
	})
	if *textScaleFlag == 0 {
		textScale.SetSelected("Системный")
	}
	profiles := widget.NewSelect([]string{"legacy", "modern", "materials"}, func(value string) { selectedProfile = value; apply() })
	profiles.SetSelected(selectedProfile)
	fonts := widget.NewCheck("Шрифт системы", func(value bool) {
		options.FontPolicy = theme.FontBundled
		if value {
			options.FontPolicy = theme.FontSystem
		}
		apply()
	})
	fonts.SetChecked(options.FontPolicy == theme.FontSystem)
	motion := widget.NewCheck("Анимации", func(value bool) { options.DisableMotion = !value; apply() })
	motion.SetChecked(true)
	readonly := widget.NewCheck("Только чтение", entry.SetReadOnly)
	rows := widget.NewTable(func() (int, int) { return 10000, 3 }, func() granio.CanvasObject { return widget.NewLabel("") }, func(id widget.TableCellID, object granio.CanvasObject) {
		text := fmt.Sprintf("%d", id.Row+1)
		if id.Col == 1 {
			text = "Файл с длинным русским названием"
		}
		if id.Col == 2 {
			text = fmt.Sprintf("/данные/каталог_%05d/日本語.txt", id.Row)
		}
		object.(*widget.Label).SetText(text)
	})
	rows.SelectionMode = widget.SelectionMultiple
	rows.SetColumnWidth(0, 80)
	rows.SetColumnWidth(1, 310)
	rows.SetColumnWidth(2, 440)
	form := widget.NewForm(widget.NewFormItem("Текст и IME", entry), widget.NewFormItem("Пароль", password))
	controls := container.NewVBox(form, container.NewHBox(profiles, contrast, density, widget.NewLabel("Размер текста"), textScale), container.NewHBox(fonts, motion, readonly, palette), gallerySystemAppearance(a.Settings()), status)
	visuals := galleryWindowVisuals(w, initialBackdrop)
	tabs := container.NewAppTabs(container.NewTabItem("Элементы", container.NewBorder(controls, nil, nil, nil, rows)), container.NewTabItem("Материалы окна", visuals), container.NewTabItem("Каталог", galleryCatalogue(w, *sample)), container.NewTabItem("Коллекции", galleryCollections()), container.NewTabItem("Материалы содержимого", galleryInAppMaterials(w)))
	if *sample != "" {
		tabs.SelectIndex(2)
	} else if initialBackdrop != desktop.BackdropOpaque {
		tabs.SelectIndex(1)
	} else if selectedProfile == "materials" {
		tabs.SelectIndex(4)
	}
	w.SetContent(tabs)
	w.Resize(granio.NewSize(1060, 760))
	if drop, ok := w.(desktop.DataDropWindow); ok {
		_ = drop.SetOnDataDropped([]string{granio.MIMETextURIList, granio.MIMETextHTML, granio.MIMETextPlain, granio.MIMEImagePNG}, func(_ granio.Position, data map[string][]byte) {
			for mime, value := range data {
				status.SetText(fmt.Sprintf("Получено %s: %d байт", mime, len(value)))
			}
		})
	}
	if *smoke {
		go func() {
			time.Sleep(700 * time.Millisecond)
			granio.Do(func() {
				image := w.Canvas().Capture()
				if image.Bounds().Empty() {
					fmt.Fprintln(os.Stderr, "получен пустой кадр")
					os.Exit(1)
				}
				fmt.Printf("Кадр галереи: %v; символов: %d; таблица: 10000×3; пример: %s\n", image.Bounds(), len([]rune(entry.Text)), *sample)
				if *capturePath != "" {
					file, err := os.Create(*capturePath)
					if err == nil {
						err = png.Encode(file, image)
						closeErr := file.Close()
						if err == nil {
							err = closeErr
						}
					}
					if err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(1)
					}
				}
				if *appearanceReport != "" {
					if err := writeGalleryAppearanceReport(*appearanceReport, beforeLoop, a.Settings(), w); err != nil {
						fmt.Fprintln(os.Stderr, err)
						os.Exit(1)
					}
				}
				w.Close()
				a.Quit()
			})
		}()
	}
	w.ShowAndRun()
}
