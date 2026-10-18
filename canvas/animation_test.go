package canvas

import (
	"image/color"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewColorAnimation(t *testing.T) {
	start := color.NRGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xFF}
	stop := color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}

	var current color.Color
	anim := NewColorRGBAAnimation(start, stop, time.Second, func(c color.Color) {
		current = c
	})

	anim.Tick(0.0)
	assert.Equal(t, color.RGBA{R: 0, G: 0, B: 0, A: 0xFF}, current)
	anim.Tick(0.5)
	assert.Equal(t, color.RGBA{R: 0x7F, G: 0x7F, B: 0x7F, A: 0xFF}, current)
	anim.Tick(1.0)
	assert.Equal(t, color.RGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}, current)
}

func TestNewPositionAnimation(t *testing.T) {
	start := granio.NewPos(110, 10)
	stop := granio.NewPos(10, 110)

	var current granio.Position
	anim := NewPositionAnimation(start, stop, time.Second, func(p granio.Position) {
		current = p
	})

	anim.Tick(0.0)
	assert.Equal(t, start, current)
	anim.Tick(0.5)
	assert.Equal(t, granio.NewPos(60, 60), current)
	anim.Tick(1.0)
	assert.Equal(t, stop, current)
}

func TestNewSizeAnimation(t *testing.T) {
	start := granio.NewSize(110, 10)
	stop := granio.NewSize(10, 110)

	var current granio.Size
	anim := NewSizeAnimation(start, stop, time.Second, func(s granio.Size) {
		current = s
	})

	anim.Tick(0.0)
	assert.Equal(t, start, current)
	anim.Tick(0.5)
	assert.Equal(t, granio.NewSize(60, 60), current)
	anim.Tick(1.0)
	assert.Equal(t, stop, current)
}
