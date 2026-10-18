package test

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v2"
	fynecanvas "github.com/Uzer007-admin/granio/v2/canvas"
)

func Test_driver_AbsolutePositionForObject(t *testing.T) {
	d := &driver{}
	w := d.CreateWindow("Test Window")
	o := fynecanvas.NewRectangle(color.Black)
	w.SetContent(o)
	w.Resize(granio.NewSize(320, 200))

	t.Run("for padded window", func(t *testing.T) {
		w.SetPadded(true)
		assert.Equal(t, granio.NewPos(4, 4), d.AbsolutePositionForObject(o), "safe area offset (2,3) is subtracted")
	})

	t.Run("for non-padded window", func(t *testing.T) {
		w.SetPadded(false)
		assert.Equal(t, granio.NewPos(0, 0), d.AbsolutePositionForObject(o), "safe area offset (2,3) is subtracted")
	})
}

func TestDriver_CreateWindow(t *testing.T) {
	d := &driver{}
	w := d.CreateWindow("Test Window")

	assert.Equal(t, "Test Window", w.Title())
}
