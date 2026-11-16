package software

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/internal/cache"
	"github.com/Uzer007-admin/granio/v3/internal/painter"
	"github.com/Uzer007-admin/granio/v3/internal/scale"
	"github.com/Uzer007-admin/granio/v3/theme"

	"github.com/anthonynsimon/bild/blur"

	"golang.org/x/image/draw"
)

type gradient interface {
	Generate(int, int) image.Image
	Size() granio.Size
}

func drawBlur(c granio.Canvas, blurObj *canvas.Blur, pos granio.Position, base *image.NRGBA, clip image.Rectangle) {
	if blurObj.Radius == 0 {
		return
	}

	scaledWidth := scale.ToScreenCoordinate(c, blurObj.Size().Width)
	scaledHeight := scale.ToScreenCoordinate(c, blurObj.Size().Height)
	scaledX, scaledY := scale.ToScreenCoordinate(c, pos.X), scale.ToScreenCoordinate(c, pos.Y)
	bounds := clip.Intersect(image.Rect(scaledX, scaledY, scaledX+scaledWidth, scaledY+scaledHeight))
	if bounds.Empty() {
		return
	}

	crop := base.SubImage(bounds)
	blurred := blur.Gaussian(crop, float64(blurObj.Radius*c.Scale()))

	cornerRadius := min(painter.GetMaximumRadius(blurObj.Size()), blurObj.CornerRadius)

	if cornerRadius > 0.5 {
		applyRoundedCorners(blurred, cornerRadius*c.Scale())
	}

	draw.Draw(base, bounds, blurred, blurred.Bounds().Min, draw.Over)
}

func drawArc(c granio.Canvas, arc *canvas.Arc, pos granio.Position, base *image.NRGBA, clip image.Rectangle) {
	pad := painter.VectorPad(arc)
	scaledWidth := scale.ToScreenCoordinate(c, arc.Size().Width+pad*2)
	scaledHeight := scale.ToScreenCoordinate(c, arc.Size().Height+pad*2)
	scaledX, scaledY := scale.ToScreenCoordinate(c, pos.X-pad), scale.ToScreenCoordinate(c, pos.Y-pad)
	bounds := clip.Intersect(image.Rect(scaledX, scaledY, scaledX+scaledWidth, scaledY+scaledHeight))

	raw := painter.DrawArc(arc, pad, func(in float32) float32 {
		return float32(math.Round(float64(in) * float64(c.Scale())))
	})

	drawClippedVector(base, bounds, raw, image.Pt(scaledX, scaledY))
}

func drawCircle(c granio.Canvas, circle *canvas.Circle, pos granio.Position, base *image.NRGBA, clip image.Rectangle) {
	pad := painter.VectorPad(circle)
	scaledWidth := scale.ToScreenCoordinate(c, circle.Size().Width+pad*2)
	scaledHeight := scale.ToScreenCoordinate(c, circle.Size().Height+pad*2)
	scaledX, scaledY := scale.ToScreenCoordinate(c, pos.X-pad), scale.ToScreenCoordinate(c, pos.Y-pad)
	bounds := clip.Intersect(image.Rect(scaledX, scaledY, scaledX+scaledWidth, scaledY+scaledHeight))

	raw := painter.DrawCircle(circle, pad, func(in float32) float32 {
		return float32(math.Round(float64(in) * float64(c.Scale())))
	})

	if painter.IsShadowVisible(circle.Shadow) {
		drawShadow(c, circle, circle.Size(), circle.Shadow, pad, base, clip, pos)
	}

	drawClippedVector(base, bounds, raw, image.Pt(scaledX, scaledY))
}

func drawGradient(c granio.Canvas, g gradient, pos granio.Position, base *image.NRGBA, clip image.Rectangle) {
	bounds := g.Size()
	width := scale.ToScreenCoordinate(c, bounds.Width)
	height := scale.ToScreenCoordinate(c, bounds.Height)
	x, y := scale.ToScreenCoordinate(c, pos.X), scale.ToScreenCoordinate(c, pos.Y)
	if width <= 0 || height <= 0 || !image.Rect(x, y, x+width, y+height).Overlaps(clip) {
		return
	}
	tex := g.Generate(width, height)
	drawTex(x, y, width, height, base, tex, clip, 1.0)
}

