//go:build !mobile

package widget_test

import (
	"image/color"
	"runtime"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	internalWidget "github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMenu_Layout(t *testing.T) {
	test.NewTempApp(t)

	w := test.NewWindow(canvas.NewRectangle(color.Transparent))
	defer w.Close()
	w.SetPadded(false)
	c := w.Canvas()

	item1 := granio.NewMenuItem("A", nil)
	item2 := granio.NewMenuItem("B (long)", nil)
	sep := granio.NewMenuItemSeparator()
	item3 := granio.NewMenuItem("C", nil)
	subItem1 := granio.NewMenuItem("subitem A", nil)
	subItem1.Checked = true
	subItem2 := granio.NewMenuItem("subitem B", nil)
	subItem2.Checked = true
	subItem2.Icon = theme.InfoIcon()
	subItem3 := granio.NewMenuItem("subitem C (long)", nil)
	subItem3.Icon = theme.MenuIcon()
	subsubItem1 := granio.NewMenuItem("subsubitem A (long)", nil)
	subsubItem1.Icon = theme.FileIcon()
	subsubItem2 := granio.NewMenuItem("subsubitem B", nil)
	subItem3.ChildMenu = granio.NewMenu("", subsubItem1, subsubItem2)
	item3.ChildMenu = granio.NewMenu("", subItem1, subItem2, subItem3)
	item4 := granio.NewMenuItem("D", nil)
	subItem4a := granio.NewMenuItem("a", nil)
	subItem4a.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyA, Modifier: granio.KeyModifierControl}
	subItem4b := granio.NewMenuItem("b", nil)
	subItem4b.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyBackspace, Modifier: granio.KeyModifierAlt}
	subItem4c := granio.NewMenuItem("c", nil)
	subItem4c.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyDelete, Modifier: granio.KeyModifierSuper}
	subItem4d := granio.NewMenuItem("d", nil)
	subItem4d.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyDown, Modifier: granio.KeyModifierShift}
	subItem4e := granio.NewMenuItem("e", nil)
	subItem4e.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyEnd, Modifier: granio.KeyModifierControl | granio.KeyModifierAlt}
	subItem4f := granio.NewMenuItem("f", nil)
	subItem4f.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyEnter, Modifier: granio.KeyModifierControl | granio.KeyModifierSuper}
	subItem4g := granio.NewMenuItem("g", nil)
	subItem4g.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyEscape, Modifier: granio.KeyModifierControl | granio.KeyModifierShift}
	subItem4h := granio.NewMenuItem("h", nil)
	subItem4h.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyHome, Modifier: granio.KeyModifierAlt | granio.KeyModifierSuper}
	subItem4i := granio.NewMenuItem("i", nil)
	subItem4i.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyLeft, Modifier: granio.KeyModifierAlt | granio.KeyModifierShift}
	subItem4j := granio.NewMenuItem("j", nil)
	subItem4j.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyPageDown, Modifier: granio.KeyModifierSuper | granio.KeyModifierShift}
	subItem4k := granio.NewMenuItem("k", nil)
	subItem4k.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyPageUp, Modifier: granio.KeyModifierControl | granio.KeyModifierAlt | granio.KeyModifierSuper}
	subItem4l := granio.NewMenuItem("l", nil)
	subItem4l.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyReturn, Modifier: granio.KeyModifierControl | granio.KeyModifierAlt | granio.KeyModifierShift}
	subItem4m := granio.NewMenuItem("m", nil)
	subItem4m.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyRight, Modifier: granio.KeyModifierControl | granio.KeyModifierSuper | granio.KeyModifierShift}
	subItem4n := granio.NewMenuItem("n", nil)
	subItem4n.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeySpace, Modifier: granio.KeyModifierControl | granio.KeyModifierAlt | granio.KeyModifierSuper | granio.KeyModifierShift}
	subItem4o := granio.NewMenuItem("o", nil)
	subItem4o.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyTab, Modifier: granio.KeyModifierControl}
	subItem4p := granio.NewMenuItem("p", nil)
	subItem4p.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyUp, Modifier: granio.KeyModifierControl}
	subItem4q := granio.NewMenuItem("q", nil)
	subItem4q.Shortcut = &desktop.CustomShortcut{KeyName: granio.KeyF6}
	item4.ChildMenu = granio.NewMenu("", subItem4a, subItem4b, subItem4c, subItem4d, subItem4e, subItem4f, subItem4g, subItem4h, subItem4i, subItem4j, subItem4k, subItem4l, subItem4m, subItem4n, subItem4o, subItem4p, subItem4q)

	menu := granio.NewMenu("", item1, sep, item2, item3, sep, item4)

	var shortcutsMasterPrefixPath string
	if runtime.GOOS == "darwin" {
		shortcutsMasterPrefixPath = "menu/desktop/layout_shortcuts_darwin"
	} else {
		shortcutsMasterPrefixPath = "menu/desktop/layout_shortcuts_other"
	}

	for name, tt := range map[string]struct {
		windowSize         granio.Size
		menuPos            granio.Position
		mousePositions     []granio.Position
		want               string
		wantImage          string
		wantTestThemeImage string
	}{
		"normal": {
			windowSize: granio.NewSize(500, 300),
			menuPos:    granio.NewPos(10, 10),
			want:       "menu/desktop/layout_normal.xml",
		},
		"normal with submenus": {
			windowSize: granio.NewSize(500, 300),
			menuPos:    granio.NewPos(10, 10),
			mousePositions: []granio.Position{
				granio.NewPos(32, 103),
				granio.NewPos(102, 173),
			},
			want: "menu/desktop/layout_normal_with_submenus.xml",
		},
		"background of active submenu parents resets if sibling is focused": {
			windowSize: granio.NewSize(500, 300),
			menuPos:    granio.NewPos(10, 10),
			mousePositions: []granio.Position{
				granio.NewPos(30, 100),  // open submenu
				granio.NewPos(100, 170), // open subsubmenu
				granio.NewPos(300, 170), // focus subsubmenu item
				granio.NewPos(30, 60),   // focus sibling of submenu parent
			},
			want: "menu/desktop/layout_background_reset.xml",
		},
		"no space on right side for submenu": {
			windowSize: granio.NewSize(500, 300),
			menuPos:    granio.NewPos(410, 10),
			mousePositions: []granio.Position{
				granio.NewPos(432, 103), // open submenu
				granio.NewPos(302, 173), // open subsubmenu
			},
			want: "menu/desktop/layout_no_space_on_right.xml",
		},
		"no space on left & right side for submenu": {
			windowSize: granio.NewSize(200, 300),
			menuPos:    granio.NewPos(10, 10),
			mousePositions: []granio.Position{
				granio.NewPos(32, 103),  // open submenu
				granio.NewPos(102, 173), // open subsubmenu
			},
			want: "menu/desktop/layout_no_space_on_both_sides.xml",
		},
		"window too short for submenu": {
			windowSize: granio.NewSize(500, 150),
			menuPos:    granio.NewPos(10, 10),
			mousePositions: []granio.Position{
				granio.NewPos(30, 100),  // open submenu
				granio.NewPos(100, 130), // open subsubmenu
			},
			want: "menu/desktop/layout_window_too_short_for_submenu.xml",
		},
		"theme change": {
			windowSize:         granio.NewSize(500, 300),
			menuPos:            granio.NewPos(10, 10),
			want:               "menu/desktop/layout_theme_changed.xml",
			wantImage:          "menu/desktop/layout_normal.png",
			wantTestThemeImage: "menu/desktop/layout_theme_changed.png",
		},
		"window too short for menu": {
			windowSize: granio.NewSize(100, 50),
			menuPos:    granio.NewPos(10, 10),
			want:       "menu/desktop/layout_window_too_short.xml",
		},
		"menu with shortcuts": {
			windowSize: granio.NewSize(450, 800),
			menuPos:    granio.NewPos(10, 10),
			mousePositions: []granio.Position{
				granio.NewPos(32, 143), // open submenu
			},
			want:               shortcutsMasterPrefixPath + ".xml",
			wantImage:          shortcutsMasterPrefixPath + ".png",
			wantTestThemeImage: shortcutsMasterPrefixPath + "_theme_changed.png",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if name == "menu with shortcuts" && (runtime.GOOS == "windows" || runtime.GOOS == "darwin") {
				return // macOS and Windows are off-by-a-fraction for symbol font TODO find out why
			}
			w.Resize(tt.windowSize)
			m := widget.NewMenu(menu)
			o := internalWidget.NewOverlayContainer(m, c, nil)
			c.Overlays().Add(o)
			defer c.Overlays().Remove(o)
			m.Move(tt.menuPos)
			m.Resize(m.MinSize())
			for _, pos := range tt.mousePositions {
				test.MoveMouse(c, pos)
			}
			test.AssertRendersToMarkup(t, tt.want, w.Canvas())
			if tt.wantImage != "" {
				test.AssertImageMatches(t, tt.wantImage, c.Capture())
			}
			if tt.wantTestThemeImage != "" {
				test.WithTestTheme(t, func() {
					test.AssertImageMatches(t, tt.wantTestThemeImage, c.Capture())
				})
			}
		})
	}
}

