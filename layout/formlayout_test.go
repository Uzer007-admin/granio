package layout_test

import (
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/layout"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"

	"github.com/stretchr/testify/assert"
)

func BenchmarkFormLayout(b *testing.B) {
	b.StopTimer()

	minSize := granio.Size{}
	form := layout.NewFormLayout()
	label1 := canvas.NewRectangle(color.Black)
	content1 := canvas.NewRectangle(color.Black)
	label2 := canvas.NewRectangle(color.Black)
	content2 := canvas.NewRectangle(color.Black)

	objects := []granio.CanvasObject{label1, content1, label2, content2}

	b.StartTimer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		minSize = form.MinSize(objects)
	}

	assert.Equal(b, minSize, minSize)
}

func TestFormLayout(t *testing.T) {
	gridSize := granio.NewSize(125, 125)

	label1 := canvas.NewRectangle(color.Black)
	label1.SetMinSize(granio.NewSize(50, 50))
	content1 := canvas.NewRectangle(color.Black)
	content1.SetMinSize(granio.NewSize(100, 100))

	label2 := canvas.NewRectangle(color.Black)
	label2.SetMinSize(granio.NewSize(70, 30))
	content2 := canvas.NewRectangle(color.Black)
	content2.SetMinSize(granio.NewSize(120, 80))

	container := &granio.Container{
		Objects: []granio.CanvasObject{label1, content1, label2, content2},
	}
	container.Resize(gridSize)

	layout.NewFormLayout().Layout(container.Objects, gridSize)

	assert.Equal(t, granio.NewSize(70, 100), label1.Size())
	assert.Equal(t, granio.NewSize(120, 100), content1.Size())
	assert.Equal(t, granio.NewSize(70, 80), label2.Size())
	assert.Equal(t, granio.NewSize(120, 80), content2.Size())
}

func TestFormLayout_Text(t *testing.T) {
	size := granio.NewSize(120, 50)
	label := canvas.NewText("Label", color.Black)
	content := canvas.NewText("Content", color.Black)

	container := &granio.Container{
		Objects: []granio.CanvasObject{label, content},
	}
	container.Resize(size)
	layout.NewFormLayout().Layout(container.Objects, size)

	assert.Equal(t, label.Size().Height, content.Size().Height)
}

func TestFormLayout_Hidden(t *testing.T) {
	gridSize := granio.NewSize(190+theme.Padding(), 125)

	label1 := canvas.NewRectangle(color.Black)
	label1.SetMinSize(granio.NewSize(70, 50))
	label1.Hide()
	content1 := canvas.NewRectangle(color.Black)
	content1.SetMinSize(granio.NewSize(120, 100))
	content1.Hide()

	label2 := canvas.NewRectangle(color.Black)
	label2.SetMinSize(granio.NewSize(50, 30))
	content2 := canvas.NewRectangle(color.Black)
	content2.SetMinSize(granio.NewSize(100, 80))

	container := &granio.Container{
		Objects: []granio.CanvasObject{label1, content1, label2, content2},
	}
	container.Resize(gridSize)

	layout.NewFormLayout().Layout(container.Objects, gridSize)

	assert.Equal(t, granio.NewSize(50, 80), label2.Size())
	assert.Equal(t, granio.NewSize(140, 80), content2.Size())
	assert.Equal(t, granio.NewPos(0, 0), label2.Position())
	assert.Equal(t, granio.NewPos(50+theme.Padding(), 0), content2.Position())
}

func TestFormLayout_StretchX(t *testing.T) {
	wideSize := granio.NewSize(150, 50)

	label1 := canvas.NewRectangle(color.Black)
	label1.SetMinSize(granio.NewSize(50, 50))
	content1 := canvas.NewRectangle(color.Black)
	content1.SetMinSize(granio.NewSize(50, 50))

	container := &granio.Container{
		Objects: []granio.CanvasObject{label1, content1},
	}
	container.Resize(wideSize)

	layout.NewFormLayout().Layout(container.Objects, wideSize)

	assert.Equal(t, granio.NewSize(50, 50), label1.Size())
	assert.Equal(t, granio.NewSize(wideSize.Width-50-theme.Padding(), 50), content1.Size())
}

