package glfw

import (
	"fmt"
	"unsafe"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver"
	"github.com/Uzer007-admin/granio/v2/internal/scale"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var setWindowThemeAttribute = windows.NewLazySystemDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute")

func (w *window) setDarkMode() {
	view := w.view()
	if view == nil {
		return
	}
	var dark int32 // Win32 BOOL
	if isDark() {
		dark = 1
	}
	result, _, _ := setWindowThemeAttribute.Call(uintptr(unsafe.Pointer(view.GetWin32Window())),
		20, uintptr(unsafe.Pointer(&dark)), unsafe.Sizeof(dark)) // DWMWA_USE_IMMERSIVE_DARK_MODE
	if int32(result) < 0 && uint32(result) != 0x80070057 { // Older Windows does not know this attribute.
		granio.LogError("Failed to set dark mode", fmt.Errorf("DwmSetWindowAttribute: HRESULT 0x%08x", uint32(result)))
	}
}

func isDark() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil { // older version of Windows will not have this key
		return false
	}
	defer k.Close()

	useLight, _, err := k.GetIntegerValue("AppsUseLightTheme")
	if err != nil { // older version of Windows will not have this value
		return false
	}

	return useLight == 0
}

func (w *window) computeCanvasSize(width, height int) granio.Size {
	if w.fixedSize && w.width > 0 && w.height > 0 {
		return granio.NewSize(scale.ToFyneCoordinate(w.canvas, w.width), scale.ToFyneCoordinate(w.canvas, w.height))
	}
	return granio.NewSize(scale.ToFyneCoordinate(w.canvas, width), scale.ToFyneCoordinate(w.canvas, height))
}

// assert we are implementing driver.NativeWindow
var _ driver.NativeWindow = (*window)(nil)

func (w *window) RunNative(f func(any)) {
	context := driver.WindowsWindowContext{}
	if v := w.view(); v != nil {
		context.HWND = uintptr(unsafe.Pointer(v.GetWin32Window()))
	}

	f(context)
}