func TestMenu_Scrolling(t *testing.T) {
	test.NewTempApp(t)

	w := test.NewWindow(canvas.NewRectangle(color.Transparent))
	defer w.Close()
	w.SetPadded(false)
	c := w.Canvas()

	menu := granio.NewMenu(
		"",
		granio.NewMenuItem("A", nil),
		granio.NewMenuItem("B", nil),
		granio.NewMenuItem("C", nil),
		granio.NewMenuItem("D", nil),
		granio.NewMenuItem("E", nil),
		granio.NewMenuItem("F", nil),
	)

	// 100x100
	w.Resize(granio.NewSize(100, 100))
	m := widget.NewMenu(menu)
	o := internalWidget.NewOverlayContainer(m, c, nil)
	c.Overlays().Add(o)
	defer c.Overlays().Remove(o)
	m.Move(granio.NewPos(10, 10))
	m.Resize(m.MinSize())
	maxScrollDistance := m.MinSize().Height - 90
	test.AssertRendersToMarkup(t, "menu/desktop/scroll_top.xml", w.Canvas())

	test.Scroll(c, granio.NewPos(20, 20), 0, -50)
	test.AssertRendersToMarkup(t, "menu/desktop/scroll_middle.xml", w.Canvas())

	test.Scroll(c, granio.NewPos(20, 20), 0, -maxScrollDistance)
	test.AssertRendersToMarkup(t, "menu/desktop/scroll_bottom.xml", w.Canvas())

	test.Scroll(c, granio.NewPos(20, 20), 0, maxScrollDistance-50)
	test.AssertRendersToMarkup(t, "menu/desktop/scroll_middle.xml", w.Canvas())

	test.Scroll(c, granio.NewPos(20, 20), 0, 50)
	test.AssertRendersToMarkup(t, "menu/desktop/scroll_top.xml", w.Canvas())
}