func drawImage(c granio.Canvas, img *canvas.Image, pos granio.Position, base *image.NRGBA, clip image.Rectangle) {
	bounds := img.Size()
	if bounds.IsZero() {
		return
	}
	width := scale.ToScreenCoordinate(c, bounds.Width)
	height := scale.ToScreenCoordinate(c, bounds.Height)
	scaledX, scaledY := scale.ToScreenCoordinate(c, pos.X), scale.ToScreenCoordinate(c, pos.Y)
	if width <= 0 || height <= 0 || !image.Rect(scaledX, scaledY, scaledX+width, scaledY+height).Overlaps(clip) {
		return
	}

	var rawImg image.Image
	if img.FillMode == canvas.ImageFillCover {
		if img.Image == nil {
			return
		}
		inner := img.Image.Bounds()
		imgAspect := img.Aspect()
		objAspect := float32(width) / float32(height)

		if objAspect > imgAspect {
			newHeight := float32(width) / imgAspect
			heightPad := (newHeight - float32(height)) / 2
			pixPad := int((heightPad / newHeight) * float32(inner.Dy()))

			inner = image.Rect(inner.Min.X, inner.Min.Y+pixPad, inner.Max.X, inner.Max.Y-pixPad)
		} else if objAspect < imgAspect {
			newWidth := float32(height) * imgAspect
			widthPad := (newWidth - float32(width)) / 2
			pixPad := int((widthPad / newWidth) * float32(inner.Dx()))

			inner = image.Rect(inner.Min.X+pixPad, inner.Min.Y, inner.Max.X-pixPad, inner.Max.Y)
		}

		if source, ok := img.Image.(*image.RGBA); ok {
			// The crop shares the original premultiplied pixels. The paint path
			// reads them using the crop's bounds and stride.
			rawImg = source.SubImage(inner)
		} else {
			subImg := image.NewRGBA(inner)
			draw.Copy(subImg, inner.Min, img.Image, inner, draw.Over, nil)
			rawImg = subImg
		}
	} else {
		rawImg = painter.PaintImage(img, c, width, height)
		if img.FillMode == canvas.ImageFillContain {
			imgAspect := img.Aspect()
			objAspect := float32(width) / float32(height)

			if objAspect > imgAspect {
				newWidth := int(float32(height) * imgAspect)
				scaledX += (width - newWidth) / 2
				width = newWidth
			} else if objAspect < imgAspect {
				newHeight := int(float32(width) / imgAspect)
				scaledY += (height - newHeight) / 2
				height = newHeight
			}
		}
	}

	cornerRadius := min(painter.GetMaximumRadius(bounds), img.CornerRadius)
	drawPixels(scaledX, scaledY, width, height, img.ScaleMode, base, rawImg, clip, img.Alpha(), cornerRadius*c.Scale())
}

func drawPixels(x, y, width, height int, mode canvas.ImageScale, base *image.NRGBA, origImg image.Image, clip image.Rectangle, alpha float64, radius float32) {
	if width <= 0 || height <= 0 || !image.Rect(x, y, x+width, y+height).Overlaps(clip) {
		return
	}
	if origImg.Bounds().Dx() == width && origImg.Bounds().Dy() == height && radius < 0.5 {
		// do not scale or duplicate image since not needed, draw directly
		drawTex(x, y, width, height, base, origImg, clip, alpha)
		return
	}

	scaledBounds := image.Rect(0, 0, width, height)
	scaledImg := image.NewNRGBA(scaledBounds)
	switch mode {
	case canvas.ImageScalePixels:
		draw.NearestNeighbor.Scale(scaledImg, scaledBounds, origImg, origImg.Bounds(), draw.Over, nil)
	case canvas.ImageScaleFastest:
		draw.ApproxBiLinear.Scale(scaledImg, scaledBounds, origImg, origImg.Bounds(), draw.Over, nil)
	default:
		if mode != canvas.ImageScaleSmooth {
			granio.LogError(fmt.Sprintf("Invalid canvas.ImageScale value (%d), using canvas.ImageScaleSmooth as default value", mode), nil)
		}
		painter.ScaleImageSmooth(scaledImg, origImg)
	}

	if radius > 0.5 {
		applyRoundedCorners(scaledImg, radius)
	}

	drawTex(x, y, width, height, base, scaledImg, clip, alpha)
}

