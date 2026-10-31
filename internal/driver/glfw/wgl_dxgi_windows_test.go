//go:build windows && !no_glfw && !ci && !mobile && !wasm && !test_web_driver

package glfw

import (
	"image/color"
	"os"
	"syscall"
	"testing"
	"unsafe"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/stretchr/testify/require"
)

// Читается один пиксель собственной interop-текстуры. Этот путь отсутствует
// в рабочем выводе кадров и не захватывает рабочий стол или чужие окна.
func sampleDXGIPixel(p *wglDXGIPresenter, x, y uint32) (color.NRGBA, error) {
	desc := graphicsTextureDesc{width: 1, height: 1, mips: 1, arraySize: 1, format: 28, sampleCount: 1, usage: 3, cpuAccess: 0x20000}
	var staging *graphicsCOM
	r := p.device.call(5, uintptr(unsafe.Pointer(&desc)), 0, uintptr(unsafe.Pointer(&staging)))
	if err := graphicsResult("test staging texture", r); err != nil {
		return color.NRGBA{}, err
	}
	defer releaseGraphicsCOM(&staging)
	box := [6]uint32{x, y, 0, x + 1, y + 1, 1}
	p.immediate.call(46, uintptr(unsafe.Pointer(staging)), 0, 0, 0, 0, uintptr(unsafe.Pointer(p.shared)), 0, uintptr(unsafe.Pointer(&box)))
	var mapped struct {
		data                 unsafe.Pointer
		rowPitch, depthPitch uint32
	}
	r = p.immediate.call(14, uintptr(unsafe.Pointer(staging)), 0, 1, 0, uintptr(unsafe.Pointer(&mapped)))
	if err := graphicsResult("test Map", r); err != nil {
		return color.NRGBA{}, err
	}
	defer p.immediate.call(15, uintptr(unsafe.Pointer(staging)), 0)
	pixel := unsafe.Slice((*byte)(mapped.data), 4)
	return color.NRGBA{R: pixel[0], G: pixel[1], B: pixel[2], A: pixel[3]}, nil
}

func TestWGLDXGINativeFrameResizeStateAndFallback(t *testing.T) {
	if os.Getenv("GRANIO_WGL_PRESENT") != "dxgi" {
		t.Skip("run with GRANIO_WGL_PRESENT=dxgi")
	}
	var raw *window
	var configErr error
	runOnMain(func() {
		raw = d.CreateWindow("Проверка аппаратного OpenGL/DXGI").(*window)
		_, configErr = raw.ConfigureVisuals(desktop.VisualOptions{Backdrop: desktop.BackdropMica, Fallback: desktop.VisualFallbackOpaque})
		raw.SetPadded(false)
		raw.SetContent(container.NewGridWithRows(2, canvas.NewRectangle(color.NRGBA{R: 255, A: 255}), canvas.NewRectangle(color.NRGBA{B: 255, A: 255})))
		raw.Resize(granio.NewSize(120, 80))
		raw.Show()
	})
	w := &safeWindow{window: raw}
	defer w.Close()
	require.NoError(t, configErr)
	ensureCanvasSize(t, w, granio.NewSize(120, 80))
	var p *wglDXGIPresenter
	runOnMain(func() { p = raw.presentation.dxgi })
	if p == nil {
		if os.Getenv("GRANIO_REQUIRE_DXGI") == "1" {
			t.Fatal("requested hardware WGL/DXGI presenter was not created")
		}
		t.Skip("WGL/DXGI interop is unavailable on this GPU")
	}
	for _, size := range []granio.Size{granio.NewSize(120, 80), granio.NewSize(171, 113), granio.NewSize(120, 80)} {
		w.Resize(size)
		ensureCanvasSize(t, w, size)
		var top, bottom color.NRGBA
		var firstErr, secondErr, captureErr, zeroErr error
		var imageTop color.Color
		var frames, afterZero uint64
		var width, height int
		runOnMain(func() {
			raw.RunWithContext(func() {
				d.repaintWindow(raw)
				width, height = raw.viewport.GetFramebufferSize()
				frames = p.frames
				top, firstErr = sampleDXGIPixel(p, 10, 10)
				bottom, secondErr = sampleDXGIPixel(p, 10, uint32(height-10))
				img, err := raw.canvas.CaptureWithError()
				captureErr = err
				if err == nil {
					imageTop = img.At(10, 10)
				}
				zeroErr = p.present(0, height)
				afterZero = p.frames
			})
		})
		require.NoError(t, firstErr)
		require.NoError(t, secondErr)
		require.NoError(t, captureErr)
		require.NoError(t, zeroErr)
		require.Positive(t, width)
		require.Positive(t, height)
		require.Positive(t, frames)
		require.Equal(t, frames, afterZero, "a zero frame must not be presented")
		require.Equal(t, color.NRGBA{R: 255, A: 255}, top, "top row and RGBA order")
		require.Equal(t, color.NRGBA{B: 255, A: 255}, bottom, "bottom row must not be vertically mirrored")
		r, g, b, a := imageTop.RGBA()
		require.Equal(t, uint32(65535), r)
		require.Zero(t, g)
		require.Zero(t, b)
		require.Equal(t, uint32(65535), a)
	}
	var stateErr error
	var readBuffer int32
	var scissor uintptr
	var box [4]int32
	var copiedPixel color.NRGBA
	var copiedErr error
	runOnMain(func() {
		raw.RunWithContext(func() {
			// Зелёная метка в текстуре назначения должна быть перезаписана
			// полным красным кадром даже за пределами входного scissor.
			marker := [4]byte{0, 255, 0, 255}
			region := [6]uint32{10, 10, 0, 11, 11, 1}
			p.immediate.call(48, uintptr(unsafe.Pointer(p.shared)), 0, uintptr(unsafe.Pointer(&region)), uintptr(unsafe.Pointer(&marker[0])), 4, 0)
			// Входной scissor не должен обрезать копирование кадра, а GL_FRONT
			// не должен заставить мост скопировать неотрисованный передний буфер.
			syscall.SyscallN(uintptr(glfw.GetProcAddress("glScissor")), 4, 5, 6, 7)
			syscall.SyscallN(p.api.enable, 0x0c11)
			syscall.SyscallN(p.api.readBuffer, 0x0404)
			stateErr = p.present(p.width, p.height)
			copiedPixel, copiedErr = sampleDXGIPixel(p, 10, 10)
			readBuffer = p.integer(0x0c02)
			scissor, _, _ = syscall.SyscallN(p.api.isEnabled, 0x0c11)
			syscall.SyscallN(p.api.getInteger, 0x0c10, uintptr(unsafe.Pointer(&box)))
			syscall.SyscallN(p.api.disable, 0x0c11)
			syscall.SyscallN(p.api.readBuffer, 0x0405)
		})
	})
	require.NoError(t, stateErr)
	require.NoError(t, copiedErr)
	require.Equal(t, color.NRGBA{R: 255, A: 255}, copiedPixel, "the whole frame must replace the destination outside the incoming scissor")
	require.Equal(t, int32(0x0404), readBuffer)
	require.Equal(t, uintptr(1), scissor)
	require.Equal(t, [4]int32{4, 5, 6, 7}, box)
	status := raw.VisualStatus()
	require.Equal(t, desktop.VisualFallback, status.Phase)
	require.Equal(t, desktop.BackdropOpaque, status.Selected)
	require.Equal(t, desktop.VisualUnsupportedBackend, status.Reason)
	// Подменяется только ответ Present нашей копии presenter. Настоящий GPU
	// не сбрасывается: проверяем ограниченное восстановление после
	// DXGI_ERROR_DEVICE_REMOVED и освобождение старых interop-ресурсов.
	removed := syscall.NewCallback(func(_, _, _ uintptr) uintptr { return 0x887a0005 })
	var replacement *wglDXGIPresenter
	var recoveredPixel color.NRGBA
	var recoveredErr error
	runOnMain(func() {
		raw.RunWithContext(func() {
			p.presentMethod = removed
			raw.presentFrame()
			replacement = raw.presentation.dxgi
			if replacement != nil {
				d.repaintWindow(raw)
				recoveredPixel, recoveredErr = sampleDXGIPixel(replacement, 10, 10)
			}
		})
	})
	require.NotNil(t, replacement)
	require.NotSame(t, p, replacement)
	require.Zero(t, p.interop)
	require.Nil(t, p.device)
	require.NoError(t, recoveredErr)
	require.Equal(t, color.NRGBA{R: 255, A: 255}, recoveredPixel)
	p = replacement
	// Освобождение выполняется до удаления GLFW/WGL-контекста; повторный вызов безопасен.
	runOnMain(func() { raw.releasePresentation(); raw.releasePresentation() })
	require.Zero(t, p.interop)
	require.Zero(t, p.d3d)
	require.Nil(t, p.device)
	require.Nil(t, p.swap)
}