func TestMenu_TraverseMenu(t *testing.T) {
	test.NewTempApp(t)

	w := granio.CurrentApp().NewWindow("")
	defer w.Close()
	w.SetPadded(false)
	c := w.Canvas()

	itemWithChild := granio.NewMenuItem("Bar", nil)
	itemWithChild.ChildMenu = granio.NewMenu(
		"",
		granio.NewMenuItem("SubA", nil),
		granio.NewMenuItem("SubB", nil),
	)
	m := widget.NewMenu(granio.NewMenu(
		"",
		granio.NewMenuItem("Foo", nil),
		granio.NewMenuItemSeparator(),
		itemWithChild,
		granio.NewMenuItemSeparator(),
		granio.NewMenuItem("Baz", nil),
	))
	w.SetContent(internalWidget.NewOverlayContainer(m, c, nil))
	// + 4,5 for canvas’ safe area
	w.Resize(m.MinSize().AddWidthHeight(4, 5))
	m.Resize(m.MinSize())

	// going all the way down …
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_initial.xml", c)

	m.ActivateNext()
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_first_active.xml", c)

	m.ActivateNext()
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_second_active.xml", c)

	m.ActivateNext()
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_third_active.xml", c)

	m.ActivateNext()
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_third_active.xml", c, "does not wrap around if last item is already active")

	// … and up again
	m.ActivatePrevious()
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_second_active.xml", c)

	m.ActivatePrevious()
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_first_active.xml", c)

	m.ActivatePrevious()
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_first_active.xml", c, "does not wrap around if on top")

	// activate a submenu (show and activate first item)
	m.ActivateNext()
	assert.True(t, m.ActivateLastSubmenu())
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_submenu_first_active.xml", c)

	assert.False(t, m.ActivateLastSubmenu())
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_submenu_first_active.xml", c, "does nothing if there is no submenu at the last active item")

	// traversing through items of opened submenu
	m.ActivateNext()
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_submenu_second_active.xml", c)

	m.ActivateNext()
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_submenu_second_active.xml", c, "does not wrap around if last item is already active")

	m.ActivatePrevious()
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_submenu_first_active.xml", c)

	m.ActivatePrevious()
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_submenu_first_active.xml", c, "does not wrap around if on top")

	// closing an open submenu
	assert.True(t, m.DeactivateLastSubmenu())
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_second_active.xml", c)

	assert.False(t, m.DeactivateLastSubmenu())
	test.AssertRendersToMarkup(t, "menu/desktop/traverse_second_active.xml", c, "does nothing if there is no submenu opened")
}