func TestFormLayout_MinSize(t *testing.T) {
	label1 := canvas.NewRectangle(color.Black)
	label1.SetMinSize(granio.NewSize(50, 50))
	content1 := canvas.NewRectangle(color.Black)
	content1.SetMinSize(granio.NewSize(100, 100))

	label2 := canvas.NewRectangle(color.Black)
	label2.SetMinSize(granio.NewSize(70, 30))
	content2 := canvas.NewRectangle(color.Black)
	content2.SetMinSize(granio.NewSize(120, 80))

	container := &granio.Container{
		Objects: []granio.CanvasObject{label1, content1, label2, content2},
	}

	l := layout.NewFormLayout()
	layoutMin := l.MinSize(container.Objects)
	expectedRowWidth := 70 + 120 + theme.Padding()
	expectedRowHeight := 100 + 80 + theme.Padding()
	assert.Equal(t, granio.NewSize(expectedRowWidth, expectedRowHeight), layoutMin)

	text := canvas.NewText("Text", color.Black)
	value := widget.NewLabel("Text")
	l = layout.NewFormLayout()
	layoutMin = l.MinSize([]granio.CanvasObject{text, value})
	// check that the text minsize is padded to match a label
	assert.Equal(t, value.MinSize().Width*2+theme.Padding(), layoutMin.Width)
}

func TestFormLayout_MinSize_Hidden(t *testing.T) {
	label1 := canvas.NewRectangle(color.Black)
	label1.SetMinSize(granio.NewSize(50, 50))
	content1 := canvas.NewRectangle(color.Black)
	content1.SetMinSize(granio.NewSize(100, 100))

	label2 := canvas.NewRectangle(color.Black)
	label2.SetMinSize(granio.NewSize(70, 30))
	label2.Hide()
	content2 := canvas.NewRectangle(color.Black)
	content2.SetMinSize(granio.NewSize(120, 80))
	content2.Hide()

	container := &granio.Container{
		Objects: []granio.CanvasObject{label1, content1, label2, content2},
	}

	l := layout.NewFormLayout()
	layoutMin := l.MinSize(container.Objects)
	expectedRowWidth := 50 + 100 + theme.Padding()
	expectedRowHeight := float32(100)
	assert.Equal(t, granio.NewSize(expectedRowWidth, expectedRowHeight), layoutMin)
}

func TestFormLayout_MinSize_CanvasText_SingleRow(t *testing.T) {
	text1 := canvas.NewText("First text", color.Black)
	text2 := canvas.NewText("Second text", color.Black)

	l := layout.NewFormLayout()
	layoutMin := l.MinSize([]granio.CanvasObject{text1, text2})

	inner := theme.InnerPadding()
	min1 := text1.MinSize()
	min2 := text2.MinSize()
	expectedWidth := (min1.Width + inner*2) + (min2.Width + inner*2) + theme.Padding()

	expectedHeight := max(min1.Height, min2.Height) + inner*2

	assert.Equal(t, granio.NewSize(expectedWidth, expectedHeight), layoutMin)
}

func TestFormLayout_MinSize_CanvasText_TwoRows(t *testing.T) {
	label1 := canvas.NewText("First Text", color.Black)
	value1 := canvas.NewText("First Value", color.Black)
	label2 := canvas.NewText("Second Text", color.Black)
	value2 := canvas.NewText("Second Value", color.Black)

	l := layout.NewFormLayout()
	layoutMin := l.MinSize([]granio.CanvasObject{label1, value1, label2, value2})

	inner := theme.InnerPadding()
	l1 := label1.MinSize()
	l2 := label2.MinSize()
	v1 := value1.MinSize()
	v2 := value2.MinSize()
	labelCol := max(l1.Width+inner*2, l2.Width+inner*2)
	valueCol := max(v1.Width+inner*2, v2.Width+inner*2)
	expectedWidth := labelCol + valueCol + theme.Padding()
	row1 := max(l1.Height+inner*2, v1.Height+inner*2)
	row2 := max(l2.Height+inner*2, v2.Height+inner*2)
	expectedHeight := row1 + row2 + theme.Padding()

	assert.Equal(t, granio.NewSize(expectedWidth, expectedHeight), layoutMin)
}
