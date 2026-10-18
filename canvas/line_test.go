package canvas_test

import (
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"

	"github.com/stretchr/testify/assert"
)

func TestLine_MinSize(t *testing.T) {
	line := canvas.NewLine(color.Black)
	minSize := line.MinSize()

	assert.Positive(t, minSize.Width)
	assert.Positive(t, minSize.Height)
}

func TestLine_Move(t *testing.T) {
	line := canvas.NewLine(color.Black)

	line.Resize(granio.NewSize(10, 10))
	assert.Equal(t, granio.NewPos(0, 0), line.Position1)
	assert.Equal(t, granio.NewPos(10, 10), line.Position2)

	line.Move(granio.NewPos(5, 5))
	assert.Equal(t, granio.NewPos(5, 5), line.Position1)
	assert.Equal(t, granio.NewPos(15, 15), line.Position2)

	// rotate
	line.Position1 = granio.NewPos(0, 10)
	line.Position2 = granio.NewPos(10, 0)
	line.Move(granio.NewPos(10, 10))
	assert.Equal(t, granio.NewPos(10, 20), line.Position1)
	assert.Equal(t, granio.NewPos(20, 10), line.Position2)
}

func TestLine_Resize(t *testing.T) {
	line := canvas.NewLine(color.Black)

	line.Resize(granio.NewSize(10, 0))
	size := line.Size()

	assert.Equal(t, float32(10), size.Width)
	assert.Equal(t, float32(0), size.Height)

	// rotate
	line.Position1 = granio.NewPos(0, 10)
	line.Position2 = granio.NewPos(10, 0)
	line.Resize(granio.NewSize(20, 20))
	assert.Equal(t, granio.NewPos(0, 20), line.Position1)
	assert.Equal(t, granio.NewPos(20, 0), line.Position2)
}

func TestLine_StrokeColor(t *testing.T) {
	c := color.White
	line := canvas.NewLine(c)

	assert.Equal(t, c, line.StrokeColor)
}