func TestMenu_TriggerTraversedMenu(t *testing.T) {
	var triggered string
	var dismissed bool
	setupMenu := func() *widget.Menu {
		triggered = ""
		dismissed = false
		itemWithChild := granio.NewMenuItem("Bar", func() { triggered = "2nd" })
		itemWithChild.ChildMenu = granio.NewMenu(
			"",
			granio.NewMenuItem("SubA", func() { triggered = "1st sub" }),
			granio.NewMenuItem("SubB", nil),
			granio.NewMenuItem("SubC", func() { triggered = "3rd sub" }),
		)
		m := widget.NewMenu(granio.NewMenu(
			"",
			granio.NewMenuItem("Foo", func() { triggered = "1st" }),
			granio.NewMenuItemSeparator(),
			itemWithChild,
			granio.NewMenuItemSeparator(),
			granio.NewMenuItem("Baz", func() { triggered = "3rd" }),
		))
		m.OnDismiss = func() { dismissed = true }
		w := granio.CurrentApp().NewWindow("")
		t.Cleanup(w.Close)
		w.SetContent(internalWidget.NewOverlayContainer(m, w.Canvas(), nil))
		return m
	}

	t.Run("without active item", func(t *testing.T) {
		m := setupMenu()
		m.TriggerLast()
		assert.Equal(t, "", triggered)
		assert.True(t, dismissed)
	})
	t.Run("first item in submenu", func(t *testing.T) {
		m := setupMenu()
		m.ActivateNext()
		m.ActivateNext()
		require.True(t, m.ActivateLastSubmenu())
		m.TriggerLast()
		assert.Equal(t, "1st sub", triggered)
		assert.True(t, dismissed)
	})
	t.Run("last item in submenu", func(t *testing.T) {
		m := setupMenu()
		m.ActivateNext()
		m.ActivateNext()
		require.True(t, m.ActivateLastSubmenu())
		m.ActivateNext()
		m.ActivateNext()
		m.TriggerLast()
		assert.Equal(t, "3rd sub", triggered)
		assert.True(t, dismissed)
	})
	t.Run("item in menu", func(t *testing.T) {
		m := setupMenu()
		m.ActivateNext()
		m.ActivateNext()
		m.ActivateNext()
		m.TriggerLast()
		assert.Equal(t, "3rd", triggered)
		assert.True(t, dismissed)
	})
	t.Run("item with (closed) submenu", func(t *testing.T) {
		m := setupMenu()
		m.ActivateNext()
		m.ActivateNext()
		m.TriggerLast()
		assert.Equal(t, "2nd", triggered)
		assert.True(t, dismissed)
	})
	t.Run("item without action", func(t *testing.T) {
		m := setupMenu()
		m.ActivateNext()
		m.ActivateNext()
		require.True(t, m.ActivateLastSubmenu())
		m.ActivateNext()
		m.TriggerLast()
		assert.Equal(t, "", triggered)
		assert.True(t, dismissed)
	})
}
