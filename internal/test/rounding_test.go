//go:build !tamago && !noos

package test

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func TestPixCloseEnoughRoundingBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name string
		a, b []uint8
		want bool
	}{
		{"equal", []uint8{0, 255, 128, 255}, []uint8{0, 255, 128, 255}, true},
		{"sparse rounding", []uint8{0, 255, 128, 255}, []uint8{1, 255, 128, 255}, true},
		{"visible change", []uint8{0, 255, 128, 255}, []uint8{31, 255, 128, 255}, false},
		{"widespread change", []uint8{0, 255, 128, 255}, []uint8{1, 254, 128, 255}, false},
		{"different sizes", []uint8{0, 255}, []uint8{0, 255, 128, 255}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := pixCloseEnough(tc.a, tc.b); got != tc.want {
				t.Fatalf("pixCloseEnough() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPixelsForImageNormalizesAlpha(t *testing.T) {
	straight := image.NewNRGBA(image.Rect(0, 0, 2, 1))
	straight.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 1})
	straight.SetNRGBA(1, 0, color.NRGBA{R: 255, G: 255, B: 255})
	premultiplied := image.NewRGBA(straight.Bounds())
	premultiplied.SetRGBA(0, 0, color.RGBA{R: 1, A: 1})
	if !bytes.Equal(pixelsForImage(t, straight), pixelsForImage(t, premultiplied)) {
		t.Fatal("Equivalent straight and premultiplied colors must compare equally")
	}
	if straight.NRGBAAt(1, 0).R != 255 {
		t.Fatal("Image comparison modified the source image")
	}
}
