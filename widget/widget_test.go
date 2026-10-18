package widget

import (
	"image/color"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	internalWidget "github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/layout"
	"github.com/Uzer007-admin/granio/v2/test"
)

type myWidget struct {
	DisableableWidget

	refreshed chan bool
}

func (m *myWidget) Refresh() {
	m.refreshed <- true
}

func (m *myWidget) CreateRenderer() granio.WidgetRenderer {
	m.ExtendBaseWidget(m)
	return internalWidget.NewSimpleRenderer(&granio.Container{})
}

func TestApplyThemeCalled(t *testing.T) {
	widget := &myWidget{refreshed: make(chan bool, 10)}

	test.NewTempWindow(t, widget)
	granio.CurrentApp().Settings().SetTheme(test.NewTheme())

	func() {
		select {
		case <-widget.refreshed:
		case <-time.After(1 * time.Second):
			assert.Fail(t, "Timed out waiting for theme apply")
		}
	}()
}

func TestApplyThemeCalledChild(t *testing.T) {
	child := &myWidget{refreshed: make(chan bool, 10)}
	parent := &granio.Container{Layout: layout.NewVBoxLayout(), Objects: []granio.CanvasObject{child}}

	test.NewTempWindow(t, parent)
	granio.CurrentApp().Settings().SetTheme(test.NewTheme())
	func() {
		select {
		case <-child.refreshed:
		case <-time.After(1 * time.Second):
			assert.Fail(t, "Timed out waiting for child theme apply")
		}
	}()
}

func TestSimpleRenderer(t *testing.T) {
	test.NewTempApp(t)

	c := &granio.Container{Layout: layout.NewStackLayout(), Objects: []granio.CanvasObject{
		newTestWidget(canvas.NewRectangle(color.Gray{Y: 0x79})),
		newTestWidget(canvas.NewText("Hi", color.Black)),
	}}

	window := test.NewWindow(c)
	defer window.Close()

	test.AssertImageMatches(t, "simple_renderer.png", window.Canvas().Capture())
}

type testWidget struct {
	BaseWidget
	obj granio.CanvasObject
}

func newTestWidget(o granio.CanvasObject) granio.Widget {
	t := &testWidget{obj: o}
	t.ExtendBaseWidget(t)
	return t
}

func (t *testWidget) CreateRenderer() granio.WidgetRenderer {
	return NewSimpleRenderer(t.obj)
}

func waitForBinding() {
	time.Sleep(time.Millisecond * 100) // data resolves on background thread
}
