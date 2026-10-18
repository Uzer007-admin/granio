package container

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/stretchr/testify/assert"
)

func TestContainer_NoResize(t *testing.T) {
	rect := &canvas.Rectangle{}
	rect.SetMinSize(granio.NewSize(100, 100))

	container := NewHBox(rect)
	assert.Equal(t, granio.NewSize(0, 0), container.Size())

	container.Resize(container.MinSize())
	assert.Equal(t, rect.MinSize(), container.Size())
}
