package canvas_test

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"testing"

	// This is the decoder registry imported by Fyne's font renderer.
	_ "github.com/go-text/render"
)

func TestTIFFPackBitsLimits(t *testing.T) {
	for _, byteOrder := range []struct {
		name  string
		order binary.ByteOrder
	}{
		{"little endian", binary.LittleEndian},
		{"big endian", binary.BigEndian},
	} {
		for _, tc := range []struct {
			name   string
			rows   uint32
			packed []byte
		}{
			{"repeated run", 1, []byte{0x81, 0xab}},
			{"literal run", 1, []byte{8, 0xab, 1, 2, 3, 4, 5, 6, 7, 8}},
			{"inflated rows per strip", ^uint32(0), []byte{0x81, 0xab}},
		} {
			t.Run(byteOrder.name+"/"+tc.name, func(t *testing.T) {
				data := packBitsTIFF(byteOrder.order, tc.rows, tc.packed)
				config, format, err := image.DecodeConfig(bytes.NewReader(data))
				if err != nil || format != "tiff" || config.Width != 1 || config.Height != 1 {
					t.Fatalf("valid tiny TIFF config: config=%v format=%q error=%v", config, format, err)
				}
				// Config bounds alone cannot prevent excessive strip expansion.
				if _, _, err := image.Decode(bytes.NewReader(data)); err == nil {
					t.Fatal("PackBits data exceeding the decoder block limit was accepted")
				}
			})
		}
		t.Run(byteOrder.name+"/valid literal", func(t *testing.T) {
			data := packBitsTIFF(byteOrder.order, 1, []byte{0, 0xab})
			img, format, err := image.Decode(bytes.NewReader(data))
			if err != nil || format != "tiff" {
				t.Fatalf("valid PackBits TIFF rejected: format=%q error=%v", format, err)
			}
			if got := color.GrayModel.Convert(img.At(0, 0)).(color.Gray).Y; got != 0xab {
				t.Fatalf("decoded pixel=%x, want ab", got)
			}
		})
		t.Run(byteOrder.name+"/valid exact block boundary", func(t *testing.T) {
			// TIFF permits up to 8 bytes per pixel (RGBA64) in the shared block limit.
			data := packBitsTIFF(byteOrder.order, 1, []byte{7, 0xab, 1, 2, 3, 4, 5, 6, 7})
			img, _, err := image.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("exact limit rejected: %v", err)
			}
			if got := color.GrayModel.Convert(img.At(0, 0)).(color.Gray).Y; got != 0xab {
				t.Fatalf("pixel=%x, want ab", got)
			}
		})
	}
}

// packBitsTIFF creates a complete 1x1 grayscale TIFF, including its strip table.
// The inputs are small even when they describe malicious expansion.
func packBitsTIFF(order binary.ByteOrder, rows uint32, packed []byte) []byte {
	tags := [][3]uint32{
		{256, 4, 1}, {257, 4, 1}, {258, 3, 8}, {259, 3, 32773},
		{262, 3, 1}, {273, 4, 0}, {277, 3, 1}, {278, 4, rows},
		{279, 4, uint32(len(packed))},
	}
	offset := 8 + 2 + 12*len(tags) + 4
	tags[5][2] = uint32(offset)
	data := make([]byte, offset+len(packed))
	copy(data, "II")
	if order == binary.BigEndian {
		copy(data, "MM")
	}
	order.PutUint16(data[2:4], 42)
	order.PutUint32(data[4:8], 8)
	order.PutUint16(data[8:10], uint16(len(tags)))
	for i, tag := range tags {
		entry := data[10+i*12 : 10+(i+1)*12]
		order.PutUint16(entry, uint16(tag[0]))
		order.PutUint16(entry[2:4], uint16(tag[1]))
		order.PutUint32(entry[4:8], 1)
		if tag[1] == 3 {
			order.PutUint16(entry[8:10], uint16(tag[2]))
		} else {
			order.PutUint32(entry[8:12], tag[2])
		}
	}
	copy(data[offset:], packed)
	return data
}