func drawBezierCurve(c granio.Canvas, bezierCurve *canvas.BezierCurve, pos granio.Position, base *image.NRGBA, clip image.Rectangle) {
	pad := painter.VectorPad(bezierCurve)
	scaledWidth := scale.ToScreenCoordinate(c, bezierCurve.Size().Width+pad*2)
	scaledHeight := scale.ToScreenCoordinate(c, bezierCurve.Size().Height+pad*2)
	scaledX, scaledY := scale.ToScreenCoordinate(c, pos.X-pad), scale.ToScreenCoordinate(c, pos.Y-pad)
	bounds := clip.Intersect(image.Rect(scaledX, scaledY, scaledX+scaledWidth, scaledY+scaledHeight))

	raw := painter.DrawBezierCurve(bezierCurve, pad, func(in float32) float32 {
		return float32(math.Round(float64(in) * float64(c.Scale())))
	})

	drawClippedVector(base, bounds, raw, image.Pt(scaledX, scaledY))
}

func drawLine(c granio.Canvas, line *canvas.Line, pos granio.Position, base *image.NRGBA, clip image.Rectangle) {
	pad := painter.VectorPad(line)
	scaledWidth := scale.ToScreenCoordinate(c, line.Size().Width+pad*2)
	scaledHeight := scale.ToScreenCoordinate(c, line.Size().Height+pad*2)
	scaledX, scaledY := scale.ToScreenCoordinate(c, pos.X-pad), scale.ToScreenCoordinate(c, pos.Y-pad)
	bounds := clip.Intersect(image.Rect(scaledX, scaledY, scaledX+scaledWidth, scaledY+scaledHeight))

	raw := painter.DrawLine(line, pad, func(in float32) float32 {
		return float32(math.Round(float64(in) * float64(c.Scale())))
	})

	drawClippedVector(base, bounds, raw, image.Pt(scaledX, scaledY))
}

func drawClippedVector(base *image.NRGBA, bounds image.Rectangle, raw image.Image, origin image.Point) {
	src := raw.Bounds().Min.Add(bounds.Min.Sub(origin))
	draw.Draw(base, bounds, raw, src, draw.Over)
}

func drawTex(x, y, width, height int, base *image.NRGBA, tex image.Image, clip image.Rectangle, alpha float64) {
	outBounds := image.Rect(x, y, x+width, y+height)
	clippedBounds := clip.Intersect(outBounds)
	srcPt := tex.Bounds().Min.Add(clippedBounds.Min.Sub(outBounds.Min))
	if alpha == 1.0 {
		draw.Draw(base, clippedBounds, tex, srcPt, draw.Over)
	} else {
		mask := &image.Uniform{C: color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: uint8(float64(0xff) * alpha)}}
		draw.DrawMask(base, clippedBounds, tex, srcPt, mask, srcPt, draw.Over)
	}
}

func drawText(c granio.Canvas, text *canvas.Text, pos granio.Position, base *image.NRGBA, clip image.Rectangle, bitmaps *textBitmapCache) {
	bounds := text.MinSize()
	width := scale.ToScreenCoordinate(c, bounds.Width+painter.VectorPad(text)) // potentially italic overspill
	height := scale.ToScreenCoordinate(c, bounds.Height+painter.TextVectorPad) // space below for descenders / underline

	textColor := text.Color
	if textColor == nil {
		textColor = theme.Color(theme.ColorNameForeground)
	}

	size := text.Size()
	offsetX := float32(0)
	offsetY := float32(0)
	switch text.Alignment {
	case granio.TextAlignTrailing:
		offsetX = size.Width - bounds.Width
	case granio.TextAlignCenter:
		offsetX = (size.Width - bounds.Width) / 2
	}
	if size.Height > bounds.Height {
		offsetY = (size.Height - bounds.Height) / 2
	}
	scaledX := scale.ToScreenCoordinate(c, pos.X+offsetX)
	scaledY := scale.ToScreenCoordinate(c, pos.Y+offsetY)
	imgBounds := image.Rect(scaledX, scaledY, scaledX+width, scaledY+height)
	clippedBounds := clip.Intersect(imgBounds)
	if !clippedBounds.Empty() && width > 0 && height > 0 {
		txtImg := rasterizeText(c, text, textColor, width, height, bitmaps)
		srcPt := clippedBounds.Min.Sub(imgBounds.Min)
		draw.Draw(base, clippedBounds, txtImg, srcPt, draw.Over)
	}

	if text.TextStyle.Underline || text.TextStyle.Strikethrough {
		_, baseline := cache.GetFontMetrics(text.Text, text.TextSize, text.TextStyle, text.FontSource)
		line := canvas.NewLine(textColor)
		line.Resize(granio.NewSize(bounds.Width, 0))
		if text.TextStyle.Underline {
			underlinePos := granio.NewPos(pos.X, pos.Y+baseline+painter.UnderlineOffsetFromBaseline)
			drawLine(c, line, underlinePos, base, clip)
		}
		if text.TextStyle.Strikethrough {
			strikePos := granio.NewPos(pos.X, pos.Y+baseline*painter.StrikethroughToBaselineFactor)
			drawLine(c, line, strikePos, base, clip)
		}
	}
}

