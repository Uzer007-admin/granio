package widget_test

import (
	"image/color"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"
)

func TestScrollContainer_Theme(t *testing.T) {
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(granio.NewSize(250, 250))
	scroll := widget.NewScroll(rect)

	w := test.NewTempWindow(t, scroll)
	w.SetPadded(false)
	w.Resize(granio.NewSize(100, 100))
	test.AssertImageMatches(t, "scroll/theme_initial.png", w.Canvas().Capture())

	test.WithTestTheme(t, func() {
		time.Sleep(100 * time.Millisecond)
		scroll.Refresh()
		test.AssertImageMatches(t, "scroll/theme_changed.png", w.Canvas().Capture())
	})
}

func TestScrollContainer_ThemeOverride(t *testing.T) {
	rect := canvas.NewRectangle(color.Transparent)
	rect.SetMinSize(granio.NewSize(250, 250))
	scroll := widget.NewScroll(rect)
	scroll.Resize(granio.NewSize(100, 100))

	w := test.NewTempWindow(t, scroll)
	w.SetPadded(false)
	w.Resize(granio.NewSize(100, 100))
	test.ApplyTheme(t, test.NewTheme())
	test.AssertImageMatches(t, "scroll/theme_changed.png", w.Canvas().Capture())

	normal := test.Theme()
	bg := canvas.NewRectangle(normal.Color(theme.ColorNameBackground, theme.VariantDark))
	w.SetContent(container.NewStack(bg, container.NewThemeOverride(scroll, normal)))
	w.Resize(granio.NewSize(100, 100))
	// TODO why is this off by a 1bit RGB difference?
	// test.AssertImageMatches(t, "scroll/theme_initial.png", w.Canvas().Capture())
}
