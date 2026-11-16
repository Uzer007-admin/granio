package software

import (
	"image"
	"image/color"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/driver/desktop"
	"github.com/Uzer007-admin/granio/v3/internal"
	"github.com/Uzer007-admin/granio/v3/internal/scale"
	"github.com/Uzer007-admin/granio/v3/theme"
)

// FramePainter рисует сохранённое дерево объектов в переиспользуемый CPU-кадр.
// Его методы вызываются в потоке окна, как и методы GL-painter.
type FramePainter struct {
	canvas     *frameCanvas
	frame      *image.NRGBA
	clip       image.Rectangle
	text       textBitmapCache
	material   softwareMaterial
	background uint32
}

type frameCanvas struct {
	granio.Canvas
	framebufferScale float32
}

func (c *frameCanvas) Scale() float32 { return c.Canvas.Scale() * c.framebufferScale }

func NewFramePainter(c granio.Canvas) *FramePainter {
	return &FramePainter{canvas: &frameCanvas{Canvas: c, framebufferScale: 1}}
}
func (*FramePainter) Init()                    {}
func (*FramePainter) Free(granio.CanvasObject) {}
func (*FramePainter) MaxTextureSize() int      { return 16384 }
func (p *FramePainter) SetFrameBufferScale(factor float32) {
	if factor > 0 {
		p.canvas.framebufferScale = factor
	}
}

func (p *FramePainter) SetOutputSize(width, height int) {
	// Проверяем размер до умножения и ограничиваем выделение памяти.
	const maxPixels = (256 << 20) / 4
	if width <= 0 || height <= 0 || width > maxPixels/height {
		p.frame = nil
		p.clip = image.Rectangle{}
		return
	}
	bounds := image.Rect(0, 0, width, height)
	if p.frame == nil || p.frame.Bounds() != bounds {
		p.frame = image.NewNRGBA(bounds)
	}
	p.clip = bounds
}

func (p *FramePainter) Clear() {
	if p.frame == nil {
		return
	}
	background := color.NRGBAModel.Convert(theme.Color(theme.ColorNameBackground)).(color.NRGBA)
	background.A = 255
	p.background = uint32(background.R)<<16 | uint32(background.G)<<8 | uint32(background.B)
	fillRectFastPath(p.frame, p.frame.Bounds(), background)
	p.clip = p.frame.Bounds()
	p.material.budget.ResetFrame()
}

func (p *FramePainter) Paint(obj granio.CanvasObject, pos granio.Position, _ granio.Size, _ *internal.ClipItem) {
	if p.frame == nil || p.clip.Empty() {
		return
	}
	if label, ok := obj.(*canvas.Text); ok {
		drawText(p.canvas, label, pos, p.frame, p.clip, &p.text)
		return
	}
	if material, ok := obj.(*canvas.Material); ok {
		p.paintMaterial(material, pos)
		return
	}
	paintObject(p.canvas, obj, pos, p.frame, p.clip)
}

func (p *FramePainter) StartClipping(pos granio.Position, size granio.Size) {
	if p.frame == nil {
		return
	}
	p.clip = image.Rect(scale.ToScreenCoordinate(p.canvas, pos.X), scale.ToScreenCoordinate(p.canvas, pos.Y),
		scale.ToScreenCoordinate(p.canvas, pos.X+size.Width), scale.ToScreenCoordinate(p.canvas, pos.Y+size.Height)).Intersect(p.frame.Bounds())
}

func (p *FramePainter) StopClipping() {
	if p.frame != nil {
		p.clip = p.frame.Bounds()
	}
}

// Frame возвращает буфер до следующего рисования; Capture создаёт копию.
func (p *FramePainter) Frame() *image.NRGBA { return p.frame }

// MaterialStatistics reports the live scratch buffers and sampled regions.
func (p *FramePainter) MaterialStatistics() map[string]uint64 {
	m := &p.material
	return map[string]uint64{
		"captures":     m.captures,
		"pool_bytes":   uint64(cap(m.pixels)*3 + cap(m.work)*3 + cap(m.mask) + cap(m.x)*12),
		"budget_bytes": 2 << 20,
	}
}

func (p *FramePainter) Capture(granio.Canvas) image.Image {
	if p.frame == nil {
		return nil
	}
	frame := image.NewNRGBA(p.frame.Bounds())
	copy(frame.Pix, p.frame.Pix)
	return frame
}

func (p *FramePainter) CaptureCanvas(c granio.Canvas, paint func()) (image.Image, error) {
	if paint == nil || c == nil || c.Size().Width <= 0 || c.Size().Height <= 0 {
		return nil, desktop.ErrCaptureUnavailable
	}
	paint()
	if p.frame == nil {
		return nil, desktop.ErrCaptureUnavailable
	}
	return p.Capture(c), nil
}
