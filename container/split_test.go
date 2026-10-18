package container

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func TestSplitContainer_MinSize(t *testing.T) {
	rectA := canvas.NewRectangle(color.Black)
	rectA.SetMinSize(granio.NewSize(10, 10))
	rectB := canvas.NewRectangle(color.Black)
	rectB.SetMinSize(granio.NewSize(10, 10))
	t.Run("Horizontal", func(t *testing.T) {
		minSize := NewHSplit(rectA, rectB).MinSize()
		assert.Equal(t, rectA.MinSize().Width+rectB.MinSize().Width+dividerThickness(nil), minSize.Width)
		assert.Equal(t, max(rectA.MinSize().Height, max(rectB.MinSize().Height, dividerLength(nil))), minSize.Height)
	})
	t.Run("Vertical", func(t *testing.T) {
		minSize := NewVSplit(rectA, rectB).MinSize()
		assert.Equal(t, max(rectA.MinSize().Width, max(rectB.MinSize().Width, dividerLength(nil))), minSize.Width)
		assert.Equal(t, rectA.MinSize().Height+rectB.MinSize().Height+dividerThickness(nil), minSize.Height)
	})
}

func TestSplitContainer_Resize(t *testing.T) {
	for name, tt := range map[string]struct {
		horizontal       bool
		size             granio.Size
		wantLeadingPos   granio.Position
		wantLeadingSize  granio.Size
		wantTrailingPos  granio.Position
		wantTrailingSize granio.Size
	}{
		"horizontal": {
			true,
			granio.NewSize(100, 100),
			granio.NewPos(0, 0),
			granio.NewSize(50-dividerThickness(nil)/2, 100),
			granio.NewPos(50+dividerThickness(nil)/2, 0),
			granio.NewSize(50-dividerThickness(nil)/2, 100),
		},
		"vertical": {
			false,
			granio.NewSize(100, 100),
			granio.NewPos(0, 0),
			granio.NewSize(100, 50-dividerThickness(nil)/2),
			granio.NewPos(0, 50+dividerThickness(nil)/2),
			granio.NewSize(100, 50-dividerThickness(nil)/2),
		},
		"horizontal insufficient width": {
			true,
			granio.NewSize(20, 100),
			granio.NewPos(0, 0),
			// minSize of leading is 1/3 of minSize of trailing
			granio.NewSize((20-dividerThickness(nil))/4, 100),
			granio.NewPos((20-dividerThickness(nil))/4+dividerThickness(nil), 0),
			granio.NewSize((20-dividerThickness(nil))*3/4, 100),
		},
		"vertical insufficient height": {
			false,
			granio.NewSize(100, 20),
			granio.NewPos(0, 0),
			// minSize of leading is 1/3 of minSize of trailing
			granio.NewSize(100, (20-dividerThickness(nil))/4),
			granio.NewPos(0, (20-dividerThickness(nil))/4+dividerThickness(nil)),
			granio.NewSize(100, (20-dividerThickness(nil))*3/4),
		},
		"horizontal zero width": {
			true,
			granio.NewSize(0, 100),
			granio.NewPos(0, 0),
			granio.NewSize(0, 100),
			granio.NewPos(dividerThickness(nil), 0),
			granio.NewSize(0, 100),
		},
		"horizontal zero height": {
			true,
			granio.NewSize(100, 0),
			granio.NewPos(0, 0),
			granio.NewSize(50-dividerThickness(nil)/2, 0),
			granio.NewPos(50+dividerThickness(nil)/2, 0),
			granio.NewSize(50-dividerThickness(nil)/2, 0),
		},
		"vertical zero width": {
			false,
			granio.NewSize(0, 100),
			granio.NewPos(0, 0),
			granio.NewSize(0, 50-dividerThickness(nil)/2),
			granio.NewPos(0, 50+dividerThickness(nil)/2),
			granio.NewSize(0, 50-dividerThickness(nil)/2),
		},
		"vertical zero height": {
			false,
			granio.NewSize(100, 0),
			granio.NewPos(0, 0),
			granio.NewSize(100, 0),
			granio.NewPos(0, dividerThickness(nil)),
			granio.NewSize(100, 0),
		},
	} {
		t.Run(name, func(t *testing.T) {
			objA := canvas.NewRectangle(color.White)
			objB := canvas.NewRectangle(color.Black)
			objA.SetMinSize(granio.NewSize(10, 10))
			objB.SetMinSize(granio.NewSize(30, 30))
			var c *Split
			if tt.horizontal {
				c = NewHSplit(objA, objB)
			} else {
				c = NewVSplit(objA, objB)
			}
			c.Resize(tt.size)

			assert.Equal(t, tt.wantLeadingPos, objA.Position(), "leading position")
			assert.Equal(t, tt.wantLeadingSize, objA.Size(), "leading size")
			assert.Equal(t, tt.wantTrailingPos, objB.Position(), "trailing position")
			assert.Equal(t, tt.wantTrailingSize, objB.Size(), "trailing size")
		})
	}
}

