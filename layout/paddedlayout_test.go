package layout_test

import (
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/layout"
	"github.com/Uzer007-admin/granio/v2/theme"

	"github.com/stretchr/testify/assert"
)

func TestPaddedLayout(t *testing.T) {
	size := granio.NewSize(100, 100)

	obj := canvas.NewRectangle(color.Black)
	c := &granio.Container{
		Objects: []granio.CanvasObject{obj},
	}
	c.Resize(size)

	layout.NewPaddedLayout().Layout(c.Objects, size)

	assert.Equal(t, obj.Size().Width, size.Width-2*theme.Padding())
	assert.Equal(t, obj.Size().Height, size.Height-2*theme.Padding())
}

func TestPaddedLayout_MinSize(t *testing.T) {
	text := canvas.NewText("FooBar", color.Black)
	minSize := text.MinSize()

	c := container.NewWithoutLayout(text)
	layoutMin := layout.NewPaddedLayout().MinSize(c.Objects)

	assert.Equal(t, minSize.Width+2*theme.Padding(), layoutMin.Width)
	assert.Equal(t, minSize.Height+2*theme.Padding(), layoutMin.Height)
}

func TestContainer_PaddedLayout_MinSize(t *testing.T) {
	text := canvas.NewText("FooBar", color.Black)
	minSize := text.MinSize()

	c := container.NewWithoutLayout(text)
	c.Layout = layout.NewPaddedLayout()
	layoutMin := c.MinSize()

	assert.Equal(t, minSize.Width+2*theme.Padding(), layoutMin.Width)
	assert.Equal(t, minSize.Height+2*theme.Padding(), layoutMin.Height)
}