func rasterizeText(c granio.Canvas, text *canvas.Text, textColor color.Color, width, height int, bitmaps *textBitmapCache) *image.RGBA {
	face := painter.CachedFontFace(text.TextStyle, text.FontSource, text)
	pixelScale := c.Scale()
	var key textBitmapKey
	if bitmaps != nil {
		r, g, b, a := textColor.RGBA()
		key = textBitmapKey{text: text.Text, size: text.TextSize, style: text.TextStyle,
			font: face, width: width, height: height, scale: pixelScale,
			color: color.RGBA64{R: uint16(r), G: uint16(g), B: uint16(b), A: uint16(a)}}
		if cached := bitmaps.get(key); cached != nil {
			return cached
		}
	}
	bitmap := image.NewRGBA(image.Rect(0, 0, width, height))
	painter.DrawString(bitmap, text.Text, textColor, face.Fonts, text.TextSize, pixelScale, text.TextStyle)
	if bitmaps != nil {
		bitmaps.add(key, bitmap)
	}
	return bitmap
}

func drawRaster(c granio.Canvas, rast *canvas.Raster, pos granio.Position, base *image.NRGBA, clip image.Rectangle) {
	bounds := rast.Size()
	if bounds.IsZero() {
		return
	}
	width := scale.ToScreenCoordinate(c, bounds.Width)
	height := scale.ToScreenCoordinate(c, bounds.Height)
	scaledX, scaledY := scale.ToScreenCoordinate(c, pos.X), scale.ToScreenCoordinate(c, pos.Y)
	if width <= 0 || height <= 0 || !image.Rect(scaledX, scaledY, scaledX+width, scaledY+height).Overlaps(clip) {
		return
	}

	pix := rast.Generator(width, height)
	if pix == nil {
		return
	}
	if pix.Bounds().Size() != image.Pt(width, height) {
		drawPixels(scaledX, scaledY, width, height, rast.ScaleMode, base, pix, clip, 1.0, 0.0)
	} else {
		drawTex(scaledX, scaledY, width, height, base, pix, clip, 1.0)
	}
}

func drawOblongStroke(c granio.Canvas, obj granio.CanvasObject, width, height float32, shadow canvas.Shadow, pos granio.Position, base *image.NRGBA, clip image.Rectangle) {
	pad := painter.VectorPad(obj)
	scaledWidth := scale.ToScreenCoordinate(c, width+pad*2)
	scaledHeight := scale.ToScreenCoordinate(c, height+pad*2)
	scaledX, scaledY := scale.ToScreenCoordinate(c, pos.X-pad), scale.ToScreenCoordinate(c, pos.Y-pad)
	bounds := clip.Intersect(image.Rect(scaledX, scaledY, scaledX+scaledWidth, scaledY+scaledHeight))

	if painter.IsShadowVisible(shadow) {
		drawShadow(c, obj, granio.NewSize(width, height), shadow, pad, base, clip, pos)
	}
	if bounds.Empty() {
		return
	}
	raw := painter.DrawRectangle(obj.(*canvas.Rectangle), width, height, pad, func(in float32) float32 {
		return float32(math.Round(float64(in) * float64(c.Scale())))
	})
	drawClippedVector(base, bounds, raw, image.Pt(scaledX, scaledY))
}

func drawPolygon(c granio.Canvas, polygon *canvas.RegularPolygon, pos granio.Position, base *image.NRGBA, clip image.Rectangle) {
	pad := painter.VectorPad(polygon)
	scaledWidth := scale.ToScreenCoordinate(c, polygon.Size().Width+pad*2)
	scaledHeight := scale.ToScreenCoordinate(c, polygon.Size().Height+pad*2)
	scaledX, scaledY := scale.ToScreenCoordinate(c, pos.X-pad), scale.ToScreenCoordinate(c, pos.Y-pad)
	bounds := clip.Intersect(image.Rect(scaledX, scaledY, scaledX+scaledWidth, scaledY+scaledHeight))

	raw := painter.DrawPolygon(polygon, pad, func(in float32) float32 {
		return float32(math.Round(float64(in) * float64(c.Scale())))
	})

	drawClippedVector(base, bounds, raw, image.Pt(scaledX, scaledY))
}

