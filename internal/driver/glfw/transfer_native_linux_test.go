//go:build linux && !wayland && !no_glfw && !wasm && !test_web_driver

package glfw

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/build"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/require"
)

func TestExternalX11DataDrop(t *testing.T) {
	source := os.Getenv("FYNE_XDND_SOURCE")
	if os.Getenv("FYNE_TEST_DRAGDROP") != "1" || source == "" || build.IsWayland {
		t.Skip("requires independent Xdnd source on X11")
	}
	w := createWindowWithResizeCallback("External Xdnd acceptance")
	defer w.Close()
	w.SetContent(widget.NewLabel("Native drop target"))
	received := make(chan map[string][]byte, 4)
	var handle uintptr
	var setupError error
	runOnMain(func() {
		setupError = w.window.SetOnDataDropped([]string{"application/x-fyne-drag-test", granio.MIMETextHTML, granio.MIMETextPlain, granio.MIMETextURIList}, func(position granio.Position, data map[string][]byte) { received <- data })
		w.window.Resize(granio.NewSize(360, 240))
		w.window.Show()
		handle = uintptr(w.viewport.GetX11Window())
	})
	require.NoError(t, setupError)
	for _, fixture := range []struct {
		name, mime string
		data       []byte
	}{{"unicode", granio.MIMETextPlain, []byte("Привет 日本 😀")}, {"html", granio.MIMETextHTML, []byte("<b>日本 😀</b>")}, {"files", granio.MIMETextURIList, []byte("file:///tmp/example%20file.txt\r\n")}, {"incr", "application/x-fyne-drag-test", bytes.Repeat([]byte{0, 7, 255, 17}, 128*1024)}} {
		t.Run(fixture.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "payload")
			require.NoError(t, os.WriteFile(filename, fixture.data, 0600))
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, source, fmt.Sprintf("%d", handle), fixture.mime, filename)
			output, err := command.CombinedOutput()
			require.NoError(t, err, string(output))
			select {
			case data := <-received:
				require.Equal(t, fixture.data, data[fixture.mime])
				t.Logf("external Xdnd %s: %d bytes", fixture.mime, len(fixture.data))
			case <-time.After(2 * time.Second):
				t.Fatal("missing external drop callback")
			}
		})
	}
}
