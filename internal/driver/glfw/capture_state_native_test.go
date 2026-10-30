//go:build !no_glfw && !ci && !mobile && !wasm && !test_web_driver

package glfw

import (
	"errors"
	"image"
	"image/color"
	"testing"
	"unsafe"

	"github.com/Uzer007-admin/gl/capabilities"
	capnative "github.com/Uzer007-admin/gl/capabilities/native"
	"github.com/Uzer007-admin/gl/material"
	"github.com/Uzer007-admin/gl/readback"
	es "github.com/Uzer007-admin/gl/v2.0/gles2"
	core "github.com/Uzer007-admin/gl/v2.1/material"
	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	paintergl "github.com/Uzer007-admin/granio/v2/internal/painter/gl"
	"github.com/stretchr/testify/require"
)

type captureTestAPI interface {
	material.Procedures
	readback.Procedures
	GenBuffers(int32, *uint32)
	DeleteBuffers(int32, *uint32)
	BufferData(uint32, int, unsafe.Pointer, uint32)
	GetFramebufferViewport(bool) (int32, int32, [4]int32)
}

func loadCaptureTestAPI() (capabilities.Context, captureTestAPI, error) {
	c, err := capnative.Query(glfw.GetProcAddress)
	if err != nil {
		return c, nil, err
	}
	if c.API == capabilities.OpenGLES {
		api, err := es.LoadWithProcAddrFunc(glfw.GetProcAddress)
		return c, api, err
	}
	api, err := core.LoadWithProcAddrFunc(glfw.GetProcAddress)
	return c, api, err
}

func TestNativeCaptureRestoresPackAndFramebufferState(t *testing.T) {
	w := createWindowWithResizeCallback("Capture GL state")
	defer w.Close()
	w.SetPadded(false)
	pixels := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			pixels.SetNRGBA(x, y, color.NRGBA{G: 255, A: 255})
		}
	}
	picture := canvas.NewImageFromImage(pixels)
	w.SetContent(picture)
	w.Resize(granio.NewSize(83, 61))
	runOnMain(func() { w.window.Show() })
	ensureCanvasSize(t, w, granio.NewSize(83, 61))
	var actual image.Image
	var failure error
	var before, after [16]int32
	var glError uint32
	runOnMain(func() {
		w.window.RunWithContext(func() {
			ctx, api, err := loadCaptureTestAPI()
			if err != nil {
				failure = err
				return
			}
			t.Logf("GL version=%s; vendor=%s; renderer=%s", ctx.Version, ctx.Vendor, ctx.Renderer)
			device, err := material.NewGLDevice(ctx, api)
			if err != nil {
				failure = err
				return
			}
			handle, err := device.CreateTarget(17, 11)
			if err != nil {
				failure = err
				return
			}
			defer device.DeleteTarget(handle)
			rows := ctx.API == capabilities.OpenGL || ctx.AtLeast(3, 0) || ctx.HasExtension("GL_NV_pack_subimage")
			unpackRows := ctx.API == capabilities.OpenGL || ctx.AtLeast(3, 0) || ctx.HasExtension("GL_EXT_unpack_subimage")
			pbo := ctx.API == capabilities.OpenGL || ctx.AtLeast(3, 0) || ctx.HasExtension("GL_NV_pixel_buffer_object")
			split := ctx.API == capabilities.OpenGL && (ctx.AtLeast(3, 0) || ctx.HasExtension("GL_ARB_framebuffer_object") || ctx.HasExtension("GL_EXT_framebuffer_blit")) || ctx.API == capabilities.OpenGLES && ctx.AtLeast(3, 0)
			var buffer uint32
			if pbo {
				api.GenBuffers(1, &buffer)
				api.BindBuffer(0x88EB, buffer)
				api.BufferData(0x88EB, 16, nil, 0x88E4)
				api.BindBuffer(0x88EC, buffer)
				defer func() { api.BindBuffer(0x88EB, 0); api.BindBuffer(0x88EC, 0); api.DeleteBuffers(1, &buffer) }()
			}
			api.PixelStorei(0x0D05, 8)
			defer api.PixelStorei(0x0D05, 4)
			api.PixelStorei(0x0CF5, 8)
			defer api.PixelStorei(0x0CF5, 4)
			if rows {
				api.PixelStorei(0x0D02, 97)
				api.PixelStorei(0x0D03, 3)
				api.PixelStorei(0x0D04, 5)
				defer func() { api.PixelStorei(0x0D02, 0); api.PixelStorei(0x0D03, 0); api.PixelStorei(0x0D04, 0) }()
			}
			if unpackRows {
				api.PixelStorei(0x0CF2, 97)
				api.PixelStorei(0x0CF3, 3)
				api.PixelStorei(0x0CF4, 5)
				defer func() { api.PixelStorei(0x0CF2, 0); api.PixelStorei(0x0CF3, 0); api.PixelStorei(0x0CF4, 0) }()
			}
			snapshot := func() (state [16]int32) {
				draw, read, viewport := api.GetFramebufferViewport(split)
				state[0], state[1] = draw, read
				copy(state[2:6], viewport[:])
				api.GetIntegerv(0x0D05, &state[6])
				if rows {
					api.GetIntegerv(0x0D02, &state[7])
					api.GetIntegerv(0x0D03, &state[8])
					api.GetIntegerv(0x0D04, &state[9])
				}
				if pbo {
					api.GetIntegerv(0x88ED, &state[10])
					api.GetIntegerv(0x88EF, &state[11])
				}
				api.GetIntegerv(0x0CF5, &state[12])
				if unpackRows {
					api.GetIntegerv(0x0CF2, &state[13])
					api.GetIntegerv(0x0CF3, &state[14])
					api.GetIntegerv(0x0CF4, &state[15])
				}
				return
			}
			failure = device.RenderTarget(material.Target{Handle: handle, Width: 17, Height: 11}, func() error {
				api.Viewport(4, 5, 13, 7)
				before = snapshot()
				fresh := canvas.NewImageFromImage(pixels)
				w.canvas.SetContent(fresh) // no cached texture exists for this object
				actual, err = w.canvas.CaptureWithError()
				after = snapshot()
				return err
			})
			glError = api.GetError()
		})
	})
	require.NoError(t, failure)
	require.Zero(t, glError)
	require.Equal(t, before, after, "capture must restore the host FBO, viewport, pack layout and PBO binding")
	require.NotNil(t, actual)
	require.Equal(t, image.Rect(0, 0, 83, 61), actual.Bounds())
	r, g, b, _ := actual.At(41, 30).RGBA()
	require.Less(t, r, uint32(0xff))
	require.Greater(t, g, uint32(0xff00))
	require.Less(t, b, uint32(0xff))
}