func drawArbitraryPolygon(c granio.Canvas, polygon *canvas.ArbitraryPolygon, pos granio.Position, base *image.NRGBA, clip image.Rectangle) {
	pad := painter.VectorPad(polygon)
	scaledWidth := scale.ToScreenCoordinate(c, polygon.Size().Width+pad*2)
	scaledHeight := scale.ToScreenCoordinate(c, polygon.Size().Height+pad*2)
	scaledX, scaledY := scale.ToScreenCoordinate(c, pos.X-pad), scale.ToScreenCoordinate(c, pos.Y-pad)
	bounds := clip.Intersect(image.Rect(scaledX, scaledY, scaledX+scaledWidth, scaledY+scaledHeight))

	raw := painter.DrawArbitraryPolygon(polygon, pad, func(in float32) float32 {
		return float32(math.Round(float64(in) * float64(c.Scale())))
	})

	drawClippedVector(base, bounds, raw, image.Pt(scaledX, scaledY))
}

func drawRectangle(c granio.Canvas, rect *canvas.Rectangle, pos granio.Position, base *image.NRGBA, clip image.Rectangle) {
	r := rect.EffectiveCornerRadii(rect.Size())
	drawOblong(c, rect, rect.FillColor, rect.StrokeColor, rect.StrokeWidth, r.TopRight, r.TopLeft, r.BottomRight, r.BottomLeft, rect.Aspect, rect.Shadow, pos, base, clip)
}

func drawOblong(c granio.Canvas, obj granio.CanvasObject, fill, stroke color.Color, strokeWidth, topRightRadius, topLeftRadius, bottomRightRadius, bottomLeftRadius, aspect float32, shadow canvas.Shadow, pos granio.Position, base *image.NRGBA, clip image.Rectangle) {
	width, height := obj.Size().Components()
	if aspect != 0 {
		frameAspect := width / height

		xPad, yPad := float32(0), float32(0)
		if frameAspect > aspect {
			newWidth := height * aspect
			xPad = (width - newWidth) / 2
			width = newWidth
		} else if frameAspect < aspect {
			newHeight := width / aspect
			yPad = (height - newHeight) / 2
			height = newHeight
		}

		pos = pos.AddXY(xPad, yPad)
	}

	if (stroke != nil && strokeWidth > 0) || topRightRadius != 0 || topLeftRadius != 0 || bottomRightRadius != 0 || bottomLeftRadius != 0 { // use a rasterizer if there is a stroke or radius
		drawOblongStroke(c, obj, width, height, shadow, pos, base, clip)
		return
	}

	scaledWidth := scale.ToScreenCoordinate(c, width)
	scaledHeight := scale.ToScreenCoordinate(c, height)
	scaledX, scaledY := scale.ToScreenCoordinate(c, pos.X), scale.ToScreenCoordinate(c, pos.Y)
	bounds := clip.Intersect(image.Rect(scaledX, scaledY, scaledX+scaledWidth, scaledY+scaledHeight))

	if painter.IsShadowVisible(shadow) {
		drawShadow(c, obj, granio.NewSize(width, height), shadow, 0, base, clip, pos)
	}

	if fillRectFastPath(base, bounds, fill) {
		return
	}

	blendSolidRect(base, bounds, fill)
}

// blendSolidRect implements source-over for a uniform input. It keeps the same
// 16-bit premultiplied arithmetic as image/draw, including on translucent targets.
func blendSolidRect(base *image.NRGBA, bounds image.Rectangle, fill color.Color) {
	bounds = bounds.Intersect(base.Bounds())
	if fill == nil || bounds.Empty() {
		return
	}
	const opaque = 0xffff
	r, g, b, a := fill.RGBA()
	remaining := opaque - a
	opaqueScale := remaining * 257
	// Amortize conversion over large overlays. Small shapes use the same
	// arithmetic directly, so drawing a small patch needs no table setup.
	var lookup [3][256]uint8
	useTable := bounds.Dx()*bounds.Dy() >= 4096
	if useTable {
		for i := range lookup[0] {
			dst := uint32(i) * opaqueScale / opaque
			lookup[0][i] = uint8((dst + r) >> 8)
			lookup[1][i] = uint8((dst + g) >> 8)
			lookup[2][i] = uint8((dst + b) >> 8)
		}
	}
	offset := base.PixOffset(bounds.Min.X, bounds.Min.Y)
	rowBytes := bounds.Dx() * 4
	for y := 0; y < bounds.Dy(); y++ {
		row := base.Pix[offset : offset+rowBytes]
		for len(row) != 0 {
			pixel := row[:4:4]
			row = row[4:]
			if pixel[3] == 255 {
				if useTable {
					pixel[0], pixel[1], pixel[2] = lookup[0][pixel[0]], lookup[1][pixel[1]], lookup[2][pixel[2]]
					continue
				}
				pixel[0] = uint8((uint32(pixel[0])*opaqueScale/opaque + r) >> 8)
				pixel[1] = uint8((uint32(pixel[1])*opaqueScale/opaque + g) >> 8)
				pixel[2] = uint8((uint32(pixel[2])*opaqueScale/opaque + b) >> 8)
				continue
			}
			dstAlpha := uint32(pixel[3])
			outAlpha := (dstAlpha*257)*remaining/opaque + a
			pixel[0] = blendSolidChannel(pixel[0], r, dstAlpha, remaining, outAlpha)
			pixel[1] = blendSolidChannel(pixel[1], g, dstAlpha, remaining, outAlpha)
			pixel[2] = blendSolidChannel(pixel[2], b, dstAlpha, remaining, outAlpha)
			pixel[3] = uint8(outAlpha >> 8)
		}
		offset += base.Stride
	}
}

