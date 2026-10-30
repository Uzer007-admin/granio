//go:build !android && !ios && !mobile && !wasm && !test_web_driver

package gl

import (
	"errors"
	"fmt"
	"image"

	"github.com/Uzer007-admin/gl/capabilities"
	capnative "github.com/Uzer007-admin/gl/capabilities/native"
	"github.com/Uzer007-admin/gl/material"
	"github.com/Uzer007-admin/gl/readback"
	es "github.com/Uzer007-admin/gl/v2.0/gles2"
	core "github.com/Uzer007-admin/gl/v2.1/material"
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
)

type captureProcedures interface {
	material.Procedures
	readback.Procedures
}
type nativeCapture struct {
	api                      captureProcedures
	reader                   *readback.Reader
	device                   *material.GLDevice
	maximum                  int
	desktop                  bool
	unpackRows, unpackBuffer bool
	painter                  *painter
}

func newCaptureBackend(p *painter) (captureBackend, error) {
	context, err := capnative.Query(glfw.GetProcAddress)
	if err != nil {
		return nil, err
	}
	var api captureProcedures
	if context.API == capabilities.OpenGLES {
		api, err = es.LoadWithProcAddrFunc(glfw.GetProcAddress)
	} else {
		api, err = core.LoadWithProcAddrFunc(glfw.GetProcAddress)
	}
	if err != nil {
		return nil, err
	}
	reader, err := readback.New(context, api)
	if err != nil {
		return nil, err
	}
	device, err := material.NewGLDevice(context, api)
	if err != nil && !errors.Is(err, material.ErrUnsupported) {
		return nil, err
	}
	return &nativeCapture{api: api, reader: reader, device: device, maximum: context.MaxTextureSize,
		desktop:      context.API == capabilities.OpenGL,
		unpackRows:   context.API == capabilities.OpenGL || context.AtLeast(3, 0) || context.HasExtension("GL_EXT_unpack_subimage"),
		unpackBuffer: context.API == capabilities.OpenGL && (context.AtLeast(2, 1) || context.HasExtension("GL_ARB_pixel_buffer_object") || context.HasExtension("GL_EXT_pixel_buffer_object")) || context.API == capabilities.OpenGLES && (context.AtLeast(3, 0) || context.HasExtension("GL_NV_pixel_buffer_object")),
		painter:      p}, nil
}

// The draw callback may upload freshly generated CPU textures. The host's
// unpack layout/PBO must not reinterpret their Go pointers or row strides.
func (n *nativeCapture) draw(draw func()) {
	var saved [5]int32
	n.api.GetIntegerv(0x0CF5, &saved[0]) // UNPACK_ALIGNMENT
	if n.unpackRows {
		n.api.GetIntegerv(0x0CF2, &saved[1]) // UNPACK_ROW_LENGTH
		n.api.GetIntegerv(0x0CF3, &saved[2]) // UNPACK_SKIP_ROWS
		n.api.GetIntegerv(0x0CF4, &saved[3]) // UNPACK_SKIP_PIXELS
	}
	if n.unpackBuffer {
		n.api.GetIntegerv(0x88EF, &saved[4])
	}
	defer func() {
		n.api.PixelStorei(0x0CF5, saved[0])
		if n.unpackRows {
			n.api.PixelStorei(0x0CF2, saved[1])
			n.api.PixelStorei(0x0CF3, saved[2])
			n.api.PixelStorei(0x0CF4, saved[3])
		}
		if n.unpackBuffer {
			n.api.BindBuffer(0x88EC, uint32(saved[4]))
		}
	}()
	n.api.PixelStorei(0x0CF5, 1)
	if n.unpackRows {
		n.api.PixelStorei(0x0CF2, 0)
		n.api.PixelStorei(0x0CF3, 0)
		n.api.PixelStorei(0x0CF4, 0)
	}
	if n.unpackBuffer {
		n.api.BindBuffer(0x88EC, 0)
	}
	draw()
}

func captureSize(c granio.Canvas) (width, height, bytes int, err error) {
	size := c.Size()
	width, height = c.PixelCoordinateForPosition(granio.NewPos(size.Width, size.Height))
	bytes, err = readback.RGBA8Size(width, height)
	return
}

func (n *nativeCapture) Capture(c granio.Canvas, draw func()) (image.Image, error) {
	width, height, bytes, err := captureSize(c)
	if err != nil {
		return nil, err
	}
	if width > n.maximum || height > n.maximum {
		return nil, readback.ErrDimensions
	}
	if n.device == nil {
		// GL 2.1 implementations without texture FBOs can still capture a live
		// drawable. A suspended native framebuffer is explicitly unavailable.
		current := glfw.GetCurrentContext()
		if current == nil {
			return nil, capabilities.ErrNoContext
		}
		fw, fh := current.GetFramebufferSize()
		if fw < width || fh < height {
			return nil, fmt.Errorf("GL capture: drawable %dx%d is smaller than %dx%d", fw, fh, width, height)
		}
		var viewport [4]int32
		n.api.GetIntegerv(0x0BA2, &viewport[0])
		defer n.api.Viewport(viewport[0], viewport[1], viewport[2], viewport[3])
		n.draw(draw)
		return n.Read(c, back)
	}
	handle, err := n.device.CreateTarget(width, height)
	if err != nil {
		return nil, err
	}
	defer n.device.DeleteTarget(handle)
	pixels := make([]byte, bytes)
	err = n.device.RenderTarget(material.Target{Handle: handle, Width: width, Height: height}, func() error {
		n.draw(draw)
		return n.reader.ReadRGBA8(0, 0, width, height, pixels)
	})
	if err != nil {
		return nil, err
	}
	return &captureImage{pix: pixels, width: width, height: height, preserveAlpha: n.painter.windowSurface != WindowSurfaceLegacy}, nil
}

func (n *nativeCapture) Read(c granio.Canvas, source uint32) (image.Image, error) {
	width, height, bytes, err := captureSize(c)
	if err != nil {
		return nil, err
	}
	if width > n.maximum || height > n.maximum {
		return nil, readback.ErrDimensions
	}
	if n.desktop {
		api := n.api.(interface{ ReadBuffer(uint32) })
		var previous int32
		n.api.GetIntegerv(0x0C02, &previous)
		api.ReadBuffer(source)
		defer api.ReadBuffer(uint32(previous))
	}
	pixels := make([]byte, bytes)
	if err = n.reader.ReadRGBA8(0, 0, width, height, pixels); err != nil {
		return nil, err
	}
	return &captureImage{pix: pixels, width: width, height: height, preserveAlpha: n.painter.windowSurface != WindowSurfaceLegacy}, nil
}
