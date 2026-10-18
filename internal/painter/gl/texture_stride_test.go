package gl

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func TestPackedRGBASubImage(t *testing.T) {
	parent := image.NewRGBA(image.Rect(-2, -3, 3, 2))
	for y := -3; y < 2; y++ {
		for x := -2; x < 3; x++ {
			parent.SetRGBA(x, y, color.RGBA{R: byte(x + 3), G: byte(y + 4), A: 255})
		}
	}
	sub := parent.SubImage(image.Rect(-1, -2, 2, 1)).(*image.RGBA)
	want := make([]byte, 0, 3*3*4)
	for y := -2; y < 1; y++ {
		for x := -1; x < 2; x++ {
			c := parent.RGBAAt(x, y)
			want = append(want, c.R, c.G, c.B, c.A)
		}
	}
	if got := packedRGBAPixels(sub); !bytes.Equal(got, want) {
		t.Fatalf("packed rows=%v, want %v", got, want)
	}
	if c := parent.RGBAAt(0, 0); c != (color.RGBA{R: 3, G: 4, A: 255}) {
		t.Fatal("packing changed the source image")
	}
	for _, img := range []*image.RGBA{parent, parent.SubImage(image.Rect(-1, -2, 2, -1)).(*image.RGBA)} {
		pixels := packedRGBAPixels(img)
		if &pixels[0] != &img.Pix[0] {
			t.Fatal("packed or single-row images should retain the zero-copy fast path")
		}
	}
	if got := packedRGBAPixels(image.NewRGBA(image.Rect(0, 0, 0, 1))); len(got) != 0 {
		t.Fatal("empty image must have no pixels")
	}
}
