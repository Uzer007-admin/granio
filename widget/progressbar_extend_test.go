package widget

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/stretchr/testify/assert"
)

type extendedProgressBar struct {
	ProgressBar
}

func newExtendedProgressBar() *extendedProgressBar {
	ret := &extendedProgressBar{}
	ret.ExtendBaseWidget(ret)

	return ret
}

func TestProgressBarRenderer_Extended_Layout(t *testing.T) {
	bar := newExtendedProgressBar()
	bar.Resize(granio.NewSize(100, 100))
	r := test.TempWidgetRenderer(t, bar).(*progressRenderer)

	assert.Equal(t, 0.0, bar.Value)
	assert.Equal(t, float32(0), r.bar.Size().Width)

	bar.SetValue(.5)
	assert.Equal(t, .5, bar.Value)
	assert.Equal(t, float32(50), r.bar.Size().Width)
}
