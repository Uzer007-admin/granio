package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/data/binding"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEntryCompositionIsUncommittedAndAtomic(t *testing.T) {
	test.NewApp()
	data := binding.NewString()
	require.NoError(t, data.Set("a界z"))
	entry := NewEntryWithData(data)
	entry.Resize(granio.NewSize(240, 40))
	entry.CursorColumn = 1
	changes := 0
	entry.OnChanged = func(string) { changes++ }
	entry.UpdateComposition(desktop.CompositionEvent{Text: "に", Cursor: 1})
	require.Equal(t, "a界z", entry.Text)
	require.Equal(t, "aに界z", entry.textProvider().String())
	entry.UpdateComposition(desktop.CompositionEvent{Text: "日本", Cursor: 2})
	require.Equal(t, "a日本界z", entry.textProvider().String())
	committed, _ := data.Get()
	require.Equal(t, "a界z", committed)
	require.Zero(t, changes)
	entry.CommitComposition("日本語")
	require.Equal(t, "a日本語界z", entry.Text)
	require.Equal(t, 1, changes)
	committed, _ = data.Get()
	require.Equal(t, entry.Text, committed)
	entry.Undo()
	require.Equal(t, "a界z", entry.Text)
	entry.Redo()
	require.Equal(t, "a日本語界z", entry.Text)
}

func TestEntryCompositionCancelSelectionAndPasswordPrivacy(t *testing.T) {
	test.NewApp()
	entry := NewEntry()
	entry.SetText("selected")
	entry.Resize(granio.NewSize(240, 40))
	entry.selectAll()
	entry.UpdateComposition(desktop.CompositionEvent{Text: "漢字", Cursor: 2})
	require.Equal(t, "漢字", entry.textProvider().String())
	entry.CancelComposition()
	require.Equal(t, "selected", entry.Text)
	require.True(t, entry.sel.selecting)
	entry.CommitComposition("置換")
	require.Equal(t, "置換", entry.Text)
	entry.Undo()
	require.Equal(t, "selected", entry.Text)
	entry.Password = true
	state := entry.TextInputState()
	require.True(t, state.Sensitive)
	require.Empty(t, state.Text)
	entry.SetText("")
	entry.UpdateComposition(desktop.CompositionEvent{Text: "かな", Cursor: 2})
	require.Equal(t, 2, entry.CursorTextOffset())
	entry.FocusLost()
	require.False(t, entry.CompositionActive())
	require.Empty(t, entry.Text)
}
