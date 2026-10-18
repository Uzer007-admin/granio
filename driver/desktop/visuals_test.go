package desktop_test

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/stretchr/testify/require"
)

func TestWindowVisualsOptionalContract(t *testing.T) {
	window := test.NewWindow(nil)
	defer window.Close()
	_, err := desktop.ConfigureWindowVisuals(window, desktop.VisualOptions{})
	require.ErrorIs(t, err, desktop.ErrVisualUnsupported)
	for _, options := range []desktop.VisualOptions{
		{Backdrop: desktop.BackdropSystemBlur + 1}, {Fallback: desktop.VisualFallbackFail + 1}, {Frame: 1},
	} {
		_, err = desktop.ConfigureWindowVisuals(window, options)
		require.ErrorIs(t, err, desktop.ErrVisualOptions)
	}
	for backdrop := desktop.BackdropOpaque; backdrop <= desktop.BackdropSystemBlur; backdrop++ {
		require.NoError(t, desktop.ValidateVisualOptions(desktop.VisualOptions{Backdrop: backdrop}))
		require.True(t, desktop.BackdropSupport(0x3f).Supports(backdrop))
	}
	require.False(t, desktop.BackdropSupport(0xff).Supports(8))
}
