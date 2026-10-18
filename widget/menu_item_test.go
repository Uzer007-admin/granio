package widget

import (
	"bytes"
	_ "embed"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	internalWidget "github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"
)

//go:embed testdata/fyne.png
var iconData []byte

func TestMenuItem_Disabled(t *testing.T) {
	t.Run("can render disabled item", func(t *testing.T) {
		i := granio.NewMenuItem("Disabled", func() {})
		m := granio.NewMenu("top", []*granio.MenuItem{i}...)
		i.Disabled = true
		w := newMenuItem(i, NewMenu(m))
		r := cache.Renderer(w)

		assert.Equal(t, theme.Color(theme.ColorNameDisabled), r.(*menuItemRenderer).text.Color)
	})

	t.Run("should apply themeing to disabled item SVG icon", func(t *testing.T) {
		i := granio.NewMenuItem("Disabled", func() {})
		m := granio.NewMenu("top", i)
		i.Disabled = true
		i.Icon = theme.AccountIcon()
		w := newMenuItem(i, NewMenu(m))
		r := cache.Renderer(w)

		assert.Equal(t, theme.NewDisabledResource(theme.AccountIcon()), r.(*menuItemRenderer).icon.Resource)
	})

	t.Run("should not apply themeing to disabled item with PNG icon", func(t *testing.T) {
		i := granio.NewMenuItem("Disabled", func() {})
		m := granio.NewMenu("top", i)
		i.Disabled = true
		i.Icon = granio.NewStaticResource("fyne.png", iconData)
		w := newMenuItem(i, NewMenu(m))
		r := cache.Renderer(w)

		assert.Equal(t, i.Icon, r.(*menuItemRenderer).icon.Resource)
	})

	t.Run("should not log error when rendering disabled menu item with PNG icon", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer func() {
			log.SetOutput(os.Stderr)
		}()

		test.NewTempApp(t)

		w := granio.CurrentApp().NewWindow("")
		defer w.Close()
		w.SetPadded(false)
		c := w.Canvas()

		item := granio.NewMenuItem("Foo", func() {})
		item.Icon = granio.NewStaticResource("fyne.png", iconData)
		item.Disabled = true
		m := NewMenu(granio.NewMenu("", item))
		size := m.MinSize()
		w.Resize(size.Add(granio.NewSize(10, 10)))
		m.Resize(size)
		o := internalWidget.NewOverlayContainer(m, c, func() {})
		w.SetContent(o)

		assert.Equal(t, "", buf.String()) // should not log an error
	})

	t.Run("should not log error when rendering disabled menu item with SVG icon", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer func() {
			log.SetOutput(os.Stderr)
		}()

		test.NewTempApp(t)

		w := granio.CurrentApp().NewWindow("")
		defer w.Close()
		w.SetPadded(false)
		c := w.Canvas()

		item := granio.NewMenuItem("Foo", func() {})
		item.Icon = theme.AccountIcon()
		item.Disabled = true
		m := NewMenu(granio.NewMenu("", item))
		size := m.MinSize()
		w.Resize(size.Add(granio.NewSize(10, 10)))
		m.Resize(size)
		o := internalWidget.NewOverlayContainer(m, c, func() {})
		w.SetContent(o)

		assert.Equal(t, "", buf.String()) // should not log an error
	})
}

func TestMenuItem_Refresh(t *testing.T) {
	i := granio.NewMenuItem("before", func() {})
	m := granio.NewMenu("top", i)
	w := newMenuItem(i, NewMenu(m))
	r := cache.Renderer(w).(*menuItemRenderer)

	assert.Equal(t, "before", r.text.Text)

	i.Label = "after"
	w.Refresh()

	assert.Equal(t, "after", r.text.Text)
}
