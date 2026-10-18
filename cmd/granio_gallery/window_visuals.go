package main

import (
	"fmt"
	"image/color"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/widget"
)

var galleryBackdrops = []desktop.WindowBackdrop{desktop.BackdropOpaque, desktop.BackdropTransparent, desktop.BackdropSystemBlur, desktop.BackdropMica, desktop.BackdropMicaAlt, desktop.BackdropDesktopAcrylic}

func galleryBackdrop(value string) (desktop.WindowBackdrop, error) {
	switch value {
	case "opaque":
		return desktop.BackdropOpaque, nil
	case "mica":
		return desktop.BackdropMica, nil
	case "mica-alt":
		return desktop.BackdropMicaAlt, nil
	case "acrylic":
		return desktop.BackdropDesktopAcrylic, nil
	case "transparent":
		return desktop.BackdropTransparent, nil
	case "system-blur":
		return desktop.BackdropSystemBlur, nil
	default:
		return 0, fmt.Errorf("unknown backdrop %q (opaque, transparent, system-blur, mica, mica-alt, acrylic)", value)
	}
}

// The alpha scene uses ordinary Fyne objects. Unpainted areas reveal the native
// background; controls keep their own opaque surfaces and standard native frame.
func galleryWindowVisuals(w granio.Window, initial desktop.WindowBackdrop) granio.CanvasObject {
	status := widget.NewLabel("Подготовка окна")
	status.Wrapping = granio.TextWrapWord
	visuals, ok := w.(desktop.WindowVisuals)
	if !ok {
		status.SetText("Этот драйвер использует непрозрачное окно")
	}
	values := make([]string, len(galleryBackdrops))
	for i, b := range galleryBackdrops {
		values[i] = b.String()
	}
	selectBackdrop := widget.NewSelect(values, func(name string) {
		for _, b := range galleryBackdrops {
			if b.String() == name {
				_, err := desktop.ConfigureWindowVisuals(w, desktop.VisualOptions{Backdrop: b, PrepareAlpha: true})
				if err != nil {
					status.SetText(err.Error())
				}
				break
			}
		}
	})
	if ok {
		cancel := visuals.ObserveVisualStatus(func(s desktop.VisualStatus) {
			status.SetText(fmt.Sprintf("%s → %s · %s · %s", s.Requested, s.Selected, s.Phase, s.Reason))
		})
		w.SetOnClosed(cancel)
		selectBackdrop.SetSelected(initial.String())
	} else {
		selectBackdrop.Disable()
	}
	patches := make([]granio.CanvasObject, 0, 5)
	for i, a := range []uint8{0, 64, 128, 192, 255} {
		r := canvas.NewRectangle(color.NRGBA{R: 32, G: 176, B: 240, A: a})
		r.SetMinSize(granio.NewSize(100, 95))
		r.CornerRadius = 12
		patches = append(patches, container.NewVBox(widget.NewLabel(fmt.Sprintf("Alpha %d%%", i*25)), r))
	}
	text := widget.NewLabel("Transparent показывает окна сзади. System Blur запрашивает системное размытие; при недоступности эффектов окно становится непрозрачным. Mica и Mica Alt используют системный фон Windows. Поместите под галереей цветное окно.")
	text.Wrapping = granio.TextWrapWord
	page := container.NewVBox(container.NewHBox(widget.NewLabel("Материал окна"), selectBackdrop), status, text,
		container.NewGridWithColumns(5, patches...), widget.NewButton("Непрозрачный control", func() {}))
	return container.NewBorder(page, nil, nil, nil, canvas.NewRectangle(color.Transparent))
}
