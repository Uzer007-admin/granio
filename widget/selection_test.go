package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestListMultipleSelectionToggleRangeAndCallbacks(t *testing.T) {
	test.NewApp()
	list := NewList(func() int { return 10 }, func() granio.CanvasObject { return NewLabel("row") }, func(int, granio.CanvasObject) {})
	list.SelectionMode = SelectionMultiple
	list.Resize(granio.NewSize(200, 300))
	var added, removed []int
	list.OnSelected = func(id int) { added = append(added, id) }
	list.OnUnselected = func(id int) { removed = append(removed, id) }
	list.pointerSelect(1, 0)
	list.pointerSelect(3, granio.KeyModifierControl)
	require.Equal(t, []int{1, 3}, list.Selected())
	list.pointerSelect(5, granio.KeyModifierShift)
	require.Equal(t, []int{3, 4, 5}, list.Selected())
	list.pointerSelect(4, granio.KeyModifierControl)
	require.Equal(t, []int{3, 5}, list.Selected())
	require.Equal(t, []int{1, 3, 4, 5}, added)
	require.Equal(t, []int{1, 4}, removed)
	copy := list.Selected()
	copy[0] = 9
	require.Equal(t, []int{3, 5}, list.Selected())
	list.TypedShortcut(&granio.ShortcutSelectAll{})
	require.Len(t, list.Selected(), 10)
	list.Unselect(7)
	require.Len(t, list.Selected(), 9)
}

func TestTreeMultipleSelectionUsesVisibleOrder(t *testing.T) {
	test.NewApp()
	tree := NewTreeWithStrings(map[string][]string{"": {"z", "a"}, "z": {"child"}})
	tree.SelectionMode = SelectionMultiple
	tree.Resize(granio.NewSize(220, 250))
	tree.OpenBranch("z")
	tree.pointerSelect("z", 0)
	tree.pointerSelect("a", granio.KeyModifierShift)
	require.Equal(t, []string{"z", "child", "a"}, tree.Selected())
	tree.CloseBranch("z")
	require.True(t, tree.isSelected("child"), "closing a branch lost selected descendants")
	tree.pointerSelect("z", granio.KeyModifierControl)
	require.Equal(t, []string{"child", "a"}, tree.Selected())
	tree.Unselect("child")
	require.Equal(t, []string{"a"}, tree.Selected())
	tree.UnselectAll()
	require.Empty(t, tree.Selected())
}

func TestTableMultipleSelectionMarkersStayBoundedToViewport(t *testing.T) {
	test.NewApp()
	table := NewTable(func() (int, int) { return 10000, 4 }, func() granio.CanvasObject { return NewLabel("cell") }, func(TableCellID, granio.CanvasObject) {})
	table.SelectionMode = SelectionMultiple
	table.Resize(granio.NewSize(240, 200))
	table.SelectRange(TableCellID{}, TableCellID{9999, 3})
	require.Len(t, table.Selected(), 40000)
	content := cache.Renderer(table.cells).(*tableCellsRenderer)
	require.Less(t, len(content.selectionMarkers), 100, "allocated a marker for every selected cell")
	table.pointerSelect(TableCellID{1, 1}, granio.KeyModifierControl)
	require.False(t, table.isSelected(TableCellID{1, 1}))
	table.UnselectAll()
	require.Empty(t, table.Selected())
	for _, marker := range content.selectionMarkers {
		require.False(t, marker.Visible())
	}
}
