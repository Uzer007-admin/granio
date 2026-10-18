package layout_test

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/layout"
	"github.com/Uzer007-admin/granio/v2/theme"

	"github.com/stretchr/testify/assert"
)

// NewMinSizeRect returns a new Rectangle instance
func NewMinSizeRect(minSize granio.Size) *canvas.Rectangle {
	rect := &canvas.Rectangle{}
	rect.SetMinSize(minSize)

	return rect
}

func TestHBoxLayout_Simple(t *testing.T) {
	cellSize := granio.NewSize(50, 50)

	obj1 := NewMinSizeRect(cellSize)
	obj2 := NewMinSizeRect(cellSize)
	obj3 := NewMinSizeRect(cellSize)

	c := container.NewHBox(obj1, obj2, obj3)

	// We are not in a window. Resize the container to a default size.
	c.Resize(c.MinSize())

	assert.Equal(t, c.MinSize(), granio.NewSize(150+(theme.Padding()*2), 50))

	assert.Equal(t, obj1.Size(), cellSize)
	cell2Pos := granio.NewPos(50+theme.Padding(), 0)
	assert.Equal(t, cell2Pos, obj2.Position())
	cell3Pos := granio.NewPos(100+theme.Padding()*2, 0)
	assert.Equal(t, cell3Pos, obj3.Position())
}

func TestCustomPaddedHBoxLayout_Simple(t *testing.T) {
	cellSize := granio.NewSize(50, 50)

	obj1 := NewMinSizeRect(cellSize)
	obj2 := NewMinSizeRect(cellSize)
	obj3 := NewMinSizeRect(cellSize)

	padding := float32(15)
	c := container.New(layout.NewCustomPaddedHBoxLayout(padding), obj1, obj2, obj3)

	// We are not in a window. Resize the container to a default size.
	c.Resize(c.MinSize())

	assert.Equal(t, c.MinSize(), granio.NewSize(150+padding*2, 50))

	assert.Equal(t, obj1.Size(), cellSize)
	cell2Pos := granio.NewPos(50+padding, 0)
	assert.Equal(t, cell2Pos, obj2.Position())
	cell3Pos := granio.NewPos(100+padding*2, 0)
	assert.Equal(t, cell3Pos, obj3.Position())
}

func TestHBoxLayout_HiddenItem(t *testing.T) {
	cellSize := granio.NewSize(50, 50)

	obj1 := NewMinSizeRect(cellSize)
	obj2 := NewMinSizeRect(cellSize)
	obj2.Hide()
	obj3 := NewMinSizeRect(cellSize)

	c := container.NewHBox(obj1, obj2, obj3)

	// We are not in a window. Resize the container to a default size.
	c.Resize(c.MinSize())

	assert.Equal(t, c.MinSize(), granio.NewSize(100+theme.Padding(), 50))

	assert.Equal(t, obj1.Size(), cellSize)
	cell3Pos := granio.NewPos(50+theme.Padding(), 0)
	assert.Equal(t, cell3Pos, obj3.Position())
}

func TestHBoxLayout_Wide(t *testing.T) {
	cellSize := granio.NewSize(50, 100)

	obj1 := NewMinSizeRect(cellSize)
	obj2 := NewMinSizeRect(cellSize.Subtract(granio.NewSize(0, 25)))
	obj3 := NewMinSizeRect(cellSize)

	c := container.NewHBox(obj1, obj2, obj3)
	c.Resize(granio.NewSize(308, 100))
	assert.Equal(t, granio.NewSize(150+(theme.Padding()*2), 100), c.MinSize())

	assert.Equal(t, float32(50), obj1.Size().Width)
	assert.Equal(t, float32(50), obj2.Size().Width)
	cell2Pos := granio.NewPos(50+theme.Padding(), 0)
	assert.Equal(t, cell2Pos, obj2.Position())
	cell3Pos := granio.NewPos(100+theme.Padding()*2, 0)
	assert.Equal(t, cell3Pos, obj3.Position())
}

func TestHBoxLayout_Tall(t *testing.T) {
	cellSize := granio.NewSize(50, 100)

	obj1 := NewMinSizeRect(cellSize)
	obj2 := NewMinSizeRect(cellSize.Subtract(granio.NewSize(0, 25)))
	obj3 := NewMinSizeRect(cellSize)

	c := container.NewHBox(obj1, obj2, obj3)

	// We are not in a window. Resize the container to a default size.
	c.Resize(c.MinSize())

	assert.Equal(t, c.MinSize(), granio.NewSize(150+(theme.Padding()*2), 100))

	assert.Equal(t, obj1.Size(), cellSize)
	assert.Equal(t, obj2.Size(), cellSize)
	cell2Pos := granio.NewPos(50+theme.Padding(), 0)
	assert.Equal(t, cell2Pos, obj2.Position())
	cell3Pos := granio.NewPos(100+theme.Padding()*2, 0)
	assert.Equal(t, cell3Pos, obj3.Position())
}

