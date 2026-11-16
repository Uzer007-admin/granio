package software

import (
	"image"
	"math"
	"time"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/internal/painter"
	"github.com/Uzer007-admin/granio/v3/internal/scale"
)

type blurRGB struct{ r, g, b uint8 }

type softwareMaterialKey struct {
	material *canvas.Material
	region   image.Rectangle
	frame    image.Rectangle
	revision uint64
	base     uint32
	radius   float32
	step     int
}

type blurMaskKey struct {
	size    image.Point
	corners granio.CornerRadii
}

type blurSample struct {
	lo, hi int32
	weight uint32
}

type softwareMaterial struct {
	budget   painter.SoftwareBlurBudget
	key      softwareMaterialKey
	pixels   []blurRGB
	work     []blurRGB
	mask     []uint8
	shape    blurMaskKey
	x        []blurSample
	xStep    int
	captures uint64
}

func (p *FramePainter) paintMaterial(m *canvas.Material, pos granio.Position) {
	allowed := m.EffectsAllowed()
	if allowed && m.CanPaintTint() {
		drawRectangle(p.canvas, m.TintRectangle(), pos, p.frame, p.clip)
		if m.Tint.A == 255 {
			m.SetEffect(canvas.MaterialOpaque)
		} else {
			m.SetEffect(canvas.MaterialTint)
		}
		return
	}
	if allowed && m.SoftwareBlur && m.BlurRadius > 0 {
		start := time.Now()
		size := m.Size()
		x := scale.ToScreenCoordinate(p.canvas, pos.X)
		y := scale.ToScreenCoordinate(p.canvas, pos.Y)
		shape := image.Rect(x, y, scale.ToScreenCoordinate(p.canvas, pos.X+size.Width), scale.ToScreenCoordinate(p.canvas, pos.Y+size.Height))
		drawable := shape.Intersect(p.clip)
		if !drawable.Empty() {
			if step := p.material.budget.Begin(m, shape.Dx(), shape.Dy(), 4, start); step != 0 {
				p.material.draw(p, m, shape, drawable, int(step))
				p.material.budget.End(time.Since(start))
				m.SetEffect(canvas.MaterialSoftwareBlur)
				return
			}
		}
	}
	fallback := m.FallbackRectangle()
	drawRectangle(p.canvas, fallback, pos, p.frame, p.clip)
	_, _, _, alpha := fallback.FillColor.RGBA()
	mode := canvas.MaterialOpaque
	if alpha < 65535 {
		mode = canvas.MaterialTint
	}
	m.SetEffect(mode)
}

func (s *softwareMaterial) draw(p *FramePainter, m *canvas.Material, shape, drawable image.Rectangle, step int) {
	source := shape.Intersect(p.frame.Bounds())
	w := (source.Dx() + step - 1) / step
	h := (source.Dy() + step - 1) / step
	radius := m.BlurRadius * p.canvas.Scale()
	key := softwareMaterialKey{m, source, p.frame.Bounds(), m.BackdropRevision(), p.background, radius, step}
	if m.Live || !m.TrackBackdrop || key != s.key {
		s.captures++
		s.pixels = resizeScratch(s.pixels, w*h)
		s.work = resizeScratch(s.work, w*h)
		for y := range h {
			y0, y1 := source.Min.Y+y*step, min(source.Max.Y-1, source.Min.Y+y*step+step-1)
			row := s.pixels[y*w : (y+1)*w]
			for x := range row {
				x0, x1 := source.Min.X+x*step, min(source.Max.X-1, source.Min.X+x*step+step-1)
				a := p.frame.NRGBAAt((3*x0+x1)/4, (3*y0+y1)/4)
				b := p.frame.NRGBAAt((x0+3*x1)/4, (3*y0+y1)/4)
				c := p.frame.NRGBAAt((3*x0+x1)/4, (y0+3*y1)/4)
				d := p.frame.NRGBAAt((x0+3*x1)/4, (y0+3*y1)/4)
				row[x] = blurRGB{
					uint8((uint32(a.R) + uint32(b.R) + uint32(c.R) + uint32(d.R) + 2) >> 2),
					uint8((uint32(a.G) + uint32(b.G) + uint32(c.G) + uint32(d.G) + 2) >> 2),
					uint8((uint32(a.B) + uint32(b.B) + uint32(c.B) + uint32(d.B) + 2) >> 2),
				}
			}
		}
		box := max(1, int(math.Round(float64(radius)/float64(2*step))))
		blurRows(s.work, s.pixels, w, h, box, 1, w)
		blurRows(s.pixels, s.work, h, w, box, w, 1)
		s.key = key
	}
	s.prepareMask(shape.Size(), m.EffectiveCornerRadii(), p.canvas.Scale())
	if len(s.x) != source.Dx() || s.xStep != step {
		s.x = resizeScratch(s.x, source.Dx())
		for x := range s.x {
			s.x[x] = sampleBlur(x, step, w)
		}
		s.xStep = step
	}
	tint := m.Tint
	tintAlpha := uint32(tint.A)
	blurMix := uint32(min(255, radius*255/float32(step)))
	for y := drawable.Min.Y; y < drawable.Max.Y; y++ {
		sy := sampleBlur(y-source.Min.Y, step, h)
		row0, row1 := s.pixels[int(sy.lo)*w:][:w], s.pixels[int(sy.hi)*w:][:w]
		mask := s.mask[(y-shape.Min.Y)*shape.Dx()+drawable.Min.X-shape.Min.X:][:drawable.Dx()]
		row := p.frame.Pix[p.frame.PixOffset(drawable.Min.X, y):][:4*drawable.Dx()]
		for i, sx := range s.x[drawable.Min.X-source.Min.X:][:drawable.Dx()] {
			coverage := uint32(mask[i])
			if coverage == 0 {
				continue
			}
			a, b, c, d := row0[sx.lo], row0[sx.hi], row1[sx.lo], row1[sx.hi]
			dst := row[i*4 : i*4+4]
			dst[0] = mixBlurPixel(dst[0], interpolateBlur(a.r, b.r, c.r, d.r, sx.weight, sy.weight), tint.R, blurMix, tintAlpha, coverage)
			dst[1] = mixBlurPixel(dst[1], interpolateBlur(a.g, b.g, c.g, d.g, sx.weight, sy.weight), tint.G, blurMix, tintAlpha, coverage)
			dst[2] = mixBlurPixel(dst[2], interpolateBlur(a.b, b.b, c.b, d.b, sx.weight, sy.weight), tint.B, blurMix, tintAlpha, coverage)
		}
	}
}

