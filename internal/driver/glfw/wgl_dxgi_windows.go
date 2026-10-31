//go:build windows && !wasm && !test_web_driver

package glfw

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"github.com/Uzer007-admin/glfw/v3.5/glfw"
	"golang.org/x/sys/windows"
)

// WGL остаётся владельцем контекста и заднего буфера. DXGI получает готовый
// кадр через WGL_NV_DX_interop2; чтение пикселей в память CPU не требуется.
// Все методы вызываются на потоке текущего WGL-контекста.
type wglDXGIPresenter struct {
	hwnd, context                         uintptr
	device, immediate, swap, back, shared *graphicsCOM
	interop, object                       uintptr
	fbo, renderbuffer                     uint32
	width, height                         int
	frames                                uint64
	d3d                                   windows.Handle
	api                                   wglDXGIApi
}

// Индексы методов соответствуют интерфейсам Windows SDK: IUnknown,
// ID3D11Device, ID3D11DeviceContext и IDXGISwapChain.
type graphicsCOM struct{ vtable *[115]uintptr }

//go:uintptrescapes
func (c *graphicsCOM) call(method int, args ...uintptr) uintptr {
	values := [12]uintptr{uintptr(unsafe.Pointer(c))}
	copy(values[1:], args)
	r, _, _ := syscall.SyscallN(c.vtable[method], values[:len(args)+1]...)
	runtime.KeepAlive(c)
	return r
}

// Фиксированные аргументы исключают выделение вариадического среза
// на горячем пути CopyResource/Present.
//
//go:uintptrescapes
func (c *graphicsCOM) call2(method int, a, b uintptr) uintptr {
	r, _, _ := syscall.SyscallN(c.vtable[method], uintptr(unsafe.Pointer(c)), a, b)
	runtime.KeepAlive(c)
	return r
}

func releaseGraphicsCOM(c **graphicsCOM) {
	if *c != nil {
		(*c).call(2)
		*c = nil
	}
}

func graphicsResult(operation string, result uintptr) error {
	if int32(result) < 0 {
		return fmt.Errorf("%s: HRESULT 0x%08x", operation, uint32(result))
	}
	return nil
}

type wglDXGIApi struct {
	open, close, register, unregister, lock, unlock     uintptr
	genFBO, deleteFBO, bindFBO, genRB, deleteRB, bindRB uintptr
	attach, check, blit                                 uintptr
	getInteger, isEnabled, enable, disable, readBuffer  uintptr
	currentContext                                      uintptr
}

func loadWGLDXGIApi() (wglDXGIApi, error) {
	a := wglDXGIApi{}
	entries := []struct {
		name string
		out  *uintptr
	}{
		{"wglDXOpenDeviceNV", &a.open}, {"wglDXCloseDeviceNV", &a.close},
		{"wglDXRegisterObjectNV", &a.register}, {"wglDXUnregisterObjectNV", &a.unregister},
		{"wglDXLockObjectsNV", &a.lock}, {"wglDXUnlockObjectsNV", &a.unlock},
		{"glGenFramebuffers", &a.genFBO}, {"glDeleteFramebuffers", &a.deleteFBO},
		{"glBindFramebuffer", &a.bindFBO}, {"glGenRenderbuffers", &a.genRB},
		{"glDeleteRenderbuffers", &a.deleteRB}, {"glBindRenderbuffer", &a.bindRB},
		{"glFramebufferRenderbuffer", &a.attach}, {"glCheckFramebufferStatus", &a.check},
		{"glBlitFramebuffer", &a.blit}, {"glGetIntegerv", &a.getInteger},
		{"glIsEnabled", &a.isEnabled}, {"glEnable", &a.enable},
		{"glDisable", &a.disable}, {"glReadBuffer", &a.readBuffer},
		{"wglGetCurrentContext", &a.currentContext},
	}
	if !glfw.ExtensionSupported("WGL_NV_DX_interop2") {
		return a, fmt.Errorf("WGL_NV_DX_interop2 is unavailable")
	}
	for _, entry := range entries {
		address := uintptr(glfw.GetProcAddress(entry.name))
		// Некоторые ICD возвращают эти значения вместо NULL при отказе.
		if address <= 3 || address == ^uintptr(0) {
			return a, fmt.Errorf("WGL/DXGI entry point %s is unavailable", entry.name)
		}
		*entry.out = address
	}
	return a, nil
}

type graphicsRational struct{ numerator, denominator uint32 }
type graphicsModeDesc struct {
	width, height             uint32
	refresh                   graphicsRational
	format, scanline, scaling uint32
}
type graphicsSwapDesc struct {
	mode                       graphicsModeDesc
	sampleCount, sampleQuality uint32
	usage, count               uint32
	window                     uintptr
	windowed                   int32
	effect, flags              uint32
}
type graphicsTextureDesc struct {
	width, height, mips, arraySize, format uint32
	sampleCount, sampleQuality             uint32
	usage, bind, cpuAccess, misc           uint32
}

