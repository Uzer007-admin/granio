package common

import (
	"testing"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/layout"
	"github.com/Uzer007-admin/granio/v3/test"
	"github.com/Uzer007-admin/granio/v3/widget"
	"github.com/stretchr/testify/assert"
)

func TestDeduplicatedObjectQueue(t *testing.T) {
	var queue deduplicatedObjectQueue
	var drained []granio.CanvasObject
	consume := func(object granio.CanvasObject) { drained = append(drained, object) }
	assert.Zero(t, queue.drain(consume))

	obj1 := &granio.Container{Layout: layout.NewCenterLayout()}
	queue.In(obj1)
	queue.In(obj1)

	obj2 := &granio.Container{Layout: layout.NewFormLayout()}
	queue.In(obj2)
	assert.Equal(t, uint64(2), queue.drain(consume))
	assert.Equal(t, []granio.CanvasObject{obj1, obj2}, drained)
	assert.Zero(t, queue.drain(consume))

	queue.In(obj1)
	assert.Equal(t, uint64(1), queue.drain(func(object granio.CanvasObject) {
		queue.In(object)
	}))
	drained = drained[:0]
	assert.Equal(t, uint64(1), queue.drain(consume))
	assert.Equal(t, []granio.CanvasObject{obj1}, drained)
}

func TestOverlayStack(t *testing.T) {
	stack := overlayStack{}
	stack.Canvas = test.NewTempWindow(t, widget.NewLabel("Empty")).Canvas()

	obj1 := &granio.Container{Layout: layout.NewCenterLayout()}
	stack.Add(obj1)
	assert.Equal(t, 1, len(stack.List()))
	assert.Equal(t, 1, len(stack.renderCaches))

	obj2 := &granio.Container{Layout: layout.NewFormLayout()}
	stack.Add(obj2)
	assert.Equal(t, 2, len(stack.List()))
	assert.Equal(t, 2, len(stack.renderCaches))

	stack.Remove(obj1)
	assert.Zero(t, len(stack.List()))
	assert.Zero(t, len(stack.renderCaches))

	stack.Remove(obj2)
	assert.Zero(t, len(stack.List()))
	assert.Zero(t, len(stack.renderCaches))

	stack.Add(nil)
	assert.Zero(t, len(stack.List()))
	assert.Zero(t, len(stack.renderCaches))

	stack.Remove(nil)
	assert.Zero(t, len(stack.List()))
	assert.Zero(t, len(stack.renderCaches))
}
