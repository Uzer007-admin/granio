//go:build !mobile

package widget

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"

	"github.com/stretchr/testify/assert"
)

func TestMenu_ItemHovered(t *testing.T) {
	sub1 := granio.NewMenuItem("sub1", nil)
	sub1.ChildMenu = granio.NewMenu(
		"",
		granio.NewMenuItem("sub1 A", nil),
		granio.NewMenuItem("sub1 B", nil),
	)
	sub2sub := granio.NewMenuItem("sub2sub", nil)
	sub2sub.ChildMenu = granio.NewMenu(
		"",
		granio.NewMenuItem("sub2sub A", nil),
		granio.NewMenuItem("sub2sub B", nil),
	)
	sub2 := granio.NewMenuItem("sub2", nil)
	sub2.ChildMenu = granio.NewMenu(
		"",
		granio.NewMenuItem("sub2 A", nil),
		granio.NewMenuItem("sub2 B", nil),
		sub2sub,
	)
	m := NewMenu(
		granio.NewMenu(
			"",
			granio.NewMenuItem("Foo", nil),
			granio.NewMenuItemSeparator(),
			granio.NewMenuItem("Bar", nil),
			sub1,
			sub2,
		),
	)
	size := m.MinSize()
	m.Resize(size)

	mi := m.Items[0].(*menuItem)
	r1 := cache.Renderer(mi).(*menuItemRenderer)
	assert.False(t, r1.background.Visible())
	mi.MouseIn(nil)
	assert.Equal(t, theme.Color(theme.ColorNameFocus), r1.background.FillColor)
	assert.True(t, r1.background.Visible())
	mi.MouseOut()
	assert.False(t, r1.background.Visible())

	sub1Widget := m.Items[3].(*menuItem)
	assert.Equal(t, sub1, sub1Widget.Item)
	sub2Widget := m.Items[4].(*menuItem)
	assert.Equal(t, sub2, sub2Widget.Item)
	assert.False(t, sub1Widget.child.Visible(), "submenu is invisible initially")
	assert.False(t, sub2Widget.child.Visible(), "submenu is invisible initially")
	sub1Widget.MouseIn(nil)
	assert.True(t, sub1Widget.child.Visible(), "hovering item shows submenu")
	assert.False(t, sub2Widget.child.Visible(), "other child menu stays hidden")
	sub1Widget.MouseOut()
	assert.True(t, sub1Widget.child.Visible(), "hover out does not hide submenu")
	assert.False(t, sub2Widget.child.Visible(), "other child menu still hidden")
	sub2Widget.MouseIn(nil)
	assert.False(t, sub1Widget.child.Visible(), "hovering other item hides current submenu")
	assert.True(t, sub2Widget.child.Visible(), "other child menu shows up")

	sub2subWidget := sub2Widget.child.Items[2].(*menuItem)
	assert.Equal(t, sub2sub, sub2subWidget.Item)
	assert.False(t, sub2subWidget.child.Visible(), "2nd level submenu is invisible initially")
	sub2Widget.MouseOut()
	sub2subWidget.MouseIn(nil)
	assert.True(t, sub2Widget.child.Visible(), "1st level submenu stays visible")
	assert.True(t, sub2subWidget.child.Visible(), "2nd level submenu shows up")
	sub2subWidget.MouseOut()
	assert.True(t, sub2Widget.child.Visible(), "1st level submenu still visible")
	assert.True(t, sub2subWidget.child.Visible(), "2nd level submenu still visible")

	sub1Widget.MouseIn(nil)
	assert.False(t, sub2Widget.child.Visible(), "1st level submenu is hidden by other submenu")
	sub2Widget.MouseIn(nil)
	assert.False(t, sub2subWidget.child.Visible(), "2nd level submenu is hidden when re-entering its parent")
}

func TestMenu_ItemWithChildTapped(t *testing.T) {
	sub := granio.NewMenuItem("sub1", nil)
	sub.ChildMenu = granio.NewMenu("", granio.NewMenuItem("sub A", nil))
	m := NewMenu(granio.NewMenu("", sub))
	size := m.MinSize()
	m.Resize(size)

	subWidget := m.Items[0].(*menuItem)
	assert.Equal(t, sub, subWidget.Item)
	assert.False(t, subWidget.child.Visible(), "submenu is invisible")
	test.Tap(subWidget)
	assert.False(t, subWidget.child.Visible(), "submenu does not show up")
}
