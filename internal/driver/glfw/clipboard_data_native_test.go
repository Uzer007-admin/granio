//go:build !no_glfw && !mobile && (windows || linux) && !wasm && !test_web_driver

package glfw

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/stretchr/testify/require"
)

const clipboardTestBinary = "application/x-fyne-test-binary"

func clipboardTestPayload() []byte { return bytes.Repeat([]byte{0, 1, 0, 255, 4, 0, 6, 7}, 65536) }

func TestClipboardExternalTransferChild(t *testing.T) {
	if os.Getenv("FYNE_TEST_CLIPBOARD_CHILD") != "1" {
		t.Skip("native clipboard client subprocess")
	}
	clipboard := NewClipboard().(granio.ClipboardData)
	formats, err := clipboard.Formats()
	require.NoError(t, err)
	require.Contains(t, formats, clipboardTestBinary)
	data, err := clipboard.ReadData(clipboardTestBinary)
	require.NoError(t, err)
	require.Equal(t, clipboardTestPayload(), data)
	data, err = clipboard.ReadData(granio.MIMETextPlain)
	require.NoError(t, err)
	require.Equal(t, "日本語 / Привет", string(data))
	data, err = clipboard.ReadData(granio.MIMETextHTML)
	require.NoError(t, err)
	require.Equal(t, "<b>日本語</b>", string(data))
}

func TestClipboardMIMENativeTransfer(t *testing.T) {
	if os.Getenv("FYNE_TEST_CLIPBOARD") != "1" {
		t.Skip("requires a temporary native clipboard transfer")
	}
	clipboard := NewClipboard().(granio.ClipboardData)
	// Retain existing representations only in memory. An unsupported native
	// object cannot be reconstructed safely, so leave such a clipboard untouched.
	previous := make(map[string][]byte)
	formats, err := clipboard.Formats()
	if err == nil {
		for _, format := range formats {
			value, readErr := clipboard.ReadData(format)
			if readErr != nil {
				t.Skip("existing native clipboard contains a representation that cannot be restored")
			}
			previous[format] = value
		}
	}
	if err != nil {
		t.Skip("native clipboard is unavailable")
	}
	t.Cleanup(func() {
		current, err := clipboard.ReadData(clipboardTestBinary)
		if err == nil && bytes.Equal(current, clipboardTestPayload()) {
			require.NoError(t, clipboard.WriteData(previous))
		}
	})
	data := map[string][]byte{granio.MIMETextPlain: []byte("日本語 / Привет"), granio.MIMETextHTML: []byte("<b>日本語</b>"), clipboardTestBinary: clipboardTestPayload()}
	require.NoError(t, clipboard.WriteData(data))
	actual, err := clipboard.ReadData(clipboardTestBinary)
	require.NoError(t, err)
	require.Equal(t, data[clipboardTestBinary], actual)
	start := time.Now()
	_, err = clipboard.ReadData("application/x-fyne-missing")
	require.Error(t, err)
	require.Less(t, time.Since(start), 100*time.Millisecond)
	command := exec.Command(os.Args[0], "-test.run=^TestClipboardExternalTransferChild$", "-test.timeout=20s", "-test.v")
	command.Env = append(os.Environ(), "FYNE_TEST_CLIPBOARD_CHILD=1", "FYNE_TEST_CLIPBOARD=0")
	output, err := command.CombinedOutput()
	require.NoError(t, err, "%s", output)
	t.Log("independent native client verified Unicode, HTML and a 512 KiB binary representation")
}
