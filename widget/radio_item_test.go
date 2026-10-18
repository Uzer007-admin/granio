package widget

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"

	"github.com/stretchr/testify/assert"
)

var globalRadioRenderer granio.WidgetRenderer

func BenchmarkRadioCreateRenderer(b *testing.B) {
	var renderer granio.WidgetRenderer
	widget := &radioItem{}
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		renderer = widget.CreateRenderer()
	}

	// Avoid having the value optimized out by the compiler.
	globalRadioRenderer = renderer
}

func TestRadioItem_FocusIndicator_Centered_Vertically(t *testing.T) {
	item := newRadioItem("Hello", nil)
	render := test.TempWidgetRenderer(t, item).(*radioItemRenderer)
	render.Layout(granio.NewSize(200, 100))

	focusIndicatorSize := theme.IconInlineSize() + 2*theme.Padding()
	heightCenterOffset := (100 - focusIndicatorSize) / 2
	assert.Equal(t, granio.NewPos(theme.Padding()/2, heightCenterOffset), render.focusIndicator.Position1)
}
