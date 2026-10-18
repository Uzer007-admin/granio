//go:build windows && !no_glfw && !wasm && !test_web_driver

package glfw

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/require"
)

func externalDropFixture(name string) (string, uint32, []byte, []byte) {
	switch name {
	case "text":
		text := []byte("Привет 日本 😀")
		units := utf16.Encode([]rune(string(text)))
		data := make([]byte, (len(units)+1)*2)
		for i, unit := range units {
			binary.LittleEndian.PutUint16(data[i*2:], unit)
		}
		return granio.MIMETextPlain, 13, data, text
	case "html":
		text := []byte("<b>日本 😀</b>")
		id, _ := winClipboardFormat(granio.MIMETextHTML)
		return granio.MIMETextHTML, id, encodeClipboardHTML(text), text
	case "binary":
		data := bytes.Repeat([]byte{0, 7, 255, 17}, 128*1024)
		id, _ := winClipboardFormat("application/x-fyne-drag-test")
		return "application/x-fyne-drag-test", id, data, data
	case "files":
		units := utf16.Encode([]rune("C:\\Fyne drag\\日本.txt\x00D:\\folder\\two.txt\x00\x00"))
		data := make([]byte, 20+len(units)*2)
		binary.LittleEndian.PutUint32(data, 20)
		binary.LittleEndian.PutUint32(data[16:], 1)
		for i, unit := range units {
			binary.LittleEndian.PutUint16(data[20+i*2:], unit)
		}
		return granio.MIMETextURIList, 15, data, []byte("file:///C:/Fyne%20drag/%E6%97%A5%E6%9C%AC.txt\r\nfile:///D:/folder/two.txt\r\n")
	}
	panic("unknown drag fixture")
}

func TestExternalWindowsDropSource(t *testing.T) {
	name := os.Getenv("FYNE_TEST_DROP_SOURCE")
	if name == "" {
		t.Skip("separate OLE source process")
	}
	hwnd, err := strconv.ParseUint(os.Getenv("FYNE_TEST_DROP_HWND"), 10, 64)
	require.NoError(t, err)
	_, format, payload, _ := externalDropFixture(name)
	t.Log("source waiting for STA")
	runOnMain(func() {
		t.Log("source entering OLE on STA")
		require.True(t, windowsExternalDropTest(uintptr(hwnd), format, payload), "external DoDragDrop failed")
	})
}

func TestExternalWindowsOLEDrop(t *testing.T) {
	if os.Getenv("FYNE_TEST_DRAGDROP") != "1" {
		t.Skip("requires native external OLE drag acceptance")
	}
	w := createWindowWithResizeCallback("External drag acceptance")
	defer w.Close()
	w.SetContent(widget.NewLabel("Drop target"))
	received := make(chan map[string][]byte, 4)
	files := make(chan []granio.URI, 1)
	var hwnd uintptr
	runOnMain(func() {
		w.window.SetOnDropped(func(_ granio.Position, uris []granio.URI) { files <- uris })
		require.NoError(t, w.window.SetOnDataDropped([]string{"application/x-fyne-drag-test", granio.MIMETextHTML, granio.MIMETextPlain}, func(_ granio.Position, data map[string][]byte) { received <- data }))
		w.window.Resize(granio.NewSize(360, 220))
		w.window.Show()
		hwnd = uintptr(unsafe.Pointer(w.viewport.GetWin32Window()))
	})
	for _, name := range []string{"text", "html", "binary", "files"} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			executable, err := os.Executable()
			require.NoError(t, err)
			command := exec.CommandContext(ctx, executable, "-test.run=^TestExternalWindowsDropSource$", "-test.count=1", "-test.v", "-test.timeout=8s")
			command.Env = append(os.Environ(), "FYNE_TEST_DROP_SOURCE="+name, fmt.Sprintf("FYNE_TEST_DROP_HWND=%d", hwnd))
			output, err := command.CombinedOutput()
			require.NoError(t, err, string(output))
			mime, _, _, expected := externalDropFixture(name)
			if name == "files" {
				select {
				case uris := <-files:
					require.Len(t, uris, 2)
					require.Equal(t, string(expected), uris[0].String()+"\r\n"+uris[1].String()+"\r\n")
					t.Log("external CF_HDROP reached the legacy file callback alongside a MIME target")
				case <-time.After(3 * time.Second):
					t.Fatal("OLE file offer was lost behind the MIME target")
				}
				select {
				case value := <-received:
					t.Fatalf("unrequested MIME representation delivered: %v", value)
				default:
				}
				return
			}
			select {
			case value := <-received:
				require.Equal(t, expected, value[mime])
				t.Logf("external OLE %s: %d bytes", mime, len(expected))
			case <-time.After(3 * time.Second):
				t.Fatal("external OLE did not deliver data")
			}
		})
	}
}
