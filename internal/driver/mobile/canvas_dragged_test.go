//go:build mobile && (!windows || !ci)

package mobile

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/assert"
)

func Test_canvas_Dragged(t *testing.T) {
	dragged := false
	var draggedObj granio.Draggable
	scroll := container.NewScroll(widget.NewLabel("Hi\nHi\nHi"))
	c := newCanvas(granio.CurrentDevice()).(*canvas)
	c.SetContent(scroll)
	c.Resize(granio.NewSize(40, 24))
	assert.Equal(t, float32(0), scroll.Offset.Y)

	c.tapDown(granio.NewPos(32, 3), 0)
	c.tapMove(granio.NewPos(32, 10), 0, func(wid granio.Draggable, ev *granio.DragEvent) {
		wid.Dragged(ev)
		dragged = true
		draggedObj = wid
	})

	assert.True(t, dragged)
	assert.Equal(t, scroll, draggedObj)
	dragged = false
	c.tapMove(granio.NewPos(32, 5), 0, func(wid granio.Draggable, ev *granio.DragEvent) {
		wid.Dragged(ev)
		dragged = true
	})
	assert.True(t, dragged)
	assert.Equal(t, granio.NewPos(0, 5), scroll.Offset)
}

func Test_canvas_Dragged_nested(t *testing.T) {
	dragged := false
	var draggedObj granio.Draggable
	childScroll := container.NewHScroll(widget.NewLabel("Child *Scroll"))
	parentScroll := container.NewScroll(container.NewVBox(childScroll, widget.NewLabel("Hi\nHi\nHi")))
	c := newCanvas(granio.CurrentDevice()).(*canvas)
	c.SetContent(parentScroll)
	c.Resize(granio.NewSize(40, 24))
	assert.Equal(t, granio.NewPos(0, 0), parentScroll.Offset)
	assert.Equal(t, granio.NewPos(0, 0), childScroll.Offset)

	c.tapDown(granio.NewPos(10, 10), 0)
	c.tapMove(granio.NewPos(15, 15), 0, func(wid granio.Draggable, ev *granio.DragEvent) {
		wid.Dragged(ev)
		dragged = true
		draggedObj = wid
	})
	assert.True(t, dragged)
	assert.Equal(t, parentScroll, draggedObj)
	assert.Equal(t, granio.NewPos(0, 0), parentScroll.Offset)
	assert.Equal(t, granio.NewPos(0, 0), childScroll.Offset)

	dragged = false
	c.tapMove(granio.NewPos(9, 7), 0, func(wid granio.Draggable, ev *granio.DragEvent) {
		wid.Dragged(ev)
		dragged = true
	})
	assert.True(t, dragged)
	assert.Equal(t, granio.NewPos(0, 8), parentScroll.Offset)
	assert.Equal(t, granio.NewPos(6, 0), childScroll.Offset)
}
