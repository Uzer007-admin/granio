//go:build !windows || !ci

package mobile

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	fynecanvas "github.com/Uzer007-admin/granio/v2/canvas"
	internalWidget "github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"

	"github.com/stretchr/testify/assert"
)

func TestMobileCanvas_DismissBar(t *testing.T) {
	c := newCanvas(granio.CurrentDevice()).(*canvas)
	c.SetContent(fynecanvas.NewRectangle(theme.Color(theme.ColorNameBackground)))
	menu := granio.NewMainMenu(
		granio.NewMenu("Test"),
	)
	c.showMenu(menu)
	c.Resize(granio.NewSize(100, 100))

	assert.NotNil(t, c.menu)
	// simulate tap as the test util does not know about our menu...
	c.tapDown(granio.NewPos(80, 20), 1)
	c.tapUp(granio.NewPos(80, 20), 1, nil, nil, nil, nil)
	assert.Nil(t, c.menu)
}

func TestMobileCanvas_DismissMenu(t *testing.T) {
	c := newCanvas(granio.CurrentDevice()).(*canvas)
	c.padded = false
	c.SetContent(fynecanvas.NewRectangle(theme.Color(theme.ColorNameBackground)))
	menu := granio.NewMainMenu(
		granio.NewMenu("Test", granio.NewMenuItem("TapMe", func() {})),
	)
	c.showMenu(menu)
	c.Resize(granio.NewSize(100, 100))

	assert.NotNil(t, c.menu)
	menuObj := c.menu.(*granio.Container).Objects[1].(*granio.Container).Objects[1].(*menuLabel)
	point := &granio.PointEvent{Position: granio.NewPos(10, 10)}
	menuObj.Tapped(point)

	tapMeItem := c.Overlays().Top().(*internalWidget.OverlayContainer).Content.(*widget.PopUpMenu).Items[0].(granio.Tappable)
	tapMeItem.Tapped(point)
	assert.Nil(t, c.menu)
}

func TestMobileCanvas_Menu(t *testing.T) {
	c := &canvas{}
	labels := []string{"File", "Edit"}
	menu := granio.NewMainMenu(
		granio.NewMenu(labels[0]),
		granio.NewMenu(labels[1]),
	)

	c.showMenu(menu)
	menuObjects := c.menu.(*granio.Container).Objects[1].(*granio.Container)
	assert.Equal(t, 3, len(menuObjects.Objects))
	header, ok := menuObjects.Objects[0].(*granio.Container)
	assert.True(t, ok)
	closed, ok := header.Objects[0].(*widget.Button)
	assert.True(t, ok)
	assert.Equal(t, theme.CancelIcon().Name(), closed.Icon.Name())

	for i := 1; i < 3; i++ {
		item, ok := menuObjects.Objects[i].(*menuLabel)
		assert.True(t, ok)
		assert.Equal(t, labels[i-1], item.menu.Label)
	}
}

func TestMobileCanvas_MenuChild(t *testing.T) {
	c := &canvas{}
	c.Initialize(c, nil)
	c.Resize(granio.NewSize(100, 200))

	child := granio.NewMenu("Child", granio.NewMenuItem("One", func() {}))
	parent := granio.NewMenuItem("Parent", func() {})
	parent.ChildMenu = child
	menu := granio.NewMainMenu(
		granio.NewMenu("Top", parent),
	)

	c.showMenu(menu)
	topObj := c.menu.(*granio.Container).Objects[1].(*granio.Container).Objects[1].(*menuLabel)
	assert.Equal(t, "Top", topObj.menu.Label)

	test.Tap(topObj)
	assert.Equal(t, 1, len(c.Overlays().List()))
	rootOverlay := c.Overlays().Top()
	parentItem := rootOverlay.(*internalWidget.OverlayContainer).Content.(*widget.PopUpMenu).Items[0]
	parentDetails := test.WidgetRenderer(parentItem.(granio.Widget))
	text := parentDetails.Objects()[1].(*fynecanvas.Text)
	assert.Equal(t, "Parent", text.Text)

	test.Tap(parentItem.(granio.Tappable))
	assert.NotNil(t, c.menu) // still visible

	childMenu := parentItem.(interface{ Child() *widget.Menu }).Child()
	assert.NotNil(t, childMenu)
	assert.True(t, childMenu.Visible())
	assert.Equal(t, 1, len(childMenu.Items))

	childDetails := test.WidgetRenderer(childMenu.Items[0].(granio.Widget))
	text = childDetails.Objects()[1].(*fynecanvas.Text)
	assert.Equal(t, "One", text.Text)
}

func dummyWin(d *driver, title string) *window {
	ret := &window{title: title}
	d.windows = append(d.windows, ret)

	return ret
}

func TestMobileDriver_FindMenu(t *testing.T) {
	m1 := granio.NewMainMenu(granio.NewMenu("1"))
	m2 := granio.NewMainMenu(granio.NewMenu("2"))

	d := NewGoMobileDriver().(*driver)
	w1 := dummyWin(d, "top")
	w1.SetMainMenu(m1)
	assert.Equal(t, m1, d.findMenu(w1))

	w2 := dummyWin(d, "child")
	assert.Equal(t, m1, d.findMenu(w2))

	w2.SetMainMenu(m2)
	assert.Equal(t, m2, d.findMenu(w2))
}
