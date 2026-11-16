package painter

import (
	"image"
	_ "image/jpeg" // avoid users having to import when using image widget
	_ "image/png"  // avoid the same for PNG images
	"sync"

	"golang.org/x/image/draw"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
)

// PaintImage renders a given fyne Image to a Go standard image
// If a granio.Canvas is given and the image’s fill mode is “fill original” the image’s min size has
// to fit its original size. If it doesn’t, PaintImage does not paint the image but adjusts its min size.
// The image will then be painted on the next frame because of the min size change.
func PaintImage(img *canvas.Image, c granio.Canvas, width, height int) image.Image {
	if width <= 0 || height <= 0 {
		return nil
	}

	if img.Size().IsZero() && c == nil { // an image without size or canvas won't get rendered unless we setup
		img.Resize(granio.NewSize(float32(width), float32(height)))
	}

	dst := img.Image
	if dst == nil {
		dst = image.NewNRGBA(image.Rect(0, 0, width, height))
	}
	if img.FillMode == canvas.ImageFillCover {
		return dst
	}

	size := dst.Bounds().Size()
	if width != size.X || height != size.Y {
		dst = scaleImage(dst, width, height, img.ScaleMode)
	}
	return dst
}

func scaleImage(pixels image.Image, scaledW, scaledH int, scale canvas.ImageScale) image.Image {
	if scale == canvas.ImageScaleFastest || scale == canvas.ImageScalePixels {
		// do not perform software scaling
		return pixels
	}

	bounds := pixels.Bounds()
	pixW := int(min(float32(scaledW), float32(bounds.Dx()))) // don't push more pixels than we have to
	pixH := int(min(float32(scaledH), float32(bounds.Dy()))) // the GL calls will scale this up on GPU.
	scaledBounds := image.Rect(0, 0, pixW, pixH)
	tex := image.NewNRGBA(scaledBounds)
	if scale != canvas.ImageScaleSmooth {
		granio.LogError("Invalid canvas.ImageScale value, using canvas.ImageScaleSmooth", nil)
	}
	ScaleImageSmooth(tex, pixels)
	return tex
}

type imageScaleSize struct{ dst, src image.Point }

type imageScaler struct {
	size   imageScaleSize
	scaler draw.Scaler
}

var smoothImageScalers struct {
	sync.Mutex
	entries [4]imageScaler
	next    int
}

// ScaleImageSmooth scales the whole source into the destination. Reusing a
// scaler of the same dimensions keeps the filter weights and work buffer warm
// without retaining either image. Limit the size of pooled work buffers.
func ScaleImageSmooth(dst *image.NRGBA, src image.Image) {
	d, s := dst.Bounds(), src.Bounds()
	if d.Empty() || s.Empty() {
		return
	}
	scaler := draw.Scaler(draw.CatmullRom)
	const maxAxis, maxWorkPixels = 4096, 128 * 1024
	if d.Size() != s.Size() && max(d.Dx(), d.Dy(), s.Dx(), s.Dy()) <= maxAxis && d.Dx() <= maxWorkPixels/s.Dy() {
		size := imageScaleSize{dst: d.Size(), src: s.Size()}
		scaler = cachedImageScaler(size)
	}
	scaler.Scale(dst, d, src, s, draw.Over, nil)
}

func cachedImageScaler(size imageScaleSize) draw.Scaler {
	cache := &smoothImageScalers
	cache.Lock()
	defer cache.Unlock()
	for _, entry := range cache.entries {
		if entry.size == size {
			return entry.scaler
		}
	}
	scaler := draw.CatmullRom.NewScaler(size.dst.X, size.dst.Y, size.src.X, size.src.Y)
	cache.entries[cache.next] = imageScaler{size: size, scaler: scaler}
	cache.next = (cache.next + 1) % len(cache.entries)
	return scaler
}