func TestHBoxLayout_Spacer(t *testing.T) {
	cellSize := granio.NewSize(50, 100)

	obj1 := NewMinSizeRect(cellSize)
	obj2 := NewMinSizeRect(cellSize.Subtract(granio.NewSize(0, 25)))
	obj3 := NewMinSizeRect(cellSize)

	c := container.NewHBox(layout.NewSpacer(), obj1, obj2, obj3)
	c.Resize(granio.NewSize(300, 100))
	assert.Equal(t, c.MinSize(), granio.NewSize(150+(theme.Padding()*2), 100))

	assert.Equal(t, float32(50), obj1.Size().Width)
	assert.Equal(t, float32(50), obj2.Size().Width)
	cell2Pos := granio.NewPos(200-theme.Padding(), 0)
	assert.Equal(t, cell2Pos, obj2.Position())
	cell3Pos := granio.NewPos(250, 0)
	assert.Equal(t, cell3Pos, obj3.Position())
}

func TestHBoxLayout_MiddleSpacer(t *testing.T) {
	cellSize := granio.NewSize(50, 100)

	obj1 := NewMinSizeRect(cellSize)
	obj2 := NewMinSizeRect(cellSize.Subtract(granio.NewSize(0, 25)))
	obj3 := NewMinSizeRect(cellSize)

	c := container.NewHBox(obj1, obj2, layout.NewSpacer(), obj3)
	c.Resize(granio.NewSize(300, 100))
	assert.Equal(t, c.MinSize(), granio.NewSize(150+(theme.Padding()*2), 100))

	assert.Equal(t, float32(50), obj1.Size().Width)
	assert.Equal(t, float32(50), obj2.Size().Width)
	cell2Pos := granio.NewPos(50+theme.Padding(), 0)
	assert.Equal(t, cell2Pos, obj2.Position())
	cell3Pos := granio.NewPos(250, 0)
	assert.Equal(t, cell3Pos, obj3.Position())
}

func TestVBoxLayout_Simple(t *testing.T) {
	cellSize := granio.NewSize(50, 50)

	obj1 := NewMinSizeRect(cellSize)
	obj2 := NewMinSizeRect(cellSize)
	obj3 := NewMinSizeRect(cellSize)

	c := container.NewVBox(obj1, obj2, obj3)

	// We are not in a window. Resize the container to a default size.
	c.Resize(c.MinSize())

	assert.Equal(t, c.MinSize(), granio.NewSize(50, 150+(theme.Padding()*2)))

	assert.Equal(t, obj1.Size(), cellSize)
	cell2Pos := granio.NewPos(0, 50+theme.Padding())
	assert.Equal(t, cell2Pos, obj2.Position())
	cell3Pos := granio.NewPos(0, 100+theme.Padding()*2)
	assert.Equal(t, cell3Pos, obj3.Position())
}

func TestCustomPaddedVBoxLayout_Simple(t *testing.T) {
	cellSize := granio.NewSize(50, 50)

	obj1 := NewMinSizeRect(cellSize)
	obj2 := NewMinSizeRect(cellSize)
	obj3 := NewMinSizeRect(cellSize)

	padding := float32(15)
	c := container.New(layout.NewCustomPaddedVBoxLayout(padding), obj1, obj2, obj3)

	// We are not in a window. Resize the container to a default size.
	c.Resize(c.MinSize())

	assert.Equal(t, c.MinSize(), granio.NewSize(50, 150+padding*2))

	assert.Equal(t, obj1.Size(), cellSize)
	cell2Pos := granio.NewPos(0, 50+padding)
	assert.Equal(t, cell2Pos, obj2.Position())
	cell3Pos := granio.NewPos(0, 100+padding*2)
	assert.Equal(t, cell3Pos, obj3.Position())
}

func TestVBoxLayout_HiddenItem(t *testing.T) {
	cellSize := granio.NewSize(50, 50)

	obj1 := NewMinSizeRect(cellSize)
	obj2 := NewMinSizeRect(cellSize)
	obj2.Hide()
	obj3 := NewMinSizeRect(cellSize)

	c := container.NewVBox(obj1, obj2, obj3)

	// We are not in a window. Resize the container to a default size.
	c.Resize(c.MinSize())

	assert.Equal(t, c.MinSize(), granio.NewSize(50, 100+theme.Padding()))

	assert.Equal(t, obj1.Size(), cellSize)
	cell3Pos := granio.NewPos(0, 50+theme.Padding())
	assert.Equal(t, cell3Pos, obj3.Position())
}

