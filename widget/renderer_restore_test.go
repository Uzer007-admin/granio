package widget

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/stretchr/testify/assert"
)

func TestRendererRecreationRestoresLayout(t *testing.T) {
	test.NewApp()
	label := NewLabel("Centered after idle")
	label.Alignment = granio.TextAlignCenter
	label.Resize(granio.NewSize(240, 60))
	first := cache.Renderer(label)
	type geometry struct {
		size granio.Size
		pos  granio.Position
	}
	expected := make([]geometry, len(first.Objects()))
	for i, object := range first.Objects() {
		expected[i] = geometry{object.Size(), object.Position()}
	}
	cache.DestroyRenderer(label)
	second := cache.Renderer(label)
	assert.NotSame(t, first, second)
	if !assert.Len(t, second.Objects(), len(expected)) {
		return
	}
	for i, object := range second.Objects() {
		assert.Equal(t, expected[i], geometry{object.Size(), object.Position()}, "recreated renderer object %d", i)
	}
}
