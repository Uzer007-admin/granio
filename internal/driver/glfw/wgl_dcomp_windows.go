//go:build windows && !wasm && !test_web_driver

package glfw

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// DirectComposition принимает premultiplied RGBA непосредственно из GPU.
// Дерево присоединяется только после первого полноразмерного Present:
// начальная служебная текстура 1x1 никогда не становится содержимым окна.
type wglComposition struct {
	library                windows.Handle
	device, target, visual *graphicsCOM
	attached               bool
}

type graphicsCompositionDesc struct {
	width, height, format, stereo uint32
	sampleCount, sampleQuality    uint32
	usage, count, scaling, effect uint32
	alpha, flags                  uint32
}

var (
	graphicsDXGIDeviceIID = windows.GUID{Data1: 0x54ec77fa, Data2: 0x1377, Data3: 0x44e6, Data4: [8]byte{0x8c, 0x32, 0x88, 0xfd, 0x5f, 0x44, 0xc8, 0x4c}}
	graphicsFactory2IID   = windows.GUID{Data1: 0x50c83a1c, Data2: 0xe072, Data3: 0x4c48, Data4: [8]byte{0x87, 0xb0, 0x36, 0x30, 0xfa, 0x36, 0xa6, 0xd0}}
	graphicsDCompIID      = windows.GUID{Data1: 0xc37ea93a, Data2: 0xe7aa, Data3: 0x450d, Data4: [8]byte{0xb1, 0x6f, 0x97, 0x46, 0xcb, 0x04, 0x07, 0xf3}}
)

func (p *wglDXGIPresenter) initComposition() error {
	style, _, _ := windows.NewLazySystemDLL("user32.dll").NewProc("GetWindowLongW").Call(p.hwnd, ^uintptr(19))
	if style&0x00200000 == 0 {
		return fmt.Errorf("DirectComposition requires a window created without a redirection bitmap")
	}
	c := &wglComposition{}
	p.composition = c // destroy освобождает и частично созданное дерево.
	var dxgi, adapter, factory *graphicsCOM
	defer releaseGraphicsCOM(&dxgi)
	defer releaseGraphicsCOM(&adapter)
	defer releaseGraphicsCOM(&factory)
	if err := graphicsResult("QueryInterface IDXGIDevice", p.device.call(0, uintptr(unsafe.Pointer(&graphicsDXGIDeviceIID)), uintptr(unsafe.Pointer(&dxgi)))); err != nil {
		return err
	}
	if err := graphicsResult("DXGI GetAdapter", dxgi.call(7, uintptr(unsafe.Pointer(&adapter)))); err != nil {
		return err
	}
	if err := graphicsResult("DXGI GetParent IDXGIFactory2", adapter.call(6, uintptr(unsafe.Pointer(&graphicsFactory2IID)), uintptr(unsafe.Pointer(&factory)))); err != nil {
		return err
	}
	// FLIP_SEQUENTIAL, STRETCH и PREMULTIPLIED обязательны для этого пути.
	desc := graphicsCompositionDesc{width: 1, height: 1, format: 28, sampleCount: 1, usage: 0x20, count: 2, effect: 3, alpha: 1}
	if err := graphicsResult("CreateSwapChainForComposition", factory.call(24, uintptr(unsafe.Pointer(p.device)), uintptr(unsafe.Pointer(&desc)), 0, uintptr(unsafe.Pointer(&p.swap)))); err != nil {
		return err
	}
	var err error
	c.library, err = windows.LoadLibraryEx("dcomp.dll", 0, windows.LOAD_LIBRARY_SEARCH_SYSTEM32)
	if err != nil {
		return err
	}
	create, err := windows.GetProcAddress(c.library, "DCompositionCreateDevice")
	if err != nil {
		return err
	}
	result, _, _ := syscall.SyscallN(create, uintptr(unsafe.Pointer(dxgi)), uintptr(unsafe.Pointer(&graphicsDCompIID)), uintptr(unsafe.Pointer(&c.device)))
	if err := graphicsResult("DCompositionCreateDevice", result); err != nil {
		return err
	}
	if err := graphicsResult("DirectComposition CreateTargetForHwnd", c.device.call(6, p.hwnd, 1, uintptr(unsafe.Pointer(&c.target)))); err != nil {
		return err
	}
	if err := graphicsResult("DirectComposition CreateVisual", c.device.call(7, uintptr(unsafe.Pointer(&c.visual)))); err != nil {
		return err
	}
	return graphicsResult("DirectComposition SetContent", c.visual.call(15, uintptr(unsafe.Pointer(p.swap))))
}

func (p *wglDXGIPresenter) commitComposition() error {
	c := p.composition
	if c.attached {
		return nil // Обычный кадр требует только Present, без Commit/ожидания.
	}
	if err := graphicsResult("DirectComposition SetRoot", c.target.call(3, uintptr(unsafe.Pointer(c.visual)))); err != nil {
		return err
	}
	if err := graphicsResult("DirectComposition Commit", c.device.call(3)); err != nil {
		return err
	}
	if err := graphicsResult("DirectComposition first commit", c.device.call(4)); err != nil {
		return err
	}
	c.attached = true
	return nil
}

func (p *wglDXGIPresenter) destroyComposition() {
	c := p.composition
	if c == nil {
		return
	}
	if c.target != nil {
		c.target.call(3, 0)
	}
	if c.visual != nil {
		c.visual.call(15, 0)
	}
	if c.device != nil {
		c.device.call(3)
		// Завершение транзакции перед Release позволяет немедленно создать
		// новое дерево для того же HWND при восстановлении устройства.
		c.device.call(4)
	}
	releaseGraphicsCOM(&c.visual)
	releaseGraphicsCOM(&c.target)
	releaseGraphicsCOM(&c.device)
	if c.library != 0 {
		windows.FreeLibrary(c.library)
	}
	p.composition = nil
}