func TestNativeCaptureHiddenFirstFrameAndWorker(t *testing.T) {
	w := createWindowWithResizeCallback("Capture first hidden frame")
	defer w.Close()
	w.SetPadded(false)
	w.SetContent(canvas.NewRectangle(color.NRGBA{R: 255, A: 255}))
	w.Resize(granio.NewSize(80, 60))
	var img image.Image
	var err error
	// Call the real canvas from the test worker. No Show, SwapBuffers or
	// granio.Do wrapper precedes capture, including with migrated_fynedo.
	img, err = w.glCanvasCaptureForTest()
	require.NoError(t, err)
	require.Equal(t, image.Rect(0, 0, 80, 60), img.Bounds())
	r, g, b, _ := img.At(40, 30).RGBA()
	require.Greater(t, r, uint32(0xff00))
	require.Less(t, g, uint32(0xff))
	require.Less(t, b, uint32(0xff))
}

func (w *safeWindow) glCanvasCaptureForTest() (image.Image, error) {
	return w.window.canvas.CaptureWithError()
}

func TestNativeCaptureAlphaAndClosedSurface(t *testing.T) {
	w := createWindowWithResizeCallback("Capture alpha")
	defer w.Close()
	w.SetPadded(false)
	w.SetContent(canvas.NewRectangle(color.NRGBA{R: 255, A: 128}))
	w.Resize(granio.NewSize(80, 60))
	var supported bool
	runOnMain(func() {
		w.window.RunWithContext(func() {
			supported = w.canvas.Painter().(interface {
				SetWindowSurface(paintergl.WindowSurfaceMode) bool
			}).SetWindowSurface(paintergl.WindowSurfaceBackdrop)
		})
	})
	if !supported {
		w.Close()
		t.Skip("alpha mode unavailable")
	}
	img, err := w.glCanvasCaptureForTest()
	require.NoError(t, err)
	r, g, b, a := img.At(40, 30).RGBA()
	require.InDelta(t, 128, float64(a)/257, 1)
	require.InDelta(t, 128, float64(r)/257, 1)
	require.Zero(t, g)
	require.Zero(t, b)
	w.Close()
	img, err = w.glCanvasCaptureForTest()
	require.True(t, errors.Is(err, desktop.ErrCaptureUnavailable))
	require.Nil(t, img)
}
