//go:build mobile

package widget_test

import (
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	internalWidget "github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
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
	menu := granio.NewMenu("", item1, sep, item2, item3)

	for name, tt := range map[string]struct {
		windowSize   granio.Size
		menuPos      granio.Position
		tapPositions []granio.Position
		useTestTheme bool
		want         string
	}{
		"normal": {
			windowSize: granio.NewSize(500, 300),
			menuPos:    granio.NewPos(10, 10),
			want:       "menu/mobile/layout_normal.xml",
		},
		"normal with submenus": {
			windowSize: granio.NewSize(500, 300),
			menuPos:    granio.NewPos(10, 10),
			tapPositions: []granio.Position{
				granio.NewPos(32, 103),
				granio.NewPos(102, 173),
			},
			want: "menu/mobile/layout_normal_with_submenus.xml",
		},
		"background of active submenu parents resets if sibling is hovered": {
			windowSize: granio.NewSize(500, 300),
			menuPos:    granio.NewPos(10, 10),
			tapPositions: []granio.Position{
				granio.NewPos(30, 100),  // open submenu
				granio.NewPos(100, 170), // open subsubmenu
				granio.NewPos(300, 170), // hover subsubmenu item
				granio.NewPos(30, 60),   // hover sibling of submenu parent
			},
			want: "menu/mobile/layout_background_reset.xml",
		},
		"no space on right side for submenu": {
			windowSize: granio.NewSize(500, 300),
			menuPos:    granio.NewPos(410, 10),
			tapPositions: []granio.Position{
				granio.NewPos(432, 103), // open submenu
				granio.NewPos(302, 173), // open subsubmenu
			},
			want: "menu/mobile/layout_no_space_on_right.xml",
		},
		"no space on left & right side for submenu": {
			windowSize: granio.NewSize(200, 300),
			menuPos:    granio.NewPos(10, 10),
			tapPositions: []granio.Position{
				granio.NewPos(32, 103),  // open submenu
				granio.NewPos(102, 173), // open subsubmenu
			},
			want: "menu/mobile/layout_no_space_on_both_sides.xml",
		},
		"window too short for submenu": {
			windowSize: granio.NewSize(500, 150),
			menuPos:    granio.NewPos(10, 10),
			tapPositions: []granio.Position{
				granio.NewPos(30, 100),  // open submenu
				granio.NewPos(100, 130), // open subsubmenu
			},
			want: "menu/mobile/layout_window_too_short_for_submenu.xml",
		},
		"theme change": {
			windowSize:   granio.NewSize(500, 300),
			menuPos:      granio.NewPos(10, 10),
			useTestTheme: true,
			want:         "menu/mobile/layout_theme_changed.xml",
		},
		"window too short for menu": {
			windowSize: granio.NewSize(100, 50),
			menuPos:    granio.NewPos(10, 10),
			want:       "menu/mobile/layout_window_too_short.xml",
		},
	} {
		t.Run(name, func(t *testing.T) {
			w.Resize(tt.windowSize)
			m := widget.NewMenu(menu)
			o := internalWidget.NewOverlayContainer(m, c, nil)
			c.Overlays().Add(o)
			defer c.Overlays().Remove(o)
			m.Move(tt.menuPos)
			m.Resize(m.MinSize())
			for _, pos := range tt.tapPositions {
				test.TapCanvas(c, pos)
			}
			test.AssertRendersToMarkup(t, tt.want, w.Canvas())
			if tt.useTestTheme {
				test.AssertImageMatches(t, "menu/mobile/layout_normal.png", c.Capture())
				test.WithTestTheme(t, func() {
					test.AssertImageMatches(t, "menu/mobile/layout_theme_changed.png", c.Capture())
				})
			}
		})
	}
}

func TestMenu_Dragging(t *testing.T) {
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

	w.Resize(granio.NewSize(100, 100))
	m := widget.NewMenu(menu)
	o := internalWidget.NewOverlayContainer(m, c, nil)
	c.Overlays().Add(o)
	defer c.Overlays().Remove(o)
	m.Move(granio.NewPos(10, 10))
	m.Resize(m.MinSize())
	maxDragDistance := m.MinSize().Height - 90
	test.AssertRendersToMarkup(t, "menu/mobile/drag_top.xml", w.Canvas())

	test.Drag(c, granio.NewPos(20, 20), 0, -50)
	test.AssertRendersToMarkup(t, "menu/mobile/drag_middle.xml", w.Canvas())

	test.Drag(c, granio.NewPos(20, 20), 0, -maxDragDistance)
	test.AssertRendersToMarkup(t, "menu/mobile/drag_bottom.xml", w.Canvas())

	test.Drag(c, granio.NewPos(20, 20), 0, maxDragDistance-50)
	test.AssertRendersToMarkup(t, "menu/mobile/drag_middle.xml", w.Canvas())

	test.Drag(c, granio.NewPos(20, 20), 0, 50)
	test.AssertRendersToMarkup(t, "menu/mobile/drag_top.xml", w.Canvas())
}
