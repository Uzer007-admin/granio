package container

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/assert"
)

func TestMultipleWindows_Add(t *testing.T) {
	m := NewMultipleWindows()
	assert.Zero(t, len(m.Windows))

	m.Add(NewInnerWindow("1", widget.NewLabel("Inside")))
	assert.Equal(t, 1, len(m.Windows))
}

func TestMultipleWindows_Drag(t *testing.T) {
	w := NewInnerWindow("1", widget.NewLabel("Inside"))
	m := NewMultipleWindows(w)
	m.Resize(granio.NewSize(30, 30))
	_ = test.TempWidgetRenderer(t, m) // initialise display
	assert.Equal(t, 1, len(m.Windows))

	assert.True(t, w.Position().IsZero())
	w.OnDragged(&granio.DragEvent{Dragged: granio.Delta{DX: 10, DY: 5}})
	assert.Equal(t, float32(10), w.Position().X)
	assert.Equal(t, float32(5), w.Position().Y)

	drag := &granio.DragEvent{Dragged: granio.Delta{DX: -10, DY: -5}}
	drag.Position = granio.NewPos(5, 5)
	drag.AbsolutePosition = granio.NewPos(5, 5)
	w.OnDragged(drag)
	assert.Equal(t, float32(0), w.Position().X)
	assert.Equal(t, float32(0), w.Position().Y)
}

func TestMultipleWindows_RaiseToTop(t *testing.T) {
	w1 := NewInnerWindow("1", widget.NewLabel("Content"))
	m := NewMultipleWindows(w1)
	assert.Equal(t, w1, m.Top())

	w2 := NewInnerWindow("2", widget.NewLabel("Content"))
	m.Add(w2)
	assert.Equal(t, w2, m.Top())

	m.RaiseToTop(w1)
	assert.Equal(t, w1, m.Top())
}

func TestMultipleWindows_Top(t *testing.T) {
	m := NewMultipleWindows()
	assert.Nil(t, m.Top())

	w1 := NewInnerWindow("1", widget.NewLabel("Content"))
	m.Add(w1)
	assert.Equal(t, w1, m.Top())
}
