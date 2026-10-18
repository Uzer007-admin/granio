package widget

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v2"
)

type dummyCanvas struct {
	granio.Canvas
}

func TestOverlayContainer_Tapped_Dismiss(t *testing.T) {
	dismissed := false
	o := NewOverlayContainer(nil, &dummyCanvas{}, func() { dismissed = true })

	o.Tapped(&granio.PointEvent{})
	assert.True(t, dismissed)
}

func TestOverlayContainer_TappedSecondary_Dismiss(t *testing.T) {
	dismissed := false
	o := NewOverlayContainer(nil, &dummyCanvas{}, func() { dismissed = true })

	o.TappedSecondary(&granio.PointEvent{})
	assert.True(t, dismissed)
}

func TestOverlayContainer_Tapped_NilDismiss(t *testing.T) {
	o := NewOverlayContainer(nil, &dummyCanvas{}, nil)

	assert.NotPanics(t, func() {
		o.Tapped(&granio.PointEvent{})
	})
}

func TestOverlayContainer_TappedSecondary_NilDismiss(t *testing.T) {
	o := NewOverlayContainer(nil, &dummyCanvas{}, nil)

	assert.NotPanics(t, func() {
		o.TappedSecondary(&granio.PointEvent{})
	})
}