func TestWGLDXGIAutomaticAndAlphaSelection(t *testing.T) {
	for _, scenario := range []struct {
		name, mode string
		alpha      bool
	}{
		{"default", "", false}, {"legacy", "wgl", false}, {"prepared-alpha", "", true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			t.Setenv("GRANIO_WGL_PRESENT", scenario.mode)
			var raw *window
			var configErr error
			runOnMain(func() {
				raw = d.CreateWindow("Проверка выбора аппаратного вывода").(*window)
				if scenario.alpha {
					_, configErr = raw.ConfigureVisuals(desktop.VisualOptions{PrepareAlpha: true, Backdrop: desktop.BackdropOpaque, Fallback: desktop.VisualFallbackOpaque})
				}
				raw.SetPadded(false)
				raw.SetContent(canvas.NewRectangle(color.NRGBA{G: 255, A: 255}))
				raw.Resize(granio.NewSize(100, 80))
				raw.Show()
			})
			w := &safeWindow{window: raw}
			defer w.Close()
			require.NoError(t, configErr)
			var p *wglDXGIPresenter
			var supports bool
			runOnMain(func() {
				raw.RunWithContext(func() {
					p = raw.presentation.dxgi
					supports = raw.viewport.GetAttrib(glfw.ClientAPI) == glfw.OpenGLAPI && glfw.ExtensionSupported("WGL_NV_DX_interop2")
				})
			})
			if scenario.alpha || scenario.mode == "wgl" {
				require.Nil(t, p)
				return
			}
			if !supports {
				t.Skip("interop is unavailable for this context")
			}
			require.NotNil(t, p, "a compatible opaque window must select DXGI by default")
			var maxWidth, maxHeight int
			runOnMain(func() { maxWidth, maxHeight = raw.screenSize(granio.NewSize(32768, 32768)) })
			require.LessOrEqual(t, maxWidth, p.maxDimension)
			require.LessOrEqual(t, maxHeight, p.maxDimension)
			var lateErr error
			runOnMain(func() {
				_, lateErr = raw.ConfigureVisuals(desktop.VisualOptions{PrepareAlpha: true, Backdrop: desktop.BackdropTransparent})
			})
			require.ErrorIs(t, lateErr, desktop.ErrVisualNeedsRecreate)
		})
	}
}
