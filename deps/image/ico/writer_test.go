package ico

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"testing"
)

func readPng(filename string) (image.Image, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

func TestWriter(t *testing.T) {
	fn := "testdata/icondata.png"
	m0, err := readPng(fn)
	if err != nil {
		t.Fatal(fn, err)
	}

	var encoded bytes.Buffer
	if err := Encode(&encoded, m0); err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(&encoded)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := decoded.Bounds(), m0.Bounds(); got != want {
		t.Fatalf("decoded bounds %v, want %v", got, want)
	}
}
