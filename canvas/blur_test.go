package canvas_test

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/test"
)

func TestBlur(t *testing.T) {
	test.NewTempApp(t)
	bg := canvas.NewImageFromFile("testdata/Utah_teapot.png")
	b1 := canvas.NewBlur(35)
	w := test.NewTempWindow(t, container.NewWithoutLayout(bg, b1))
	w.SetPadded(false)
	size := granio.NewSize(300, 243)
	bg.Resize(size)
	w.Resize(size)

	b1.Move(granio.NewPos(120, 110))
	b1.Resize(granio.NewSize(140, 100))

	test.AssertRendersToImage(t, "blur.png", w.Canvas())

	b1.CornerRadius = 30
	test.AssertRendersToImage(t, "blur_rounded.png", w.Canvas())
}
