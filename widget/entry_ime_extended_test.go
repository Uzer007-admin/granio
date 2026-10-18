package widget

import (
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIMEReplacementAndRichUndo(t *testing.T) {
	e := NewEntry()
	e.SetText("a界z")
	e.AccessibilitySetSelection(2, 2)
	changes := 0
	e.OnChanged = func(string) { changes++ }
	e.CommitTextEdit("語", 1, 1)
	require.Equal(t, "a語", e.Text)
	require.Equal(t, 1, changes)
	e.Undo()
	require.Equal(t, "a界z", e.Text)
	rich := NewRichTextEntry()
	rich.SetText("abc")
	rich.AccessibilitySetSelection(1, 2)
	rich.UpdateComposition(desktop.CompositionEvent{Text: "日本語", Cursor: 2, SelectionStart: 1, SelectionEnd: 2})
	require.Equal(t, "abc", rich.Text)
	require.Len(t, rich.compositionSegments(), 5)
	rich.CommitComposition("界")
	require.Equal(t, "a界c", rich.Text)
	rich.Undo()
	require.Equal(t, "abc", rich.Text)
	rich.Redo()
	require.Equal(t, "a界c", rich.Text)
}