var graphicsTexture2DIID = windows.GUID{Data1: 0x6f15aaf2, Data2: 0xd208, Data3: 0x4e89, Data4: [8]byte{0x9a, 0xb4, 0x48, 0x95, 0x35, 0xd3, 0x4f, 0x9c}}

func newWGLDXGIPresenter(hwnd uintptr) (_ *wglDXGIPresenter, err error) {
	current := glfw.GetCurrentContext()
	if current == nil || current.GetAttrib(glfw.ClientAPI) != glfw.OpenGLAPI {
		return nil, fmt.Errorf("WGL/DXGI requires a desktop OpenGL context")
	}
	a, err := loadWGLDXGIApi()
	if err != nil {
		return nil, err
	}
	context, _, _ := syscall.SyscallN(a.currentContext)
	if context == 0 {
		return nil, fmt.Errorf("WGL/DXGI requires a current WGL context")
	}
	p := &wglDXGIPresenter{hwnd: hwnd, context: context, api: a}
	defer func() {
		if err != nil {
			p.destroy()
		}
	}()
	p.d3d, err = windows.LoadLibraryEx("d3d11.dll", 0, windows.LOAD_LIBRARY_SEARCH_SYSTEM32)
	if err != nil {
		return nil, err
	}
	create, err := windows.GetProcAddress(p.d3d, "D3D11CreateDeviceAndSwapChain")
	if err != nil {
		return nil, err
	}
	// Ненулевой начальный буфер позволяет создать и скрытое окно. Реальные
	// размеры выделяются по первому ненулевому framebuffer непосредственно перед выводом.
	desc := graphicsSwapDesc{mode: graphicsModeDesc{width: 1, height: 1, format: 28}, sampleCount: 1, usage: 0x20, count: 2, window: hwnd, windowed: 1, effect: 3}
	var level uint32
	result, _, _ := syscall.SyscallN(create, 0, 1, 0, 0, 0, 0, 7, uintptr(unsafe.Pointer(&desc)),
		uintptr(unsafe.Pointer(&p.swap)), uintptr(unsafe.Pointer(&p.device)),
		uintptr(unsafe.Pointer(&level)), uintptr(unsafe.Pointer(&p.immediate)))
	if err = graphicsResult("D3D11CreateDeviceAndSwapChain", result); err != nil {
		return nil, err
	}
	p.interop, _, _ = syscall.SyscallN(a.open, uintptr(unsafe.Pointer(p.device)))
	if p.interop == 0 {
		return nil, fmt.Errorf("wglDXOpenDeviceNV failed")
	}
	return p, nil
}

func (p *wglDXGIPresenter) current() bool {
	context, _, _ := syscall.SyscallN(p.api.currentContext)
	return context != 0 && context == p.context
}

func (p *wglDXGIPresenter) integer(name uintptr) int32 {
	var value int32
	syscall.SyscallN(p.api.getInteger, name, uintptr(unsafe.Pointer(&value)))
	return value
}

func (p *wglDXGIPresenter) freeBuffers() {
	if p.object != 0 {
		syscall.SyscallN(p.api.unregister, p.interop, p.object)
		p.object = 0
	}
	if p.fbo != 0 {
		syscall.SyscallN(p.api.deleteFBO, 1, uintptr(unsafe.Pointer(&p.fbo)))
		p.fbo = 0
	}
	if p.renderbuffer != 0 {
		syscall.SyscallN(p.api.deleteRB, 1, uintptr(unsafe.Pointer(&p.renderbuffer)))
		p.renderbuffer = 0
	}
	releaseGraphicsCOM(&p.shared)
	releaseGraphicsCOM(&p.back)
	p.width, p.height = 0, 0
}

