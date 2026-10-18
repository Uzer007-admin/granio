//go:build windows && !android && !ios && !mobile && !tinygo && !noos && !tamago

package theme

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func TestSystemUIFontUsesInstalledFacesAndKeepsCodeFallback(t *testing.T) {
	dir, err := windows.GetWindowsDirectory()
	require.NoError(t, err)
	if _, err := os.Stat(filepath.Join(dir, "Fonts", "segoeui.ttf")); err != nil {
		t.Skip("Segoe UI is not installed; bundled fallback is covered by portable tests")
	}
	m, err := NewModern(ModernOptions{FontPolicy: FontSystem})
	require.NoError(t, err)
	require.Equal(t, "segoeui.ttf", m.Font(granio.TextStyle{}).Name())
	require.Equal(t, DefaultTheme().Font(granio.TextStyle{Monospace: true}), m.Font(granio.TextStyle{Monospace: true}))
}
