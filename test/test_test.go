package test_test

import (
	"bytes"
	"image/color"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func TestAssertCanvasTappableAt(t *testing.T) {
	c := test.NewCanvas()
	b := widget.NewButton("foo", nil)
	c.SetContent(b)
	c.Resize(granio.NewSize(300, 300))
	b.Resize(granio.NewSize(100, 100))
	b.Move(granio.NewPos(100, 100))

	tt := &testing.T{}
	assert.True(t, test.AssertCanvasTappableAt(tt, c, granio.NewPos(101, 101)), "tappable found")
	assert.False(t, tt.Failed(), "test did not fail")

	tt = &testing.T{}
	assert.False(t, test.AssertCanvasTappableAt(tt, c, granio.NewPos(99, 99)), "tappable not found")
	assert.True(t, tt.Failed(), "test failed")
}

func TestAssertObjectRendersToImage(t *testing.T) {
	obj := canvas.NewCircle(color.Black)
	obj.Resize(granio.NewSize(20, 20))

	test.AssertObjectRendersToImage(t, "circle.png", obj)
}

func TestRenderObjectToMarkup(t *testing.T) {
	obj := canvas.NewCircle(color.Black)
	obj.Resize(granio.NewSize(20, 20))

	want, err := os.ReadFile("testdata/circle.xml")
	require.NoError(t, err)
	// Fix Windows newlines
	want = bytes.ReplaceAll(want, []byte("\r\n"), []byte("\n"))
	got := strings.ReplaceAll(test.RenderObjectToMarkup(obj), "\r\n", "\n")
	assert.Equal(t, string(want), got, "existing master is equal to rendered markup")
}

func TestAssertObjectRendersToMarkup(t *testing.T) {
	obj := canvas.NewCircle(color.Black)
	obj.Resize(granio.NewSize(20, 20))

	assert.True(t, test.AssertObjectRendersToMarkup(t, "circle.xml", obj), "existing master is equal to rendered markup")
}

func TestAssertRendersToImage(t *testing.T) {
	c := test.NewCanvas() // no painter, would be cycle - images will be blank
	c.SetContent(canvas.NewCircle(color.Black))
	c.Resize(granio.NewSize(10, 10))

	t.Run("non-existing master", func(t *testing.T) {
		tt := &testing.T{}
		assert.False(t, test.AssertRendersToImage(tt, "non_existing_master.png", c), "non existing master is not equal to rendered image")
		assert.True(t, tt.Failed(), "test failed")
		_, err := os.Stat("testdata/failed/non_existing_master.png")
		assert.NoError(t, err)
	})

	t.Run("matching master", func(t *testing.T) {
		tt := &testing.T{}
		assert.True(t, test.AssertRendersToImage(tt, "image_master.png", c), "existing master is equal to rendered image")
		assert.False(t, tt.Failed(), "test should not fail")
		_, err := os.Stat("testdata/failed/image_master.png")
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("diffing master", func(t *testing.T) {
		c.Resize(granio.NewSize(15, 15))

		tt := &testing.T{}
		assert.False(t, test.AssertRendersToImage(tt, "image_diffing_master.png", c), "existing master is not equal to rendered image")
		assert.True(t, tt.Failed(), "test should fail")
		_, err := os.Stat("testdata/failed/image_diffing_master.png")
		assert.NoError(t, err)
	})

	if !t.Failed() {
		_ = os.RemoveAll("testdata/failed")
	}
}

func TestRenderToMarkup(t *testing.T) {
	c := test.NewCanvas()
	c.SetContent(canvas.NewCircle(color.Black))

	want, err := os.ReadFile("testdata/markup_master.xml")
	require.NoError(t, err)
	// Fix Windows newlines
	want = bytes.ReplaceAll(want, []byte("\r\n"), []byte("\n"))
	got := strings.ReplaceAll(test.RenderToMarkup(c), "\r\n", "\n")
	assert.Equal(t, string(want), got, "existing master is equal to rendered markup")
}

func TestAssertRendersToMarkup(t *testing.T) {
	c := test.NewCanvas()
	c.SetContent(canvas.NewCircle(color.Black))

	markup := "<canvas padded size=\"9x9\">\n" +
		"\t<content>\n" +
		"\t\t<circle fillColor=\"rgba(0,0,0,255)\" pos=\"4,4\" size=\"1x1\"/>\n" +
		"\t</content>\n" +
		"</canvas>\n"

	t.Run("non-existing master", func(t *testing.T) {
		tt := &testing.T{}
		assert.False(t, test.AssertRendersToMarkup(tt, "non_existing_master.xml", c), "non existing master is not equal to rendered markup")
		assert.True(t, tt.Failed(), "test failed")
		assert.Equal(t, markup, readMarkup(t, "testdata/failed/non_existing_master.xml"), "markup was written to disk")
	})

	t.Run("matching master", func(t *testing.T) {
		tt := &testing.T{}
		assert.True(t, test.AssertRendersToMarkup(tt, "markup_master.xml", c), "existing master is equal to rendered markup")
		assert.False(t, tt.Failed(), "test should not fail")
	})

	t.Run("diffing master", func(t *testing.T) {
		tt := &testing.T{}
		assert.False(t, test.AssertRendersToMarkup(tt, "markup_diffing_master.xml", c), "existing master is not equal to rendered markup")
		assert.True(t, tt.Failed(), "test should fail")
		assert.Equal(t, markup, readMarkup(t, "testdata/failed/markup_diffing_master.xml"), "markup was written to disk")
	})

	if !t.Failed() {
		os.RemoveAll("testdata/failed")
	}
}

func TestDrag(t *testing.T) {
	c := test.NewCanvas()
	c.SetPadded(false)
	d := &draggable{}
	c.SetContent(container.NewWithoutLayout(d))
	c.Resize(granio.NewSize(30, 30))
	d.Resize(granio.NewSize(20, 20))
	d.Move(granio.NewPos(10, 10))

	test.Drag(c, granio.NewPos(5, 5), 10, 10)
	assert.Nil(t, d.event, "nothing happens if no draggable was found at position")
	assert.False(t, d.wasDragged)

	test.Drag(c, granio.NewPos(15, 15), 17, 42)
	assert.Equal(t, &granio.DragEvent{
		PointEvent: granio.PointEvent{Position: granio.Position{X: 5, Y: 5}},
		Dragged:    granio.NewDelta(17, 42),
	}, d.event)
	assert.True(t, d.wasDragged)
}

func TestFocusNext(t *testing.T) {
	c := test.NewCanvas()
	f1 := &focusable{}
	f2 := &focusable{}
	f3 := &focusable{}
	c.SetContent(container.NewWithoutLayout(f1, f2, f3))

	assert.Nil(t, c.Focused())
	assert.False(t, f1.focused)
	assert.False(t, f2.focused)
	assert.False(t, f3.focused)

	test.FocusNext(c)
	assert.Equal(t, f1, c.Focused())
	assert.True(t, f1.focused)
	assert.False(t, f2.focused)
	assert.False(t, f3.focused)

	test.FocusNext(c)
	assert.Equal(t, f2, c.Focused())
	assert.False(t, f1.focused)
	assert.True(t, f2.focused)
	assert.False(t, f3.focused)

	test.FocusNext(c)
	assert.Equal(t, f3, c.Focused())
	assert.False(t, f1.focused)
	assert.False(t, f2.focused)
	assert.True(t, f3.focused)

	test.FocusNext(c)
	assert.Equal(t, f1, c.Focused())
	assert.True(t, f1.focused)
	assert.False(t, f2.focused)
	assert.False(t, f3.focused)
}

func TestFocusPrevious(t *testing.T) {
	c := test.NewCanvas()
	f1 := &focusable{}
	f2 := &focusable{}
	f3 := &focusable{}
	c.SetContent(container.NewWithoutLayout(f1, f2, f3))

	assert.Nil(t, c.Focused())
	assert.False(t, f1.focused)
	assert.False(t, f2.focused)
	assert.False(t, f3.focused)

	test.FocusPrevious(c)
	assert.Equal(t, f3, c.Focused())
	assert.False(t, f1.focused)
	assert.False(t, f2.focused)
	assert.True(t, f3.focused)

	test.FocusPrevious(c)
	assert.Equal(t, f2, c.Focused())
	assert.False(t, f1.focused)
	assert.True(t, f2.focused)
	assert.False(t, f3.focused)

	test.FocusPrevious(c)
	assert.Equal(t, f1, c.Focused())
	assert.True(t, f1.focused)
	assert.False(t, f2.focused)
	assert.False(t, f3.focused)

	test.FocusPrevious(c)
	assert.Equal(t, f3, c.Focused())
	assert.False(t, f1.focused)
	assert.False(t, f2.focused)
	assert.True(t, f3.focused)
}

func TestScroll(t *testing.T) {
	c := test.NewCanvas()
	c.SetPadded(false)
	s := &scrollable{}
	c.SetContent(container.NewWithoutLayout(s))
	c.Resize(granio.NewSize(30, 30))
	s.Resize(granio.NewSize(20, 20))
	s.Move(granio.NewPos(10, 10))

	test.Scroll(c, granio.NewPos(5, 5), 10, 10)
	assert.Nil(t, s.event, "nothing happens if no scrollable was found at position")

	test.Scroll(c, granio.NewPos(15, 15), 17, 42)
	assert.Equal(t, &granio.ScrollEvent{Scrolled: granio.NewDelta(17, 42)}, s.event)
}

func readMarkup(t *testing.T, path string) string {
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	return string(raw)
}

var _ granio.Draggable = (*draggable)(nil)

type draggable struct {
	widget.BaseWidget
	event      *granio.DragEvent
	wasDragged bool
}

func (d *draggable) DragEnd() {
	d.wasDragged = true
}

func (d *draggable) Dragged(event *granio.DragEvent) {
	d.event = event
}

var _ granio.Focusable = (*focusable)(nil)

type focusable struct {
	widget.BaseWidget
	focused bool
}

func (f *focusable) FocusGained() {
	f.focused = true
}

func (f *focusable) FocusLost() {
	f.focused = false
}

func (*focusable) TypedKey(*granio.KeyEvent) {
}

func (*focusable) TypedRune(rune) {
}

var _ granio.Scrollable = (*scrollable)(nil)

type scrollable struct {
	widget.BaseWidget
	event *granio.ScrollEvent
}

func (s *scrollable) Scrolled(event *granio.ScrollEvent) {
	s.event = event
}