func (p *wglDXGIPresenter) resize(width, height int) (err error) {
	p.freeBuffers()
	p.immediate.call(110) // ClearState: у ResizeBuffers не должно остаться ссылок на старые буферы.
	p.immediate.call(111) // Flush
	result := p.swap.call(13, 2, uintptr(width), uintptr(height), 28, 0)
	if err = graphicsResult("DXGI ResizeBuffers", result); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			p.freeBuffers()
		}
	}()
	result = p.swap.call(9, 0, uintptr(unsafe.Pointer(&graphicsTexture2DIID)), uintptr(unsafe.Pointer(&p.back)))
	if err = graphicsResult("DXGI GetBuffer", result); err != nil {
		return err
	}
	texture := graphicsTextureDesc{width: uint32(width), height: uint32(height), mips: 1, arraySize: 1, format: 28, sampleCount: 1, bind: 0x28}
	result = p.device.call(5, uintptr(unsafe.Pointer(&texture)), 0, uintptr(unsafe.Pointer(&p.shared)))
	if err = graphicsResult("D3D11 CreateTexture2D", result); err != nil {
		return err
	}
	syscall.SyscallN(p.api.genRB, 1, uintptr(unsafe.Pointer(&p.renderbuffer)))
	syscall.SyscallN(p.api.genFBO, 1, uintptr(unsafe.Pointer(&p.fbo)))
	p.object, _, _ = syscall.SyscallN(p.api.register, p.interop, uintptr(unsafe.Pointer(p.shared)), uintptr(p.renderbuffer), 0x8d41, 2)
	if p.object == 0 {
		return fmt.Errorf("wglDXRegisterObjectNV failed")
	}
	locked, _, _ := syscall.SyscallN(p.api.lock, p.interop, 1, uintptr(unsafe.Pointer(&p.object)))
	if locked == 0 {
		return fmt.Errorf("wglDXLockObjectsNV failed")
	}
	read, draw := p.integer(0x8caa), p.integer(0x8ca6)
	syscall.SyscallN(p.api.bindFBO, 0x8d40, uintptr(p.fbo))
	syscall.SyscallN(p.api.attach, 0x8d40, 0x8ce0, 0x8d41, uintptr(p.renderbuffer))
	status, _, _ := syscall.SyscallN(p.api.check, 0x8d40)
	syscall.SyscallN(p.api.bindFBO, 0x8ca8, uintptr(read))
	syscall.SyscallN(p.api.bindFBO, 0x8ca9, uintptr(draw))
	unlocked, _, _ := syscall.SyscallN(p.api.unlock, p.interop, 1, uintptr(unsafe.Pointer(&p.object)))
	if unlocked == 0 {
		return fmt.Errorf("wglDXUnlockObjectsNV failed")
	}
	if status != 0x8cd5 {
		return fmt.Errorf("WGL/DXGI framebuffer incomplete: 0x%x", status)
	}
	p.width, p.height = width, height
	return nil
}

func (p *wglDXGIPresenter) present(width, height int) error {
	// Нулевые/minimized кадры не выделяют ресурсы и не заменяют последний кадр.
	if width <= 0 || height <= 0 {
		return nil
	}
	if !p.current() {
		return fmt.Errorf("WGL/DXGI context is not current")
	}
	if p.width != width || p.height != height {
		if err := p.resize(width, height); err != nil {
			return err
		}
	}
	locked, _, _ := syscall.SyscallN(p.api.lock, p.interop, 1, uintptr(unsafe.Pointer(&p.object)))
	if locked == 0 {
		return fmt.Errorf("wglDXLockObjectsNV failed")
	}
	// Blit учитывает scissor. Снимаем его только на копирование полного кадра,
	// затем возвращаем read/draw FBO, read buffer и состояние scissor вызывающему.
	read, draw := p.integer(0x8caa), p.integer(0x8ca6)
	scissor, _, _ := syscall.SyscallN(p.api.isEnabled, 0x0c11)
	syscall.SyscallN(p.api.disable, 0x0c11)
	syscall.SyscallN(p.api.bindFBO, 0x8ca8, 0)
	buffer := p.integer(0x0c02)
	syscall.SyscallN(p.api.readBuffer, 0x0405)
	syscall.SyscallN(p.api.bindFBO, 0x8ca9, uintptr(p.fbo))
	syscall.SyscallN(p.api.blit, 0, 0, uintptr(width), uintptr(height), 0, uintptr(height), uintptr(width), 0, 0x4000, 0x2600)
	syscall.SyscallN(p.api.readBuffer, uintptr(buffer))
	syscall.SyscallN(p.api.bindFBO, 0x8ca8, uintptr(read))
	syscall.SyscallN(p.api.bindFBO, 0x8ca9, uintptr(draw))
	if scissor != 0 {
		syscall.SyscallN(p.api.enable, 0x0c11)
	}
	unlocked, _, _ := syscall.SyscallN(p.api.unlock, p.interop, 1, uintptr(unsafe.Pointer(&p.object)))
	if unlocked == 0 {
		return fmt.Errorf("wglDXUnlockObjectsNV failed")
	}
	p.immediate.call2(47, uintptr(unsafe.Pointer(p.back)), uintptr(unsafe.Pointer(p.shared))) // CopyResource
	result := p.swap.call2(8, 0, 0)                                                           // Частоту ограничивает диспетчер кадров Granio.
	if err := graphicsResult("DXGI Present", result); err != nil {
		return err
	}
	p.frames++
	return nil
}

func (p *wglDXGIPresenter) destroy() {
	p.freeBuffers()
	if p.interop != 0 {
		syscall.SyscallN(p.api.close, p.interop)
		p.interop = 0
	}
	if p.immediate != nil {
		p.immediate.call(110)
		p.immediate.call(111)
	}
	releaseGraphicsCOM(&p.swap)
	releaseGraphicsCOM(&p.immediate)
	releaseGraphicsCOM(&p.device)
	if p.d3d != 0 {
		windows.FreeLibrary(p.d3d)
		p.d3d = 0
	}
}
