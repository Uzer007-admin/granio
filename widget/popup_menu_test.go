package widget_test

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func TestPopUpMenu_Move(t *testing.T) {
	m, w := setupPopUpMenuTest(t)
	c := w.Canvas()

	m.Show()
	test.AssertRendersToMarkup(t, "popup_menu/shown.xml", c)

	m.Move(granio.NewPos(20, 20))
	test.AssertRendersToMarkup(t, "popup_menu/moved.xml", c)

	m.Move(granio.NewPos(190, 10))
	test.AssertRendersToMarkup(t, "popup_menu/no_space_right.xml", c)

	m.Move(granio.NewPos(10, 190))
	test.AssertRendersToMarkup(t, "popup_menu/no_space_down.xml", c)
}

func TestPopUpMenu_Resize(t *testing.T) {
	m, w := setupPopUpMenuTest(t)
	c := w.Canvas()

	m.ShowAtPosition(granio.NewPos(10, 10))
	test.AssertRendersToMarkup(t, "popup_menu/shown_at_pos.xml", c)

	m.Resize(m.Size().Add(granio.NewSize(10, 10)))
	test.AssertRendersToMarkup(t, "popup_menu/grown.xml", c)

	largeSize := c.Size().Add(granio.NewSize(10, 10))
	m.Resize(largeSize)
	test.AssertRendersToMarkup(t, "popup_menu/canvas_too_small.xml", c)
	_, areaSize := c.InteractiveArea()
	assert.Equal(t, granio.NewSize(largeSize.Width, areaSize.Height), m.Size(), "width is larger than canvas; height is limited by canvas (menu scrolls)")

	m.Resize(granio.NewSize(largeSize.Width, 10))
	assert.Equal(t, granio.NewSize(largeSize.Width, m.MinSize().Height), m.Size(), "less height than the content needs grows back to the minimum")
}

func TestPopUpMenu_Show(t *testing.T) {
	m, w := setupPopUpMenuTest(t)
	c := w.Canvas()

	test.AssertRendersToMarkup(t, "popup_menu/hidden.xml", c)

	m.Show()
	test.AssertRendersToMarkup(t, "popup_menu/shown.xml", c)
}

func TestPopUpMenu_ShowAtPosition(t *testing.T) {
	m, w := setupPopUpMenuTest(t)
	c := w.Canvas()

	test.AssertRendersToMarkup(t, "popup_menu/hidden.xml", c)

	m.ShowAtPosition(granio.NewPos(10, 10))
	test.AssertRendersToMarkup(t, "popup_menu/shown_at_pos.xml", c)

	m.Hide()
	test.AssertRendersToMarkup(t, "popup_menu/hidden.xml", c)

	m.ShowAtPosition(granio.NewPos(190, 10))
	test.AssertRendersToMarkup(t, "popup_menu/no_space_right.xml", c)

	m.Hide()
	test.AssertRendersToMarkup(t, "popup_menu/hidden.xml", c)

	m.ShowAtPosition(granio.NewPos(10, 190))
	test.AssertRendersToMarkup(t, "popup_menu/no_space_down.xml", c)

	m.Hide()
	test.AssertRendersToMarkup(t, "popup_menu/hidden.xml", c)

	m.ShowAtPosition(granio.NewPos(10, 10))
	menuSize := c.Size().Add(granio.NewSize(10, 10))
	m.Resize(menuSize)
	test.AssertRendersToMarkup(t, "popup_menu/canvas_too_small.xml", c)
	_, areaSize := c.InteractiveArea()
	assert.Equal(t, granio.NewSize(menuSize.Width, areaSize.Height), m.Size(), "width is larger than canvas; height is limited by canvas (menu scrolls)")
}

func setupPopUpMenuTest(t *testing.T) (*widget.PopUpMenu, granio.Window) {
	test.NewApp()

	w := test.NewTempWindow(t, canvas.NewRectangle(color.NRGBA{G: 150, B: 150, A: 255}))
	t.Cleanup(func() { test.NewApp() })

	w.Resize(granio.NewSize(200, 200))
	m := widget.NewPopUpMenu(granio.NewMenu(
		"",
		granio.NewMenuItem("Option A", nil),
		granio.NewMenuItem("Option B", nil),
	), w.Canvas())
	return m, w
}

func setupPopUpMenuWithSubmenusTest(t *testing.T, callback func(string)) (*widget.PopUpMenu, granio.Window) {
	test.NewApp()

	w := test.NewTempWindow(t, canvas.NewRectangle(color.NRGBA{G: 150, B: 150, A: 255}))
	t.Cleanup(func() { test.NewApp() })

	w.Resize(granio.NewSize(800, 600))
	itemA := granio.NewMenuItem("Option A", func() { callback("Option A") })
	itemB := granio.NewMenuItem("Option B", func() { callback("Option B") })
	itemBA := granio.NewMenuItem("Sub Option A", func() { callback("Sub Option A") })
	itemBB := granio.NewMenuItem("Sub Option B", func() { callback("Sub Option B") })
	itemBBA := granio.NewMenuItem("Sub Sub Option A", func() { callback("Sub Sub Option A") })
	itemBBB := granio.NewMenuItem("Sub Sub Option B", func() { callback("Sub Sub Option B") })
	itemB.ChildMenu = granio.NewMenu("", itemBA, itemBB)
	itemBB.ChildMenu = granio.NewMenu("", itemBBA, itemBBB)
	m := widget.NewPopUpMenu(granio.NewMenu("", itemA, itemB), w.Canvas())
	return m, w
}
