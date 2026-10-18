package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/software"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	softpainter "github.com/Uzer007-admin/granio/v2/internal/painter/software"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

// The reference scene uses fixed strings, bundled fonts, disabled motion and an
// opaque software painter. No native compositor or installed-font data is saved.
func galleryReferenceScene() granio.CanvasObject {
	normal := widget.NewButton("Сохранить · Save", nil)
	primary := widget.NewButton("Основное действие", nil)
	primary.Importance = widget.HighImportance
	disabled := widget.NewButton("Недоступно", nil)
	disabled.Disable()
	icon := widget.NewButtonWithIcon("", theme.SearchIcon(), nil)
	icon.AccessibleName = "Поиск"
	entry := widget.NewEntry()
	entry.SetText("Русский · English · 012345")
	readonly := widget.NewEntry()
	readonly.SetText("Readonly: выделение и копирование")
	readonly.SetReadOnly(true)
	invalid := widget.NewEntry()
	invalid.SetText("Ошибка")
	invalid.Validator = func(string) error { return errors.New("Неверное значение") }
	invalid.AlwaysShowValidationError = true
	invalid.FocusGained()
	password := widget.NewPasswordEntry()
	password.SetText("password")
	form := widget.NewForm(widget.NewFormItem("Поле", entry), widget.NewFormItem("Длинная подпись для проверки переноса текста", readonly), widget.NewFormItem("Ошибка + фокус", invalid), widget.NewFormItem("Пароль", password))
	form.Orientation = widget.Vertical
	check := widget.NewCheck("Выбрано · Checked", nil)
	check.SetChecked(true)
	selectBox := widget.NewSelect([]string{"Первый", "Второй"}, nil)
	selectBox.SetSelected("Второй")
	slider := widget.NewSlider(0, 100)
	slider.SetValue(65)
	progress := widget.NewProgressBar()
	progress.SetValue(.65)
	rich := widget.NewRichText(&widget.TextSegment{Text: "Русский + English ", Style: widget.RichTextStyle{TextRole: granio.TextRoleBody, Inline: true}}, &widget.TextSegment{Text: "Bold", Style: widget.RichTextStyle{TextRole: granio.TextRoleEmphasis}})
	log := roleLabel("/данные/very_long_path/0123456789.log", granio.TextRoleCode)
	log.Alignment = granio.TextAlignTrailing
	body := container.NewVBox(roleLabel("Типографика · Typography", granio.TextRoleTitle), roleLabel("Body · Русский · English · 012345", granio.TextRoleBody), roleLabel("Caption · Подпись", granio.TextRoleCaption),
		container.NewGridWithColumns(2, normal, primary, disabled, icon), form, container.NewHBox(check, selectBox), slider, progress, rich, log,
		widget.NewCard("Карточка", "Непрозрачная поверхность", widget.NewLabel("Чёткие контуры")))
	return container.NewVScroll(body)
}

func exportGalleryReferences(a granio.App, directory string) error {
	// The ci app uses Fyne's test driver, which needs a window to resolve objects
	// during constructor Refresh calls. It stays unshown and owns no OS surface.
	anchor := a.NewWindow("Opaque references")
	defer anchor.Close()
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}
	type reference struct {
		File, SHA256, Profile string
		DPI, TextScale        float32
		Width, Height         int
	}
	entries := make([]reference, 0, 39)
	for _, profile := range []string{"light", "dark", "high-contrast", "legacy"} {
		for _, dpi := range []float32{1, 1.5, 2} {
			for _, textScale := range []float32{1, 1.5, 2, 2.25} {
				if profile == "legacy" && textScale != 1 {
					continue
				}
				variant := theme.VariantLight
				th := theme.LightTheme()
				if profile != "legacy" {
					var err error
					th, err = theme.NewModern(theme.ModernOptions{TokenVersion: theme.ModernTokenVersion, TextScale: textScale, HighContrast: profile == "high-contrast", DisableMotion: true})
					if err != nil {
						return err
					}
					if profile == "dark" {
						variant = theme.VariantDark
					}
				}
				fixed := fixedReferenceTheme{Theme: th, variant: variant}
				if profile == "legacy" {
					setReferenceTheme(a, fixed)
				} else {
					setReferenceTheme(a, fixedModernReferenceTheme{fixed})
				}
				c := software.NewCanvasWithPainter(softpainter.NewPainter())
				c.SetPadded(false)
				c.SetScale(dpi)
				c.SetContent(galleryReferenceScene())
				c.Resize(granio.NewSize(780, 1800))
				capture := canvas.CaptureOpaque(c, th.Color(theme.ColorNameBackground, variant))
				name := strings.ReplaceAll(fmt.Sprintf("v1-%s-dpi%g-text%g.png", profile, dpi, textScale), ".5", "_5")
				path := filepath.Join(directory, name)
				file, err := os.Create(path)
				if err != nil {
					return err
				}
				encodeErr := png.Encode(file, capture)
				closeErr := file.Close()
				if encodeErr != nil {
					return encodeErr
				}
				if closeErr != nil {
					return closeErr
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				hash := sha256.Sum256(data)
				entries = append(entries, reference{name, hex.EncodeToString(hash[:]), profile, dpi, textScale, capture.Bounds().Dx(), capture.Bounds().Dy()})
			}
		}
	}
	data, err := json.MarshalIndent(map[string]any{"token_version": theme.ModernTokenVersion, "backend": "software opaque; bundled fonts; fixed strings; no DWM", "references": entries}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(directory, "manifest.json"), data, 0644); err != nil {
		return err
	}
	fmt.Printf("exported %d deterministic opaque references to %s\n", len(entries), directory)
	return nil
}

func setReferenceTheme(a granio.App, th granio.Theme) {
	a.Settings().SetTheme(th)
	// App settings normally clear themed SVGs and font metrics in the event
	// loop. This exporter deliberately renders without starting that loop.
	cache.ResetThemeCaches()
}

// Preserve optional style providers while fixing the requested light/dark
// variant, independently of the OS and persisted app preference.
type fixedReferenceTheme struct {
	granio.Theme
	variant granio.ThemeVariant
}

func (t fixedReferenceTheme) Color(name granio.ThemeColorName, _ granio.ThemeVariant) color.Color {
	return t.Theme.Color(name, t.variant)
}

type fixedModernReferenceTheme struct{ fixedReferenceTheme }

func (t fixedModernReferenceTheme) Typography(role granio.TextRole) granio.TypographyStyle {
	return t.Theme.(granio.TypographyTheme).Typography(role)
}
func (t fixedModernReferenceTheme) Motion(role granio.MotionRole) granio.MotionSpec {
	return t.Theme.(granio.MotionTheme).Motion(role)
}
func (t fixedModernReferenceTheme) ComponentStyle(kind granio.ComponentKind, state granio.VisualState, _ granio.ThemeVariant) granio.ComponentStyle {
	return t.Theme.(granio.ComponentTheme).ComponentStyle(kind, state, t.variant)
}
func (t fixedModernReferenceTheme) MaterialStyle(role granio.SurfaceRole, _ granio.ThemeVariant) granio.MaterialStyle {
	return t.Theme.(granio.MaterialTheme).MaterialStyle(role, t.variant)
}
func (t fixedModernReferenceTheme) SystemColors() (granio.SystemColors, bool) {
	return t.Theme.(granio.SystemColorTheme).SystemColors()
}
func (t fixedModernReferenceTheme) HighContrast() bool {
	return t.Theme.(interface{ HighContrast() bool }).HighContrast()
}
