package layout_test

import (
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/layout"

	"github.com/stretchr/testify/assert"
)

func TestCenterLayout(t *testing.T) {
	size := granio.NewSize(100, 100)
	minSize := granio.NewSize(10, 10)

	obj := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	obj.SetMinSize(minSize)
	c := &granio.Container{
		Objects: []granio.CanvasObject{obj},
	}
	c.Resize(size)

	layout.NewCenterLayout().Layout(c.Objects, size)

	assert.Equal(t, obj.Size(), minSize)
	assert.Equal(t, granio.NewPos(45, 45), obj.Position())
}

func TestCenterLayout_MinSize(t *testing.T) {
	text := canvas.NewText("Padding", color.NRGBA{0, 0xff, 0, 0})
	minSize := text.MinSize()

	c := container.NewWithoutLayout(text)
	layoutMin := layout.NewCenterLayout().MinSize(c.Objects)

	assert.Equal(t, minSize, layoutMin)
}

func TestCenterLayout_MinSize_Hidden(t *testing.T) {
	text1 := canvas.NewText("Padding", color.NRGBA{0, 0xff, 0, 0})
	text1.Hide()
	text2 := canvas.NewText("1\n2", color.NRGBA{0, 0xff, 0, 0})

	c := container.NewWithoutLayout(text1, text2)
	layoutMin := layout.NewCenterLayout().MinSize(c.Objects)

	assert.Equal(t, text2.MinSize(), layoutMin)
}

func TestContainerCenterLayoutMinSize(t *testing.T) {
	text := canvas.NewText("Padding", color.NRGBA{0, 0xff, 0, 0})
	minSize := text.MinSize()

	c := container.NewWithoutLayout(text)
	c.Layout = layout.NewCenterLayout()
	layoutMin := c.MinSize()

	assert.Equal(t, minSize, layoutMin)
}