func TestSplitContainer_SetRatio(t *testing.T) {
	size := granio.NewSize(100, 100)
	usableLength := 100 - float64(dividerThickness(nil))

	objA := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	objB := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})

	t.Run("Horizontal", func(t *testing.T) {
		sc := NewHSplit(objA, objB)
		sc.Resize(size)
		t.Run("Leading", func(t *testing.T) {
			sc.SetOffset(0.75)
			sizeA := objA.Size()
			sizeB := objB.Size()
			assert.Equal(t, float32(0.75*usableLength), sizeA.Width)
			assert.Equal(t, float32(100), sizeA.Height)
			assert.Equal(t, float32(0.25*usableLength), sizeB.Width)
			assert.Equal(t, float32(100), sizeB.Height)
		})
		t.Run("Trailing", func(t *testing.T) {
			sc.SetOffset(0.25)
			sizeA := objA.Size()
			sizeB := objB.Size()
			assert.Equal(t, float32(0.25*usableLength), sizeA.Width)
			assert.Equal(t, float32(100), sizeA.Height)
			assert.Equal(t, float32(0.75*usableLength), sizeB.Width)
			assert.Equal(t, float32(100), sizeB.Height)
		})
	})
	t.Run("Vertical", func(t *testing.T) {
		sc := NewVSplit(objA, objB)
		sc.Resize(size)
		t.Run("Leading", func(t *testing.T) {
			sc.SetOffset(0.75)
			sizeA := objA.Size()
			sizeB := objB.Size()
			assert.Equal(t, float32(100), sizeA.Width)
			assert.Equal(t, float32(0.75*usableLength), sizeA.Height)
			assert.Equal(t, float32(100), sizeB.Width)
			assert.Equal(t, float32(0.25*usableLength), sizeB.Height)
		})
		t.Run("Trailing", func(t *testing.T) {
			sc.SetOffset(0.25)
			sizeA := objA.Size()
			sizeB := objB.Size()
			assert.Equal(t, float32(100), sizeA.Width)
			assert.Equal(t, float32(0.25*usableLength), sizeA.Height)
			assert.Equal(t, float32(100), sizeB.Width)
			assert.Equal(t, float32(0.75*usableLength), sizeB.Height)
		})
	})
}

func TestSplitContainer_SetRatio_limits(t *testing.T) {
	size := granio.NewSize(50, 50)
	objA := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	objA.SetMinSize(size)
	objB := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	objB.SetMinSize(size)
	t.Run("Horizontal", func(t *testing.T) {
		sc := NewHSplit(objA, objB)
		t.Run("Leading", func(t *testing.T) {
			sc.SetOffset(1.0)
			sc.Resize(granio.NewSize(200, 50))
			sizeA := objA.Size()
			sizeB := objB.Size()
			assert.Equal(t, 150-dividerThickness(nil), sizeA.Width)
			assert.Equal(t, float32(50), sizeA.Height)
			assert.Equal(t, float32(50), sizeB.Width)
			assert.Equal(t, float32(50), sizeB.Height)
		})
		t.Run("Trailing", func(t *testing.T) {
			sc.SetOffset(0.0)
			sc.Resize(granio.NewSize(200, 50))
			sizeA := objA.Size()
			sizeB := objB.Size()
			assert.Equal(t, float32(50), sizeA.Width)
			assert.Equal(t, float32(50), sizeA.Height)
			assert.Equal(t, 150-dividerThickness(nil), sizeB.Width)
			assert.Equal(t, float32(50), sizeB.Height)
		})
	})
	t.Run("Vertical", func(t *testing.T) {
		sc := NewVSplit(objA, objB)
		t.Run("Leading", func(t *testing.T) {
			sc.SetOffset(1.0)
			sc.Resize(granio.NewSize(50, 200))
			sizeA := objA.Size()
			sizeB := objB.Size()
			assert.Equal(t, float32(50), sizeA.Width)
			assert.Equal(t, 150-dividerThickness(nil), sizeA.Height)
			assert.Equal(t, float32(50), sizeB.Width)
			assert.Equal(t, float32(50), sizeB.Height)
		})
		t.Run("Trailing", func(t *testing.T) {
			sc.SetOffset(0.0)
			sc.Resize(granio.NewSize(50, 200))
			sizeA := objA.Size()
			sizeB := objB.Size()
			assert.Equal(t, float32(50), sizeA.Width)
			assert.Equal(t, float32(50), sizeA.Height)
			assert.Equal(t, float32(50), sizeB.Width)
			assert.Equal(t, 150-dividerThickness(nil), sizeB.Height)
		})
	})
}