func blendSolidChannel(dst uint8, src, dstAlpha, remaining, outAlpha uint32) uint8 {
	premultiplied := uint32(dst) * 257 * dstAlpha / 255
	value := uint32(uint16(premultiplied*remaining/0xffff + src))
	if outAlpha != 0 {
		value = value * 0xffff / outAlpha
	}
	return uint8(value >> 8)
}

// fillRectFastPath writes an opaque fill directly into base.Pix.
// It returns false if the fill is non-opaque or bounds are empty.
func fillRectFastPath(base *image.NRGBA, bounds image.Rectangle, fill color.Color) bool {
	if fill == nil || bounds.Empty() {
		return false
	}

	r, g, b, a := fill.RGBA()
	if a != 0xffff {
		return false
	}

	nr, ng, nb := uint8(r>>8), uint8(g>>8), uint8(b>>8) //gosec:disable G115 -- RGBA() components are 16-bit, >>8 always fits uint8
	stride := base.Stride
	rowBytes := bounds.Dx() * 4
	offset := base.PixOffset(bounds.Min.X, bounds.Min.Y)
	row := base.Pix[offset : offset+rowBytes]
	row[0], row[1], row[2], row[3] = nr, ng, nb, 0xff
	for filled := 4; filled < len(row); {
		filled += copy(row[filled:], row[:filled])
	}
	for y := bounds.Min.Y + 1; y < bounds.Max.Y; y++ {
		offset += stride
		copy(base.Pix[offset:offset+rowBytes], row)
	}
	return true
}

func drawEllipse(c granio.Canvas, ellipse *canvas.Ellipse, pos granio.Position, base *image.NRGBA, clip image.Rectangle) {
	// when rotated, the ellipse needs more space
	// add half the difference between width and height as padding
	// with padding the final size is a square
	width, height := ellipse.Size().Components()
	rotPad := float32(math.Abs(float64(width)-float64(height)) / 2)
	xPad, yPad := float32(0), float32(0)

	if width > height {
		yPad = rotPad
	} else {
		xPad = rotPad
	}

	if painter.IsShadowVisible(ellipse.Shadow) {
		drawShadow(c, ellipse, ellipse.Size(), ellipse.Shadow, painter.VectorPad(ellipse), base, clip, pos)
	}

	pos = pos.AddXY(xPad, yPad)
	pad := painter.VectorPad(ellipse) + rotPad

	scaledWidth := scale.ToScreenCoordinate(c, ellipse.Size().Width+pad*2)
	scaledHeight := scale.ToScreenCoordinate(c, ellipse.Size().Height+pad*2)
	scaledX, scaledY := scale.ToScreenCoordinate(c, pos.X-pad), scale.ToScreenCoordinate(c, pos.Y-pad)
	bounds := clip.Intersect(image.Rect(scaledX, scaledY, scaledX+scaledWidth, scaledY+scaledHeight))

	raw := painter.DrawEllipse(ellipse, pad, func(in float32) float32 {
		return float32(math.Round(float64(in) * float64(c.Scale())))
	})

	drawClippedVector(base, bounds, raw, image.Pt(scaledX, scaledY))
}

