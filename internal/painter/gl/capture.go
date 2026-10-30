package gl

import (
	"image"
	"image/color"

	"github.com/Uzer007-admin/gl/capabilities"
	"github.com/Uzer007-admin/gl/readback"

	"github.com/Uzer007-admin/granio/v2"
)

type captureImage struct {
	pix           []uint8
	width, height int

	leftPad, rightPad, topPad, bottomPad int
	preserveAlpha                        bool
}

func (c *captureImage) SubImage(inner image.Rectangle) *captureImage {
	inner = inner.Intersect(c.Bounds())
	if inner.Empty() {
		return &captureImage{}
	}
	return &captureImage{
		pix:    c.pix,
		width:  c.width,
		height: c.height,

		leftPad:       inner.Min.X,
		topPad:        inner.Min.Y,
		rightPad:      c.width - inner.Max.X,
		bottomPad:     c.height - inner.Max.Y,
		preserveAlpha: c.preserveAlpha,
	}
}

func (*captureImage) ColorModel() color.Model {
	return color.RGBAModel
}

func (c *captureImage) Bounds() image.Rectangle {
	return image.Rect(c.leftPad, c.topPad, c.width-c.rightPad, c.height-c.bottomPad)
}

func (c *captureImage) At(x, y int) color.Color {
	if !image.Pt(x, y).In(c.Bounds()) {
		return color.RGBA{}
	}
	start := ((c.height-y-1)*c.width + x) * 4
	// The framebuffer colour is already composited against the window background, but blending
	// leaves an alpha below 255 wherever a shadow or an anti-aliased edge was drawn.
	alpha := uint8(0xff)
	if c.preserveAlpha {
		alpha = c.pix[start+3]
	}
	return color.RGBA{R: c.pix[start], G: c.pix[start+1], B: c.pix[start+2], A: alpha}
}

func (p *painter) Capture(c granio.Canvas) image.Image {
	return p.captureBuffer(c, front)
}

type captureBackend interface {
	Capture(granio.Canvas, func()) (image.Image, error)
	Read(granio.Canvas, uint32) (image.Image, error)
}

func (p *painter) captureAPI() (captureBackend, error) {
	if p.capture == nil {
		backend, err := newCaptureBackend(p)
		if err != nil {
			return nil, err
		}
		p.capture = backend
	}
	return p.capture, nil
}

// CaptureCanvas renders an independent snapshot. Native targets do not depend
// on the visibility, minimization or current size of the window's back buffer.
func (p *painter) CaptureCanvas(c granio.Canvas, draw func()) (img image.Image, err error) {
	err = capabilities.ErrNoContext
	p.contextProvider.RunWithContext(func() { img, err = p.captureCanvasCurrent(c, draw) })
	return
}

func (p *painter) captureCanvasCurrent(c granio.Canvas, draw func()) (image.Image, error) {
	size := c.Size()
	width, height := c.PixelCoordinateForPosition(granio.NewPos(size.Width, size.Height))
	if _, err := readback.RGBA8Size(width, height); err != nil {
		return nil, err
	}
	backend, err := p.captureAPI()
	if err != nil {
		return nil, err
	}
	if backend != nil {
		return backend.Capture(c, draw)
	}
	draw()
	return p.captureBufferChecked(c, back)
}

// CaptureBack reads a freshly rendered canvas before presentation. The desktop
// driver uses this path because GLX front buffers can be clipped or unavailable
// under compositors, occlusion and headless X servers.
func (p *painter) CaptureBack(c granio.Canvas) image.Image { return p.captureBuffer(c, back) }

func (p *painter) captureBuffer(c granio.Canvas, buffer uint32) image.Image {
	img, err := p.captureBufferChecked(c, buffer)
	if err != nil {
		granio.LogError("Failed to capture GL canvas", err)
	}
	return img
}

func (p *painter) captureBufferChecked(c granio.Canvas, buffer uint32) (image.Image, error) {
	var img image.Image
	err := capabilities.ErrNoContext
	p.contextProvider.RunWithContext(func() { img, err = p.captureBufferCurrent(c, buffer) })
	return img, err
}

func (p *painter) captureBufferCurrent(c granio.Canvas, buffer uint32) (image.Image, error) {
	backend, err := p.captureAPI()
	if err != nil {
		return nil, err
	}
	if backend != nil {
		return backend.Read(c, buffer)
	}
	pos := granio.NewPos(c.Size().Width, c.Size().Height)
	width, height := c.PixelCoordinateForPosition(pos)
	bytes, err := readback.RGBA8Size(width, height)
	if err != nil {
		return nil, err
	}
	pixels := make([]uint8, bytes)

	p.ctx.ReadBuffer(buffer)
	p.logError()
	p.ctx.ReadPixels(0, 0, width, height, colorFormatRGBA, unsignedByte, pixels)
	p.logError()
	// Mobile/WebGL adapters do not expose selectable read buffers.
	p.ctx.ReadBuffer(back)
	p.logError()

	return &captureImage{
		pix:           pixels,
		width:         width,
		height:        height,
		preserveAlpha: p.windowSurface != WindowSurfaceLegacy,
	}, nil
}