func TestSplitContainer_ThemeOverride(t *testing.T) {
	test.NewTempApp(t)

	split := NewHSplit(canvas.NewRectangle(color.Transparent), canvas.NewRectangle(color.Transparent))
	w := test.NewWindow(split)
	defer w.Close()
	w.SetPadded(false)
	w.Resize(granio.NewSize(100, 100))
	c := w.Canvas()

	test.AssertRendersToImage(t, "split/default.png", c)

	test.ApplyTheme(t, test.NewTheme())
	test.AssertRendersToImage(t, "split/ugly.png", c)

	// set a BG that matches the theme, this is outside our container scope
	normal := test.Theme()
	bg := canvas.NewRectangle(normal.Color(theme.ColorNameBackground, theme.VariantDark))
	w.SetContent(NewStack(bg, NewThemeOverride(split, normal)))
	w.Resize(granio.NewSize(100, 100))
	test.AssertRendersToImage(t, "split/default.png", c)
}

func TestSplitContainer_swap_contents(t *testing.T) {
	dl := dividerLength(nil)
	dt := dividerThickness(nil)
	initialWidth := 10 + 10 + dt
	initialHeight := max(10, dl)
	expectedWidth := 100 + 10 + dt
	expectedHeight := max(100, dl)

	objA := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	objA.SetMinSize(granio.NewSize(10, 10))
	objB := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	objB.SetMinSize(granio.NewSize(10, 10))
	objC := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	objC.SetMinSize(granio.NewSize(100, 100))
	t.Run("Leading", func(t *testing.T) {
		sc := NewHSplit(objA, objB)
		minSize := sc.MinSize()
		assert.Equal(t, initialWidth, minSize.Width)
		assert.Equal(t, initialHeight, minSize.Height)
		sc.Leading = objC
		sc.Refresh()
		minSize = sc.MinSize()
		assert.Equal(t, expectedWidth, minSize.Width)
		assert.Equal(t, expectedHeight, minSize.Height)
	})
	t.Run("Trailing", func(t *testing.T) {
		sc := NewHSplit(objA, objB)
		minSize := sc.MinSize()
		assert.Equal(t, initialWidth, minSize.Width)
		assert.Equal(t, initialHeight, minSize.Height)
		sc.Trailing = objC
		sc.Refresh()
		minSize = sc.MinSize()
		assert.Equal(t, expectedWidth, minSize.Width)
		assert.Equal(t, expectedHeight, minSize.Height)
	})
}

func TestSplitContainer_divider_cursor(t *testing.T) {
	t.Run("Horizontal", func(t *testing.T) {
		divider := newDivider(&Split{Horizontal: true})
		assert.Equal(t, desktop.HResizeCursor, divider.Cursor())
	})
	t.Run("Vertical", func(t *testing.T) {
		divider := newDivider(&Split{Horizontal: false})
		assert.Equal(t, desktop.VResizeCursor, divider.Cursor())
	})
}

