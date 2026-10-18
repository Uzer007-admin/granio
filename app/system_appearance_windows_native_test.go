//go:build system_appearance_integration && windows && cgo && !ci && !wasm && !test_web_driver && !android && !ios && !mobile && !tinygo && !noos && !tamago

package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/stretchr/testify/require"
)

// Reads actual OS preferences; neither this test nor its fixture changes them.
func TestNativeWindowsAppearanceSnapshotAndLifetime(t *testing.T) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	baseline := nativeAppearanceResources()
	for i := range 30 {
		p, err := openNativeAppearanceProbe()
		require.NoError(t, err)
		defer p.close()
		snapshot, events := p.read()
		diagnostics := p.diagnostics()
		require.True(t, snapshot.Available)
		require.True(t, snapshot.ColorScheme.Available)
		require.True(t, snapshot.Accent.Available)
		require.True(t, snapshot.HighContrast.Available)
		require.True(t, snapshot.ReducedMotion.Available)
		require.True(t, snapshot.TextScale.Available)
		require.True(t, snapshot.Transparency.Available)
		require.True(t, snapshot.Colors.Available)
		require.GreaterOrEqual(t, snapshot.TextScale.Value, float32(1))
		require.LessOrEqual(t, snapshot.TextScale.Value, float32(2.25))
		require.EqualValues(t, 31, events, "mandatory Windows 11 notifications missing")
		require.Empty(t, snapshot.UnavailableReason)
		for range 1000 {
			p.notify()
		}
		require.Equal(t, 1, p.wait(100), "coalesced native notification lost")
		require.True(t, p.postSettingsMessage())
		require.Equal(t, 1, p.wait(100), "owned WM_SYSCOLORCHANGE message lost")
		start := time.Now()
		cancelDone := make(chan struct{})
		go func() { time.Sleep(10 * time.Millisecond); p.cancel(); close(cancelDone) }()
		for p.wait(1000) > 0 {
			require.Less(t, time.Since(start), 500*time.Millisecond)
		}
		<-cancelDone
		require.Less(t, time.Since(start), 500*time.Millisecond)
		p.close()
		require.Eventually(t, func() bool { return nativeAppearanceResources() == baseline }, time.Second, time.Millisecond, "COM delegates/wake leaked")
		if i == 0 {
			t.Logf("actual OS snapshot: %+v; subscriptions=0x%x", snapshot, events)
			if output := os.Getenv("FYNE_APPEARANCE_REPORT"); output != "" {
				data, err := json.MarshalIndent(map[string]any{"snapshot": snapshot, "subscriptions": events, "native": diagnostics, "resources_after_close": nativeAppearanceResources()}, "", "  ")
				require.NoError(t, err)
				require.NoError(t, os.WriteFile(filepath.Clean(output), data, 0600))
			}
		}
	}
}

func TestNativeWindowsAppearanceInitialAndCancel(t *testing.T) {
	baseline := nativeAppearanceResources()
	modern, err := theme.NewModern(theme.ModernOptions{FollowSystem: true})
	require.NoError(t, err)
	s := &settings{theme: modern, variant: theme.VariantLight, schema: SettingsSchema{ThemeName: "light"}}
	start := time.Now()
	s.startSystemAppearance()
	defer s.stopSystemAppearance()
	snapshot := s.SystemAppearance()
	require.True(t, snapshot.Available, snapshot.UnavailableReason)
	require.EqualValues(t, 1, snapshot.Generation)
	require.Equal(t, 14*snapshot.TextScale.Value, s.Theme().Size(theme.SizeNameText))
	require.Equal(t, theme.VariantLight, s.ThemeVariant(), "explicit variant replaced")
	s.startSystemAppearance()
	before := nativeAppearanceResources()
	time.Sleep(1100 * time.Millisecond)
	require.EqualValues(t, 1, s.SystemAppearance().Generation, "idle provider published redundant snapshots")
	require.Equal(t, before, nativeAppearanceResources())
	s.stopSystemAppearance()
	s.stopSystemAppearance()
	require.False(t, s.SystemAppearance().Available)
	require.Eventually(t, func() bool { return nativeAppearanceResources() == baseline }, time.Second, time.Millisecond)
	t.Logf("initial load and 1.1s idle/lifecycle: %s", time.Since(start))
}