func TestVBoxLayout_Wide(t *testing.T) {
	cellSize := granio.NewSize(100, 50)

	obj1 := NewMinSizeRect(cellSize)
	obj2 := NewMinSizeRect(cellSize.Subtract(granio.NewSize(25, 0)))
	obj3 := NewMinSizeRect(cellSize)

	c := container.NewVBox(obj1, obj2, obj3)

	// We are not in a window. Resize the container to a default size.
	c.Resize(c.MinSize())

	assert.Equal(t, c.MinSize(), granio.NewSize(100, 150+(theme.Padding()*2)))

	assert.Equal(t, obj1.Size(), cellSize)
	assert.Equal(t, obj2.Size(), cellSize)
	cell2Pos := granio.NewPos(0, 50+theme.Padding())
	assert.Equal(t, cell2Pos, obj2.Position())
	cell3Pos := granio.NewPos(0, 100+theme.Padding()*2)
	assert.Equal(t, cell3Pos, obj3.Position())
}

func TestVBoxLayout_Tall(t *testing.T) {
	cellSize := granio.NewSize(100, 50)

	obj1 := NewMinSizeRect(cellSize)
	obj2 := NewMinSizeRect(cellSize.Subtract(granio.NewSize(25, 0)))
	obj3 := NewMinSizeRect(cellSize)

	c := container.NewVBox(obj1, obj2, obj3)
	c.Resize(granio.NewSize(100, 308))
	assert.Equal(t, c.MinSize(), granio.NewSize(100, 150+(theme.Padding()*2)))

	assert.Equal(t, float32(50), obj1.Size().Height)
	assert.Equal(t, float32(50), obj2.Size().Height)
	cell2Pos := granio.NewPos(0, 50+theme.Padding())
	assert.Equal(t, cell2Pos, obj2.Position())
	cell3Pos := granio.NewPos(0, 100+theme.Padding()*2)
	assert.Equal(t, cell3Pos, obj3.Position())
}

func TestVBoxLayout_Spacer(t *testing.T) {
	cellSize := granio.NewSize(100, 50)

	obj1 := NewMinSizeRect(cellSize)
	obj2 := NewMinSizeRect(cellSize.Subtract(granio.NewSize(25, 0)))
	obj3 := NewMinSizeRect(cellSize)

	c := container.NewVBox(layout.NewSpacer(), obj1, obj2, obj3)
	c.Resize(granio.NewSize(100, 300))
	assert.Equal(t, c.MinSize(), granio.NewSize(100, 150+(theme.Padding()*2)))

	assert.Equal(t, float32(50), obj1.Size().Height)
	assert.Equal(t, float32(50), obj2.Size().Height)
	cell2Pos := granio.NewPos(0, 200-theme.Padding())
	assert.Equal(t, cell2Pos, obj2.Position())
	cell3Pos := granio.NewPos(0, 250)
	assert.Equal(t, cell3Pos, obj3.Position())
}

func TestVBoxLayout_MiddleSpacer(t *testing.T) {
	cellSize := granio.NewSize(100, 50)

	obj1 := NewMinSizeRect(cellSize)
	obj2 := NewMinSizeRect(cellSize.Subtract(granio.NewSize(25, 0)))
	obj3 := NewMinSizeRect(cellSize)

	c := container.NewVBox(obj1, obj2, layout.NewSpacer(), obj3)
	c.Resize(granio.NewSize(100, 300))
	assert.Equal(t, c.MinSize(), granio.NewSize(100, 150+(theme.Padding()*2)))

	assert.Equal(t, float32(50), obj1.Size().Height)
	assert.Equal(t, float32(50), obj2.Size().Height)
	cell2Pos := granio.NewPos(0, 50+theme.Padding())
	assert.Equal(t, cell2Pos, obj2.Position())
	cell3Pos := granio.NewPos(0, 250)
	assert.Equal(t, cell3Pos, obj3.Position())
}

// Test for issue #4259 - spacer in HBox with hidden item causing add'l trailing padding
func TestHBoxLayout_MiddleSpacerHiddenItem(t *testing.T) {
	cellSize := granio.NewSize(100, 50)

	obj1 := NewMinSizeRect(cellSize)
	obj2 := NewMinSizeRect(cellSize)
	obj3 := NewMinSizeRect(cellSize)

	c := container.NewHBox(obj1, obj2, layout.NewSpacer(), obj3)
	c.Resize(granio.NewSize(400, 100))
	assert.Equal(t, granio.NewPos(300, 0), obj3.Position())

	obj2.Hide()
	c.Refresh()
	assert.Equal(t, granio.NewPos(300, 0), obj3.Position())
}
