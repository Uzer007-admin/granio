//go:build windows && !no_glfw && !wasm && !test_web_driver

package glfw

import (
	"os"
	"testing"
	"time"
	"unsafe"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/require"
)

func TestRealMicrosoftJapaneseIME(t *testing.T) {
	if os.Getenv("FYNE_TEST_REAL_IME") != "1" {
		t.Skip("explicit native Japanese IME acceptance")
	}
	w := createWindowWithResizeCallback("Real Microsoft Japanese IME acceptance")
	defer w.Close()
	entry := widget.NewEntry()
	var hwnd uintptr
	var status int64
	runOnMain(func() {
		w.window.SetContent(entry)
		w.window.Resize(granio.NewSize(360, 180))
		w.window.Show()
		w.canvas.Focus(entry)
		w.window.updateInputMethod()
		hwnd = uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
		status = windowsRealJapaneseBegin(hwnd)
	})
	defer runOnMain(windowsRealJapaneseEnd)
	t.Logf("Microsoft IME activation stage=%d HRESULT=0x%x", windowsRealJapaneseStep(), uint32(status))
	require.Zero(t, status, "Microsoft IME activation HRESULT 0x%x", uint32(status))
	time.Sleep(300 * time.Millisecond)
	runOnMain(func() { status = windowsRealJapaneseOpen(hwnd) })
	t.Logf("Microsoft IME open/conversion HRESULT=0x%x", uint32(status))
	require.Zero(t, status)
	time.Sleep(300 * time.Millisecond)
	for _, key := range "NIHONGO" {
		require.True(t, windowsRealIMEKey(hwnd, int(key)))
		time.Sleep(100 * time.Millisecond)
	}
	var preedit bool
	var text string
	runOnMain(func() { preedit = entry.CompositionActive(); text = entry.Text })
	t.Logf("real Microsoft preedit active=%t committed=%q", preedit, text)
	require.True(t, windowsRealIMEKey(hwnd, 0x20))
	time.Sleep(250 * time.Millisecond)
	require.True(t, windowsRealIMEKey(hwnd, 0x0d))
	require.Eventually(t, func() bool { runOnMain(func() { text = entry.Text }); return text == "日本語" }, 5*time.Second, 20*time.Millisecond, "real Microsoft IME committed %q", text)
	runOnMain(func() { entry.Undo(); text = entry.Text })
	require.Empty(t, text, "one native composition must be one undo action")
}
