//go:build windows && !no_glfw && !mobile

package glfw

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClipboardWindowsNativeCodecs(t *testing.T) {
	fragment := []byte("<b>日本語</b>")
	require.Equal(t, fragment, decodeClipboardHTML(encodeClipboardHTML(fragment)))
	files := encodeClipboardFiles([]byte("file:///C:/My%20Files/日本.txt\r\nfile://server/share/a.txt\r\n"))
	require.Equal(t, uint32(20), binary.LittleEndian.Uint32(files))
	require.Equal(t, uint32(1), binary.LittleEndian.Uint32(files[16:]))
	require.Nil(t, encodeClipboardFiles([]byte("https://example.org/")))
	require.Error(t, validateClipboardFormat("text/plain\x00invalid"))
	id, err := winClipboardFormat("application/x-windows-clipboard-format;id=15")
	require.NoError(t, err)
	require.Equal(t, uint32(15), id)
}