func drawShadow(c granio.Canvas, obj granio.CanvasObject, objSize granio.Size, shadow canvas.Shadow, pad float32, base *image.NRGBA, clip image.Rectangle, pos granio.Position) {
	shadowOffset := shadow.Offset
	shadowBlurRadius := shadow.BlurRadius
	shadowSpread := shadow.Spread
	shadowVariant := shadow.Variant
	shadowColor := shadow.Color

	var shadowRaw *image.RGBA
	var maskRaw *image.RGBA

	vPad := pad + shadowBlurRadius
	if shadowSpread < 0 {
		vPad -= shadowSpread
	}

	switch o := obj.(type) {
	case *canvas.Rectangle:
		shadowRaw = painter.DrawRectangle(&canvas.Rectangle{
			FillColor:               shadowColor,
			CornerRadius:            o.CornerRadius,
			TopRightCornerRadius:    o.TopRightCornerRadius,
			TopLeftCornerRadius:     o.TopLeftCornerRadius,
			BottomRightCornerRadius: o.BottomRightCornerRadius,
			BottomLeftCornerRadius:  o.BottomLeftCornerRadius,
		}, max(objSize.Width+2*shadowSpread, 0), max(objSize.Height+2*shadowSpread, 0), vPad, func(in float32) float32 {
			return float32(math.Round(float64(in) * float64(c.Scale())))
		})
		maskRaw = painter.DrawRectangle(&canvas.Rectangle{
			FillColor:               color.Opaque,
			CornerRadius:            o.CornerRadius,
			TopRightCornerRadius:    o.TopRightCornerRadius,
			TopLeftCornerRadius:     o.TopLeftCornerRadius,
			BottomRightCornerRadius: o.BottomRightCornerRadius,
			BottomLeftCornerRadius:  o.BottomLeftCornerRadius,
		}, objSize.Width, objSize.Height, vPad+shadowSpread, func(in float32) float32 {
			return float32(math.Round(float64(in) * float64(c.Scale())))
		})
	case *canvas.Circle:
		shadowCircle := &canvas.Circle{FillColor: shadowColor}
		shadowCircle.Resize(objSize.AddWidthHeight(2*shadowSpread, 2*shadowSpread))
		shadowRaw = painter.DrawCircle(shadowCircle, vPad, func(in float32) float32 {
			return float32(math.Round(float64(in) * float64(c.Scale())))
		})
		maskCircle := &canvas.Circle{FillColor: color.Opaque}
		maskCircle.Resize(objSize)
		maskRaw = painter.DrawCircle(maskCircle, vPad+shadowSpread, func(in float32) float32 {
			return float32(math.Round(float64(in) * float64(c.Scale())))
		})
	case *canvas.Ellipse:
		shadowEllipse := &canvas.Ellipse{FillColor: shadowColor}
		shadowEllipse.Resize(objSize.AddWidthHeight(2*shadowSpread, 2*shadowSpread))
		shadowRaw = painter.DrawEllipse(shadowEllipse, vPad, func(in float32) float32 {
			return float32(math.Round(float64(in) * float64(c.Scale())))
		})
		maskEllipse := &canvas.Ellipse{FillColor: color.Opaque}
		maskEllipse.Resize(objSize)
		maskRaw = painter.DrawEllipse(maskEllipse, vPad+shadowSpread, func(in float32) float32 {
			return float32(math.Round(float64(in) * float64(c.Scale())))
		})
	}

	startX := pos.X + float32(shadowOffset.X) - shadowSpread - vPad
	startY := pos.Y + float32(shadowOffset.Y) - shadowSpread - vPad

	screenStartX := scale.ToScreenCoordinate(c, startX)
	screenStartY := scale.ToScreenCoordinate(c, startY)

	blurred := blur.Gaussian(shadowRaw, float64(shadowBlurRadius*c.Scale()))
	destRect := image.Rect(screenStartX, screenStartY, screenStartX+blurred.Bounds().Dx(), screenStartY+blurred.Bounds().Dy())
	shadowBounds := clip.Intersect(destRect)

	if shadowBounds.Empty() {
		return
	}

	// If DropShadow, subtract object from shadow
	if shadowVariant == canvas.DropShadow {
		dx := screenStartX - scale.ToScreenCoordinate(c, pos.X-shadowSpread-vPad)
		dy := screenStartY - scale.ToScreenCoordinate(c, pos.Y-shadowSpread-vPad)

		var fill, strokeCol color.Color
		var strokeWidth float32
		switch o := obj.(type) {
		case *canvas.Rectangle:
			fill, strokeCol, strokeWidth = o.FillColor, o.StrokeColor, o.StrokeWidth
		case *canvas.Circle:
			fill, strokeCol, strokeWidth = o.FillColor, o.StrokeColor, o.StrokeWidth
		case *canvas.Ellipse:
			fill, strokeCol, strokeWidth = o.FillColor, o.StrokeColor, o.StrokeWidth
		}

		var objAlpha float32
		if fill != nil {
			_, _, _, a := fill.RGBA()
			objAlpha = float32(a) / math.MaxUint16
		}
		if strokeCol != nil && strokeWidth > 0 {
			_, _, _, a := strokeCol.RGBA()
			sa := float32(a) / math.MaxUint16
			if sa > objAlpha {
				objAlpha = sa
			}
		}

		for y := 0; y < blurred.Bounds().Dy(); y++ {
			for x := 0; x < blurred.Bounds().Dx(); x++ {
				mx := x + dx
				my := y + dy

				_, _, _, maskA := maskRaw.At(mx, my).RGBA()
				if maskA > 0 {
					pixel := blurred.RGBAAt(x, y)
					cVal := float32(maskA) / math.MaxUint16
					den := 1.0 - cVal*objAlpha
					var invMA float32
					if den <= 0 {
						invMA = 0
					} else {
						invMA = (1.0 - cVal) / den
					}
					pixel.R = uint8(float32(pixel.R) * invMA)
					pixel.G = uint8(float32(pixel.G) * invMA)
					pixel.B = uint8(float32(pixel.B) * invMA)
					pixel.A = uint8(float32(pixel.A) * invMA)
					blurred.SetRGBA(x, y, pixel)
				}
			}
		}
	}

	srcPt := image.Point{X: shadowBounds.Min.X - screenStartX, Y: shadowBounds.Min.Y - screenStartY}
	draw.Draw(base, shadowBounds, blurred, srcPt, draw.Over)
}