func TestSplitContainer_divider_drag(t *testing.T) {
	size := granio.NewSize(10, 10)
	objA := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	objA.SetMinSize(size)
	objB := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	objB.SetMinSize(size)
	t.Run("Horizontal", func(t *testing.T) {
		split := NewHSplit(objA, objB)
		split.Resize(granio.NewSize(105, 105))
		divider := test.TempWidgetRenderer(t, split).(*splitContainerRenderer).divider
		assert.Equal(t, 0.5, split.Offset)

		divider.Dragged(&granio.DragEvent{
			PointEvent: granio.PointEvent{Position: granio.NewPos(13, 9)},
			Dragged:    granio.NewDelta(10, 0),
		})
		assert.Equal(t, 0.6, split.Offset)

		divider.DragEnd()
		assert.Equal(t, 0.6, split.Offset)

		// test max limit
		divider.Dragged(&granio.DragEvent{
			PointEvent: granio.PointEvent{Position: granio.NewPos(53, 9)},
			Dragged:    granio.NewDelta(54, 0),
		})
		divider.Dragged(&granio.DragEvent{
			PointEvent: granio.PointEvent{Position: granio.NewPos(53, 9)},
			Dragged:    granio.NewDelta(54, 0),
		})
		assert.Equal(t, 0.9, split.Offset)

		// test returns to position after over-drag
		divider.Dragged(&granio.DragEvent{
			PointEvent: granio.PointEvent{Position: granio.NewPos(-97, 9)},
			Dragged:    granio.NewDelta(-108, 0),
		})
		divider.DragEnd()
		assert.Equal(t, 0.6, split.Offset)
	})
	t.Run("Vertical", func(t *testing.T) {
		split := NewVSplit(objA, objB)
		split.Resize(granio.NewSize(105, 105))
		divider := test.TempWidgetRenderer(t, split).(*splitContainerRenderer).divider
		assert.Equal(t, 0.5, split.Offset)

		divider.Dragged(&granio.DragEvent{
			PointEvent: granio.PointEvent{Position: granio.NewPos(13, 9)},
			Dragged:    granio.NewDelta(0, 10),
		})
		assert.Equal(t, 0.6, split.Offset)

		divider.DragEnd()
		assert.Equal(t, 0.6, split.Offset)

		// test max limit
		divider.Dragged(&granio.DragEvent{
			PointEvent: granio.PointEvent{Position: granio.NewPos(9, 53)},
			Dragged:    granio.NewDelta(0, 54),
		})
		divider.Dragged(&granio.DragEvent{
			PointEvent: granio.PointEvent{Position: granio.NewPos(9, 53)},
			Dragged:    granio.NewDelta(0, 54),
		})
		assert.Equal(t, 0.9, split.Offset)

		// test returns to position after over-drag
		divider.Dragged(&granio.DragEvent{
			PointEvent: granio.PointEvent{Position: granio.NewPos(9, -97)},
			Dragged:    granio.NewDelta(0, -108),
		})
		divider.DragEnd()
		assert.Equal(t, 0.6, split.Offset)
	})
}

func TestSplitContainer_divider_drag_StartOffsetLessThanMinSize(t *testing.T) {
	size := granio.NewSize(30, 30)
	objA := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	objA.SetMinSize(size)
	objB := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	objB.SetMinSize(size)
	t.Run("Horizontal", func(t *testing.T) {
		split := NewHSplit(objA, objB)
		split.Resize(granio.NewSize(105, 105))
		divider := test.TempWidgetRenderer(t, split).(*splitContainerRenderer).divider
		t.Run("Leading", func(t *testing.T) {
			split.SetOffset(0.1)

			divider.Dragged(&granio.DragEvent{
				Dragged: granio.NewDelta(10, 0),
			})
			divider.DragEnd()

			assert.Equal(t, 0.4, split.Offset)
		})
		t.Run("Trailing", func(t *testing.T) {
			split.SetOffset(0.9)

			divider.Dragged(&granio.DragEvent{
				Dragged: granio.NewDelta(-10, 0),
			})
			divider.DragEnd()

			assert.Equal(t, 0.6, split.Offset)
		})
	})
	t.Run("Vertical", func(t *testing.T) {
		split := NewVSplit(objA, objB)
		split.Resize(granio.NewSize(105, 105))
		divider := test.TempWidgetRenderer(t, split).(*splitContainerRenderer).divider
		t.Run("Leading", func(t *testing.T) {
			split.SetOffset(0.1)

			divider.Dragged(&granio.DragEvent{
				Dragged: granio.NewDelta(0, 10),
			})
			divider.DragEnd()

			assert.Equal(t, 0.4, split.Offset)
		})
		t.Run("Trailing", func(t *testing.T) {
			split.SetOffset(0.9)

			divider.Dragged(&granio.DragEvent{
				Dragged: granio.NewDelta(0, -10),
			})
			divider.DragEnd()

			assert.Equal(t, 0.6, split.Offset)
		})
	})
}

func TestSplitContainer_divider_hover(t *testing.T) {
	t.Run("Horizontal", func(t *testing.T) {
		divider := newDivider(&Split{Horizontal: true})
		assert.False(t, divider.hovered)

		divider.MouseIn(&desktop.MouseEvent{})
		assert.True(t, divider.hovered)

		divider.MouseOut()
		assert.False(t, divider.hovered)
	})
	t.Run("Vertical", func(t *testing.T) {
		divider := newDivider(&Split{Horizontal: false})
		assert.False(t, divider.hovered)

		divider.MouseIn(&desktop.MouseEvent{})
		assert.True(t, divider.hovered)

		divider.MouseOut()
		assert.False(t, divider.hovered)
	})
}

