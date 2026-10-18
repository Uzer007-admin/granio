package widget

import (
	"strings"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/stretchr/testify/require"
)

func TestAccessibilityEditableUnicode(t *testing.T) {
	e := NewEntry()
	e.SetText("a界z")
	require.True(t, e.AccessibilitySetSelection(1, 2))
	cursor, start, end := e.AccessibilityTextSelection()
	require.Equal(t, 2, cursor)
	require.Equal(t, 1, start)
	require.Equal(t, 2, end)
	changes := 0
	e.OnChanged = func(string) { changes++ }
	require.True(t, e.AccessibilitySetText("日本語"))
	require.Equal(t, 1, changes)
	require.Equal(t, "日本語", e.Text)
	e.TypedShortcut(&granio.ShortcutUndo{})
	require.Equal(t, "a界z", e.Text)
	e.Password = true
	require.Empty(t, e.AccessibilityValue())
	cursor, start, end = e.AccessibilityTextSelection()
	require.Zero(t, cursor+start+end)
	e.Disable()
	require.False(t, e.AccessibilitySetText("x"))
	require.False(t, e.AccessibilityPerformAction("focus"))
	check := NewCheck("Enabled", nil)
	require.True(t, check.AccessibilityPerformAction("toggle"))
	require.True(t, check.AccessibilityState().Checked)
}

func TestAccessibilityScrollKeepsEditingState(t *testing.T) {
	e := NewMultiLineEntry()
	e.SetText(strings.Repeat("alpha 日本\n", 80))
	e.Resize(granio.NewSize(240, 100))
	cache.Renderer(e).Layout(e.Size())
	require.True(t, e.AccessibilitySetSelection(1, 3))
	cursor, start, end := e.AccessibilityTextSelection()
	before := e.Text
	changes := 0
	e.OnChanged = func(string) { changes++ }
	require.True(t, e.AccessibilityScrollTextRange(180, 200, true))
	require.Greater(t, e.scroller.Offset.Y, float32(0))
	a, b, c := e.AccessibilityTextSelection()
	require.Equal(t, []int{cursor, start, end}, []int{a, b, c})
	require.Equal(t, before, e.Text)
	require.Zero(t, changes)
	require.False(t, e.AccessibilityScrollTextRange(-1, 2, true))
	e.Password = true
	require.False(t, e.AccessibilityScrollTextRange(1, 2, true))
}