func resizeScratch[T any](buf []T, length int) []T {
	if cap(buf) < length {
		return make([]T, length)
	}
	return buf[:length]
}

func blurRows(dst, src []blurRGB, length, rows, radius, skip, stride int) {
	size := uint32(2*radius + 1)
	for row := range rows {
		offset := row * stride
		var sr, sg, sb uint32
		for i := -radius; i <= radius; i++ {
			c := src[offset+min(max(i, 0), length-1)*skip]
			sr, sg, sb = sr+uint32(c.r), sg+uint32(c.g), sb+uint32(c.b)
		}
		for i := range length {
			dst[offset+i*skip] = blurRGB{uint8(sr / size), uint8(sg / size), uint8(sb / size)}
			in := src[offset+min(i+radius+1, length-1)*skip]
			out := src[offset+max(i-radius, 0)*skip]
			sr, sg, sb = sr+uint32(in.r)-uint32(out.r), sg+uint32(in.g)-uint32(out.g), sb+uint32(in.b)-uint32(out.b)
		}
	}
}

func sampleBlur(pixel, step, count int) blurSample {
	cell := (2*pixel + 1 - step) * 128 / step
	cell = max(0, min(cell, (count-1)*256))
	i := cell >> 8
	return blurSample{lo: int32(i), hi: int32(min(i+1, count-1)), weight: uint32(cell & 255)}
}

func interpolateBlur(a, b, c, d uint8, x, y uint32) uint32 {
	top := uint32(a)*(256-x) + uint32(b)*x
	bottom := uint32(c)*(256-x) + uint32(d)*x
	return (top*(256-y) + bottom*y + 32768) >> 16
}

func mixBlurPixel(original uint8, filtered uint32, tint uint8, blurMix, tintAlpha, coverage uint32) uint8 {
	original32 := uint32(original)
	if blurMix != 255 {
		filtered = (filtered*blurMix + original32*(255-blurMix) + 127) / 255
	}
	if tintAlpha != 0 {
		filtered = (filtered*(255-tintAlpha) + uint32(tint)*tintAlpha + 127) / 255
	}
	if coverage == 255 {
		return uint8(filtered)
	}
	return uint8((filtered*coverage + original32*(255-coverage) + 127) / 255)
}

func (s *softwareMaterial) prepareMask(size image.Point, r granio.CornerRadii, scale float32) {
	r.TopLeft *= scale
	r.TopRight *= scale
	r.BottomLeft *= scale
	r.BottomRight *= scale
	r = r.Normalized(granio.NewSize(float32(size.X), float32(size.Y)))
	key := blurMaskKey{size, r}
	if key == s.shape {
		return
	}
	s.mask = resizeScratch(s.mask, size.X*size.Y)
	for i := range s.mask {
		s.mask[i] = 255
	}
	maskCorner := func(radius float32, right, bottom bool) {
		if radius <= 0 {
			return
		}
		extent := int(math.Ceil(float64(radius)))
		for y := range extent {
			for x := range extent {
				dx, dy := radius-float32(x)-0.5, radius-float32(y)-0.5
				distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
				alpha := min(1, max(0, radius-distance+0.5))
				if alpha < 1 {
					col, row := x, y
					if right {
						col = size.X - 1 - x
					}
					if bottom {
						row = size.Y - 1 - y
					}
					s.mask[row*size.X+col] = uint8(alpha*255 + 0.5)
				}
			}
		}
	}
	maskCorner(r.TopLeft, false, false)
	maskCorner(r.TopRight, true, false)
	maskCorner(r.BottomRight, true, true)
	maskCorner(r.BottomLeft, false, true)
	s.shape = key
}
