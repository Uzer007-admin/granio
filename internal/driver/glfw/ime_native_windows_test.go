//go:build windows && !no_glfw && !wasm && !test_web_driver

package glfw

import (
	"encoding/binary"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/require"
	"testing"
	"unicode/utf16"
	"unsafe"
)

func TestTSFTextStoreNativeContract(t *testing.T) {
	runOnMain(func() { require.True(t, windowsTSFContractTest(), "ITextStoreACP ABI/lock/UTF-16 contract failed") })
}

func TestTSFNativeEditSession(t *testing.T) {
	w := createWindow("TSF edit session")
	defer w.Close()
	entry := widget.NewEntry()
	w.SetContent(entry)
	runOnMain(func() {
		entry.SetText("a😀z")
		entry.AccessibilitySetSelection(1, 2)
		w.window.canvas.Focus(entry)
		w.window.updateInputMethod()
		hwnd := uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
		require.True(t, windowsTSFEditContractTest(hwnd), "real ITfContext edit session failed")
		require.Equal(t, "a日本z", entry.Text)
		entry.Undo()
		require.Equal(t, "a😀z", entry.Text)
	})
}

func TestIMMReconversionUTF16(t *testing.T) {
	packet, base := reconversionPacket(desktop.TextInputState{Text: "a😀日本z", SelectionStart: 2, SelectionEnd: 4}, false)
	require.Equal(t, 0, base)
	require.EqualValues(t, 6, binary.LittleEndian.Uint32(packet[8:]))
	require.EqualValues(t, 6, binary.LittleEndian.Uint32(packet[20:]), "UTF-16 prefix occupies 6 bytes")
	require.EqualValues(t, 2, binary.LittleEndian.Uint32(packet[16:]))
	units := make([]uint16, 6)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(packet[32+i*2:])
	}
	require.Equal(t, "a😀日本z", string(utf16.Decode(units)))
	private, _ := reconversionPacket(desktop.TextInputState{Text: "secret", Sensitive: true}, false)
	require.Nil(t, private)
}
