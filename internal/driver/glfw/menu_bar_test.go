//go:build !mobile && !no_glfw && !windows

package glfw

import (
	"image"
	"strconv"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/software"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/widget"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMenuBar(t *testing.T) {
	runOnMain(func() {
		test.NewTempApp(t)
	})

	var lastAction string

	m1i3i3i1 := granio.NewMenuItem("Old 1", func() { lastAction = "old 1" })
	m1i3i3i2 := granio.NewMenuItem("Old 2", func() { lastAction = "old 2" })
	m1i3i1 := granio.NewMenuItem("File 1", func() { lastAction = "file 1" })
	m1i3i2 := granio.NewMenuItem("File 2", func() { lastAction = "file 2" })
	m1i3i3 := granio.NewMenuItem("Older", nil)
	m1i3i3.ChildMenu = granio.NewMenu("", m1i3i3i1, m1i3i3i2)
	m1i1 := granio.NewMenuItem("New", func() { lastAction = "new" })
	m1i2 := granio.NewMenuItem("Open", func() { lastAction = "open" })
	m1i3 := granio.NewMenuItem("Recent", nil)
	m1i3.ChildMenu = granio.NewMenu("", m1i3i1, m1i3i2, m1i3i3)
	// TODO: remove useless separators: trailing, leading & double
	// m1 := granio.NewMenu("File", m1i1, m1i2, fyne.newMenuItemSeparator(), m1i3)
	m1 := granio.NewMenu("File", m1i1, m1i2, m1i3)

	m2i1 := granio.NewMenuItem("Copy", func() { lastAction = "copy" })
	m2i2 := granio.NewMenuItem("Paste", func() { lastAction = "paste" })
	m2 := granio.NewMenu("Edit", m2i1, m2i2)

	m3i1 := granio.NewMenuItem("Help!", func() { lastAction = "help" })
	m3 := granio.NewMenu("Help", m3i1)

	menu := granio.NewMainMenu(m1, m2, m3)

	t.Run("mouse control and basic behaviour", func(t *testing.T) {
		w := test.NewWindow(nil)
		defer w.Close()
		w.SetPadded(false)
		w.Resize(granio.NewSize(300, 300))
		c := w.Canvas()

		menuBar := NewMenuBar(menu, c)
		themeCounter := 0
		button := newNotFocusableButton("Button", func() {
			switch themeCounter % 2 {
			case 0:
				test.ApplyTheme(t, test.NewTheme())
			case 1:
				test.ApplyTheme(t, test.Theme())
			}
			themeCounter++
		})
		container := container.NewWithoutLayout(button, menuBar)
		runOnMain(func() {
			w.SetContent(container)
			w.Resize(granio.NewSize(300, 300))

			button.Resize(button.MinSize())
			button.Move(granio.NewPos(100, 50))
			menuBar.Resize(granio.NewSize(300, 0).Max(menuBar.MinSize()))
		})

		buttonPos := granio.NewPos(110, 60)
		fileMenuPos := granio.NewPos(20, 10)
		fileNewPos := granio.NewPos(20, 55)
		fileOpenPos := granio.NewPos(20, 75)
		fileRecentPos := granio.NewPos(20, 110)
		fileRecentOlderPos := granio.NewPos(120, 190)
		fileRecentOlderOld1Pos := granio.NewPos(200, 190)
		editMenuPos := granio.NewPos(70, 10)
		helpMenuPos := granio.NewPos(120, 10)
		type action struct {
			typ string
			pos granio.Position
		}
		type step struct {
			actions    []action
			wantImage  string
			wantAction string
		}
		for name, tt := range map[string]struct {
			steps []step
		}{
			"switch theme": {
				steps: []step{
					{
						actions:   []action{{"move", buttonPos}},
						wantImage: "menu_bar_hovered_content.png",
					},
					{
						actions:   []action{{"tap", buttonPos}},
						wantImage: "menu_bar_hovered_content_test_theme.png",
					},
					{
						actions:   []action{{"tap", buttonPos}},
						wantImage: "menu_bar_hovered_content.png",
					},
				},
			},
			"activate and deactivate menu": {
				[]step{
					{
						actions:   []action{{"move", fileMenuPos}},
						wantImage: "menu_bar_inactive_file.png",
					},
					{
						actions:   []action{{"tap", fileMenuPos}},
						wantImage: "menu_bar_active_file.png",
					},
					{
						actions:   []action{{"tap", fileMenuPos}},
						wantImage: "menu_bar_inactive_file.png",
					},
					{
						actions:   []action{{"move", buttonPos}},
						wantImage: "menu_bar_hovered_content.png",
					},
				},
			},
			"active menu deactivates content": {
				[]step{
					{
						actions:   []action{{"move", fileMenuPos}},
						wantImage: "menu_bar_inactive_file.png",
					},
					{
						actions:   []action{{"tap", fileMenuPos}},
						wantImage: "menu_bar_active_file.png",
					},
					{
						actions:   []action{{"move", buttonPos}},
						wantImage: "menu_bar_content_not_hoverable_with_active_menu.png",
					},
					{
						// TODO: should be hovered over content here
						// -> is this canvas logic?
						// -> it would be for overlays (probably the same issue present)
						// -> it would not be for current menu implementation (menu is not an overlay, canvas does not know about activation state)
						actions:   []action{{"tap", buttonPos}},
						wantImage: "menu_bar_tap_content_with_active_menu_does_not_trigger_action_but_dismisses_menu.png",
					},
					{
						// menu bar is inactive again (menu not shown at hover)
						actions:   []action{{"move", fileMenuPos}},
						wantImage: "menu_bar_inactive_file.png",
					},
					{
						actions:   []action{{"move", buttonPos}},
						wantImage: "menu_bar_hovered_content.png",
					},
				},
			},
			"menu action File->New": {
				[]step{
					{
						actions:   []action{{"tap", fileMenuPos}},
						wantImage: "menu_bar_active_file.png",
					},
					{
						actions:   []action{{"move", fileNewPos}},
						wantImage: "menu_bar_hovered_file_new.png",
					},
					{
						actions:    []action{{"tap", fileNewPos}},
						wantAction: "new",
						wantImage:  "menu_bar_initial.png",
					},
				},
			},
			"menu action File->Open": {
				[]step{
					{
						actions: []action{
							{"tap", fileMenuPos},
							{"move", fileOpenPos},
						},
						wantImage: "menu_bar_hovered_file_open.png",
					},
					{
						actions:    []action{{"tap", fileOpenPos}},
						wantAction: "open",
						wantImage:  "menu_bar_initial.png",
					},
				},
			},
			"menu action File->Recent->Older->Old 1": {
				[]step{
					{
						actions: []action{
							{"tap", fileMenuPos},
							{"move", fileRecentPos},
						},
						wantImage: "menu_bar_hovered_file_recent.png",
					},
					{
						actions:   []action{{"move", fileRecentOlderPos}},
						wantImage: "menu_bar_hovered_file_recent_older.png",
					},
					{
						actions:   []action{{"move", fileRecentOlderOld1Pos}},
						wantImage: "menu_bar_hovered_file_recent_older_old1.png",
					},
					{
						actions:    []action{{"tap", fileRecentOlderOld1Pos}},
						wantAction: "old 1",
						wantImage:  "menu_bar_initial.png",
					},
				},
			},
			"move mouse outside does not hide menu": {
				[]step{
					{
						actions: []action{
							{"tap", fileMenuPos},
							{"move", fileRecentPos},
							{"move", fileRecentOlderPos},
							{"move", fileRecentOlderOld1Pos},
						},
						wantImage: "menu_bar_hovered_file_recent_older_old1.png",
					},
					{
						actions:   []action{{"move", buttonPos}},
						wantImage: "menu_bar_hovered_file_recent_older.png",
					},
				},
			},
			"hover other menu item hides previous submenus": {
				[]step{
					{
						actions: []action{
							{"tap", fileMenuPos},
							{"move", fileRecentPos},
							{"move", fileRecentOlderPos},
							{"move", fileRecentOlderOld1Pos},
						},
						wantImage: "menu_bar_hovered_file_recent_older_old1.png",
					},
					{
						actions:   []action{{"move", fileNewPos}},
						wantImage: "menu_bar_hovered_file_new.png",
					},
					{
						actions:   []action{{"move", fileRecentPos}},
						wantImage: "menu_bar_hovered_file_recent.png",
					},
				},
			},
			"hover other menu bar item changes active menu": {
				[]step{
					{
						actions:   []action{{"tap", fileMenuPos}},
						wantImage: "menu_bar_active_file.png",
					},
					{
						actions:   []action{{"move", editMenuPos}},
						wantImage: "menu_bar_active_edit.png",
					},
					{
						actions:   []action{{"move", helpMenuPos}},
						wantImage: "menu_bar_active_help.png",
					},
				},
			},
			"hover other menu bar item hides previous submenus": {
				[]step{
					{
						actions: []action{
							{"tap", fileMenuPos},
							{"move", fileRecentPos},
							{"move", fileRecentOlderPos},
							{"move", fileRecentOlderOld1Pos},
						},
						wantImage: "menu_bar_hovered_file_recent_older_old1.png",
					},
					{
						actions:   []action{{"move", helpMenuPos}},
						wantImage: "menu_bar_active_help.png",
					},
					{
						actions:   []action{{"move", fileMenuPos}},
						wantImage: "menu_bar_active_file.png",
					},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				var capture image.Image
				runOnMain(func() {
					test.MoveMouse(c, granio.NewPos(0, 0))
					test.TapCanvas(c, granio.NewPos(0, 0))

					capture = c.Capture()
				})
				if test.AssertImageMatches(t, "menu_bar_initial.png", capture) {
					for i, s := range tt.steps {
						t.Run("step "+strconv.Itoa(i+1), func(t *testing.T) {
							lastAction = ""
							for _, a := range s.actions {
								switch a.typ {
								case "move":
									runOnMain(func() {
										test.MoveMouse(c, a.pos)
									})
								case "tap":
									runOnMain(func() {
										test.MoveMouse(c, a.pos)
										test.TapCanvas(c, a.pos)
									})
								}
							}
							var capture2 image.Image
							runOnMain(func() {
								capture2 = c.Capture()
							})
							test.AssertImageMatches(t, s.wantImage, capture2)
							assert.Equal(t, s.wantAction, lastAction, "last action should match expected")
						})
					}
				}
			})
		}
	})

	t.Run("keyboard control", func(t *testing.T) {
		w := test.NewWindow(nil)
		defer w.Close()
		w.SetPadded(false)
		w.Resize(granio.NewSize(300, 300))
		c := w.Canvas()

		menuBar := NewMenuBar(menu, c)
		themeCounter := 0
		button := newNotFocusableButton("Button", func() {
			switch themeCounter % 2 {
			case 0:
				test.ApplyTheme(t, test.NewTheme())
			case 1:
				test.ApplyTheme(t, test.Theme())
			}
			themeCounter++
		})
		container := container.NewWithoutLayout(button, menuBar)
		w.SetContent(container)
		w.Resize(granio.NewSize(300, 300))
		button.Resize(button.MinSize())
		button.Move(granio.NewPos(100, 50))
		menuBar.Resize(granio.NewSize(300, 0).Max(menuBar.MinSize()))

		fileMenuPos := granio.NewPos(20, 10)
		for name, tt := range map[string]struct {
			keys       []granio.KeyName
			wantAction string
		}{
			"traverse_menu_bar_items_right_1": {
				keys: []granio.KeyName{granio.KeyRight},
			},
			"traverse_menu_bar_items_right_2": {
				keys: []granio.KeyName{granio.KeyRight, granio.KeyRight},
			},
			"traverse_menu_bar_items_right_3": {
				keys: []granio.KeyName{granio.KeyRight, granio.KeyRight, granio.KeyRight},
			},
			"traverse_menu_bar_items_left_1": {
				keys: []granio.KeyName{granio.KeyLeft},
			},
			"traverse_menu_bar_items_left_2": {
				keys: []granio.KeyName{granio.KeyLeft, granio.KeyLeft},
			},
			"traverse_menu_bar_items_left_3": {
				keys: []granio.KeyName{granio.KeyLeft, granio.KeyLeft, granio.KeyLeft},
			},
			"traverse_menu_down_1": {
				keys: []granio.KeyName{granio.KeyDown},
			},
			"traverse_menu_down_2": {
				keys: []granio.KeyName{granio.KeyDown, granio.KeyDown},
			},
			"traverse_menu_down_3": {
				keys: []granio.KeyName{granio.KeyDown, granio.KeyDown, granio.KeyDown},
			},
			"traverse_menu_up_1": {
				keys: []granio.KeyName{granio.KeyDown, granio.KeyDown, granio.KeyDown, granio.KeyUp},
			},
			"traverse_menu_up_2": {
				keys: []granio.KeyName{granio.KeyDown, granio.KeyDown, granio.KeyDown, granio.KeyUp, granio.KeyUp},
			},
			"open_submenu_1": {
				keys: []granio.KeyName{granio.KeyDown, granio.KeyDown, granio.KeyDown, granio.KeyRight},
			},
			"open_submenu_2": {
				keys: []granio.KeyName{granio.KeyDown, granio.KeyDown, granio.KeyDown, granio.KeyRight, granio.KeyDown, granio.KeyDown, granio.KeyRight},
			},
			"close_submenu_1": {
				keys: []granio.KeyName{granio.KeyDown, granio.KeyDown, granio.KeyDown, granio.KeyRight, granio.KeyDown, granio.KeyDown, granio.KeyRight, granio.KeyLeft},
			},
			"close_submenu_2": {
				keys: []granio.KeyName{granio.KeyDown, granio.KeyDown, granio.KeyDown, granio.KeyRight, granio.KeyDown, granio.KeyDown, granio.KeyRight, granio.KeyLeft, granio.KeyLeft},
			},
			"trigger_with_enter": {
				keys:       []granio.KeyName{granio.KeyDown, granio.KeyEnter},
				wantAction: "new",
			},
			"trigger_with_return": {
				keys:       []granio.KeyName{granio.KeyDown, granio.KeyDown, granio.KeyReturn},
				wantAction: "open",
			},
			"trigger_with_space": {
				keys:       []granio.KeyName{granio.KeyRight, granio.KeyDown, granio.KeySpace},
				wantAction: "copy",
			},
			"trigger_submenu_item": {
				keys:       []granio.KeyName{granio.KeyDown, granio.KeyDown, granio.KeyDown, granio.KeyRight, granio.KeyDown, granio.KeyDown, granio.KeyRight, granio.KeyReturn},
				wantAction: "old 1",
			},
			"trigger_without_active_item": {
				keys:       []granio.KeyName{granio.KeyEnter},
				wantAction: "",
			},
		} {
			t.Run(name, func(t *testing.T) {
				test.MoveMouse(c, granio.NewPos(0, 0))
				test.TapCanvas(c, granio.NewPos(0, 0))
				test.TapCanvas(c, fileMenuPos) // activate menu
				require.Equal(t, menuBar.Items[0], c.Focused())

				var captured image.Image
				runOnMain(func() {
					captured = c.Capture()
				})
				if test.AssertImageMatches(t, "menu_bar_active_file.png", captured) {
					lastAction = ""
					runOnMain(func() {
						for _, key := range tt.keys {
							c.Focused().TypedKey(&granio.KeyEvent{
								Name: key,
							})
						}
					})
					test.AssertRendersToMarkup(t, "menu_bar_kbdctrl_"+name+".xml", c)
					assert.Equal(t, tt.wantAction, lastAction, "last action should match expected")
				}
			})
		}

		t.Run("moving mouse over unfocused item moves focus", func(t *testing.T) {
			test.MoveMouse(c, granio.NewPos(0, 0))
			test.TapCanvas(c, granio.NewPos(0, 0))
			test.MoveMouse(c, fileMenuPos)
			test.TapCanvas(c, fileMenuPos) // activate menu
			require.Equal(t, menuBar.Items[0], c.Focused())
			c.Focused().TypedKey(&granio.KeyEvent{Name: granio.KeyRight})
			c.Focused().TypedKey(&granio.KeyEvent{Name: granio.KeyRight})
			require.Equal(t, menuBar.Items[2], c.Focused())

			test.MoveMouse(c, fileMenuPos.Add(granio.NewPos(1, 0)))
			assert.Equal(t, menuBar.Items[0], c.Focused())

			runOnMain(func() {
				test.AssertImageMatches(t, "menu_bar_active_file.png", c.Capture())
			})
		})
	})
}

func TestMenuBar_Toggle(t *testing.T) {
	runOnMain(func() {
		test.NewTempApp(t)
	})

	m1i1 := granio.NewMenuItem("New", nil)
	m1i2 := granio.NewMenuItem("Open", nil)
	m1 := granio.NewMenu("File", m1i1, m1i2)

	m2i1 := granio.NewMenuItem("Copy", nil)
	m2i2 := granio.NewMenuItem("Paste", nil)
	m2 := granio.NewMenu("Edit", m2i1, m2i2)

	menu := granio.NewMainMenu(m1, m2)

	t.Run("when menu bar is inactive", func(t *testing.T) {
		w := test.NewWindow(nil)
		defer w.Close()
		w.SetPadded(false)
		w.Resize(granio.NewSize(300, 300))
		c := w.Canvas()
		menuBar := NewMenuBar(menu, c)
		w.SetContent(container.NewWithoutLayout(menuBar))
		w.Resize(granio.NewSize(300, 300))
		menuBar.Resize(granio.NewSize(300, 0).Max(menuBar.MinSize()))

		require.False(t, menuBar.IsActive())
		test.AssertRendersToMarkup(t, "menu_bar_toggle_deactivated.xml", c)

		menuBar.Toggle()
		assert.True(t, menuBar.IsActive())
		assert.Equal(t, c.Focused(), menuBar.Items[0])
		test.AssertRendersToMarkup(t, "menu_bar_toggle_first_item_active.xml", c)
	})

	t.Run("when menu bar is active (first menu item active)", func(t *testing.T) {
		w := test.NewWindow(nil)
		defer w.Close()
		w.SetPadded(false)
		w.Resize(granio.NewSize(300, 300))
		c := w.Canvas()
		menuBar := NewMenuBar(menu, c)
		w.SetContent(container.NewWithoutLayout(menuBar))
		w.Resize(granio.NewSize(300, 300))
		menuBar.Resize(granio.NewSize(300, 0).Max(menuBar.MinSize()))

		menuBar.Toggle()
		require.True(t, menuBar.IsActive())
		test.AssertRendersToMarkup(t, "menu_bar_toggle_first_item_active.xml", c)

		menuBar.Toggle()
		assert.False(t, menuBar.IsActive())
		assert.Nil(t, c.Focused())
		test.AssertRendersToMarkup(t, "menu_bar_toggle_deactivated.xml", c)
	})

	t.Run("when menu bar is active (second menu item active)", func(t *testing.T) {
		w := test.NewWindow(nil)
		defer w.Close()
		w.SetPadded(false)
		w.Resize(granio.NewSize(300, 300))
		c := w.Canvas()
		menuBar := NewMenuBar(menu, c)
		w.SetContent(container.NewWithoutLayout(menuBar))
		w.Resize(granio.NewSize(300, 300))
		menuBar.Resize(granio.NewSize(300, 0).Max(menuBar.MinSize()))

		menuBar.Toggle()
		c.(software.WindowlessCanvas).FocusNext()
		require.True(t, menuBar.IsActive())
		test.AssertRendersToMarkup(t, "menu_bar_toggle_second_item_active.xml", c)

		menuBar.Toggle()
		assert.False(t, menuBar.IsActive())
		assert.Nil(t, c.Focused())
		test.AssertRendersToMarkup(t, "menu_bar_toggle_deactivated.xml", c)
	})
}

type notFocusableButton struct {
	widget.Label
	f func()
}

func newNotFocusableButton(l string, f func()) *notFocusableButton {
	n := &notFocusableButton{f: f}
	n.ExtendBaseWidget(n)
	n.Text = l
	return n
}

func (n *notFocusableButton) Tapped(*granio.PointEvent) {
	n.f()
}
