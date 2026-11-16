package widget_test

import (
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/internal/widget"
	"github.com/Uzer007-admin/granio/v3/test"
)

func TestNewSimpleRenderer(t *testing.T) {
	r := canvas.NewRectangle(color.Transparent)
	o := &simpleWidget{obj: r}
	o.ExtendBaseWidget(o)
	w := test.NewTempWindow(t, o)
	w.Resize(granio.NewSize(100, 100))

	test.AssertRendersToMarkup(t, "simple_renderer.xml", w.Canvas())
}

type simpleWidget struct {
	widget.Base
	obj granio.CanvasObject
}

func (s *simpleWidget) CreateRenderer() granio.WidgetRenderer {
	return widget.NewSimpleRenderer(s.obj)
}