// applyRoundedCorners rounds the corners of the image in-place
func applyRoundedCorners(img image.Image, radius float32) {
	aaWidth := float32(0.5)
	outerR2 := (radius + aaWidth) * (radius + aaWidth)
	innerR2 := (radius - aaWidth) * (radius - aaWidth)

	applyCorner := func(startX, endX, startY, endY int, cx, cy float32) {
		for y := startY; y < endY; y++ {
			for x := startX; x < endX; x++ {
				dx := float32(x) - cx
				dy := float32(y) - cy
				dist2 := dx*dx + dy*dy

				switch i := img.(type) {
				case *image.NRGBA:
					off := i.PixOffset(x, y)
					switch {
					case dist2 >= outerR2:
						i.Pix[off+3] = 0 // Fully transparent
					case dist2 > innerR2:
						// Linear falloff based on squared distance
						t := (outerR2 - dist2) / (outerR2 - innerR2) // t ranges from 0 to 1
						i.Pix[off+3] = uint8(float32(i.Pix[off+3]) * t)
					}
				case *image.RGBA:
					off := i.PixOffset(x, y)
					switch {
					case dist2 >= outerR2:
						i.Pix[off+0] = 0
						i.Pix[off+1] = 0
						i.Pix[off+2] = 0
						i.Pix[off+3] = 0
					case dist2 > innerR2:
						t := (outerR2 - dist2) / (outerR2 - innerR2)
						i.Pix[off+0] = uint8(float32(i.Pix[off+0]) * t)
						i.Pix[off+1] = uint8(float32(i.Pix[off+1]) * t)
						i.Pix[off+2] = uint8(float32(i.Pix[off+2]) * t)
						i.Pix[off+3] = uint8(float32(i.Pix[off+3]) * t)
					}
				}
			}
		}
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	rInt := int(math.Ceil(float64(radius)))
	r := minInt(rInt, minInt(w, h))
	minX, minY := bounds.Min.X, bounds.Min.Y
	maxX, maxY := bounds.Max.X, bounds.Max.Y

	// Top-left
	applyCorner(minX, minX+r, minY, minY+r, float32(minX)+radius, float32(minY)+radius)

	// Top-right
	applyCorner(maxX-r, maxX, minY, minY+r, float32(maxX)-radius, float32(minY)+radius)

	// Bottom-left
	applyCorner(minX, minX+r, maxY-r, maxY, float32(minX)+radius, float32(maxY)-radius)

	// Bottom-right
	applyCorner(maxX-r, maxX, maxY-r, maxY, float32(maxX)-radius, float32(maxY)-radius)
}

func minInt(x, y int) int {
	if x < y {
		return x
	}
	return y
}