func TestSplitContainer_divider_MinSize(t *testing.T) {
	t.Run("Horizontal", func(t *testing.T) {
		divider := newDivider(&Split{Horizontal: true})
		minSize := divider.MinSize()
		assert.Equal(t, dividerThickness(nil), minSize.Width)
		assert.Equal(t, dividerLength(nil), minSize.Height)
	})
	t.Run("Vertical", func(t *testing.T) {
		divider := newDivider(&Split{Horizontal: false})
		minSize := divider.MinSize()
		assert.Equal(t, dividerLength(nil), minSize.Width)
		assert.Equal(t, dividerThickness(nil), minSize.Height)
	})
}

func TestSplitContainer_Hidden(t *testing.T) {
	dt := dividerThickness(nil)
	size := granio.NewSize(10, 10)
	objA := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})
	objA.SetMinSize(size)
	objB := canvas.NewRectangle(color.NRGBA{0, 0, 0, 0})

	t.Run("Horizontal_Leading", func(t *testing.T) {
		sc := NewHSplit(objA, objB)
		sc.Resize(granio.NewSize(100, 100))
		assert.Equal(t, 50-(dt/2), sc.Trailing.Size().Width)

		objA.Hide()
		sc.Refresh()
		// sole visible object should take up all space
		// (none reserved for other obj or divider)
		assert.Equal(t, float32(100), sc.Trailing.Size().Width)
	})
	t.Run("Horizontal_Trailing", func(t *testing.T) {
		objA.Show()
		sc := NewHSplit(objA, objB)
		sc.Resize(granio.NewSize(100, 100))
		assert.Equal(t, 50-(dt/2), sc.Leading.Size().Width)

		objB.Hide()
		sc.Refresh()
		assert.Equal(t, float32(100), sc.Leading.Size().Width)
	})
	t.Run("Vertical_Leading", func(t *testing.T) {
		objB.Show()
		sc := NewVSplit(objA, objB)
		sc.Resize(granio.NewSize(100, 100))
		assert.Equal(t, 50-(dt/2), sc.Trailing.Size().Height)

		objA.Hide()
		sc.Refresh()
		assert.Equal(t, float32(100), sc.Trailing.Size().Height)
	})
	t.Run("Vertical_Trailing", func(t *testing.T) {
		objA.Show()
		sc := NewVSplit(objA, objB)
		sc.Resize(granio.NewSize(100, 100))
		assert.Equal(t, 50-(dt/2), sc.Leading.Size().Height)

		objB.Hide()
		sc.Refresh()
		assert.Equal(t, float32(100), sc.Leading.Size().Height)
	})
	t.Run("MinSize", func(t *testing.T) {
		objA.Show()
		objB.Hide()
		objB.SetMinSize(granio.NewSize(10, 10))
		sc := NewVSplit(objA, objB)
		assert.Equal(t, objA.MinSize(), sc.MinSize())
	})
}

func TestSplitContainer_UpdateOffsetDoesNotRefreshContent(t *testing.T) {
	objA := &refreshCountingWidget{}
	objB := &refreshCountingWidget{}
	split := NewHSplit(objA, objB)
	split.SetOffset(0.4)
	assert.Equal(t, 0, objA.refreshCount)
	assert.Equal(t, 0, objB.refreshCount)

	split.Refresh()
	assert.Equal(t, 1, objA.refreshCount)
	assert.Equal(t, 1, objB.refreshCount)
}

func TestSplitContainer_HoverDividerDoesNotRefreshContent(t *testing.T) {
	objA := &refreshCountingWidget{}
	objB := &refreshCountingWidget{}
	split := NewHSplit(objA, objB)
	r := cache.Renderer(split).(*splitContainerRenderer)
	r.divider.MouseIn(&desktop.MouseEvent{})

	assert.Equal(t, 0, objA.refreshCount)
	assert.Equal(t, 0, objB.refreshCount)

	split.Refresh()
	assert.Equal(t, 1, objA.refreshCount)
	assert.Equal(t, 1, objB.refreshCount)
}

type refreshCountingWidget struct {
	widget.BaseWidget

	refreshCount int
}

func (*refreshCountingWidget) CreateRenderer() granio.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.Transparent))
}

func (r *refreshCountingWidget) Refresh() {
	r.refreshCount++
	r.BaseWidget.Refresh()
}
