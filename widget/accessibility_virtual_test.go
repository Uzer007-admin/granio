package widget

import (
	"fmt"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/stretchr/testify/require"
)

func TestAccessibilityVirtualCollectionsDoNotRealizeOnRead(t *testing.T) {
	test.NewApp()
	calls := 0
	l := NewList(func() int { return 1000000 }, func() granio.CanvasObject { return NewLabel("") }, func(id ListItemID, object granio.CanvasObject) {
		calls++
		object.(*Label).SetText(fmt.Sprintf("row %d 日本", id))
	})
	require.Equal(t, 1000000, l.AccessibilityVirtualCount(""))
	require.Zero(t, calls)
	item, ok := l.AccessibilityVirtualItem("999999")
	require.True(t, ok)
	require.Equal(t, "row 999999 日本", item.Label)
	require.Equal(t, 1, calls)
	require.Nil(t, l.scroller)
	require.True(t, l.AccessibilityVirtualAction("999999", "select"))
	require.Equal(t, []string{"999999"}, l.AccessibilityVirtualSelection())
	_, ok = l.AccessibilityVirtualItem("1000000")
	require.False(t, ok)
	table := NewTable(func() (int, int) { return 100000, 3 }, func() granio.CanvasObject { return NewLabel("") }, func(id TableCellID, object granio.CanvasObject) {
		object.(*Label).SetText(fmt.Sprintf("%d:%d", id.Row, id.Col))
	})
	cell, ok := table.AccessibilityVirtualItem("99999:2")
	require.True(t, ok)
	require.Equal(t, "99999:2", cell.Label)
	require.Nil(t, table.cells)
}
func TestAccessibilityVirtualClosedTree(t *testing.T) {
	test.NewApp()
	tree := NewTree(func(id string) []string {
		switch id {
		case "":
			return []string{"branch", "other"}
		case "branch":
			return []string{"child"}
		}
		return nil
	}, func(id string) bool { return id == "" || id == "branch" }, func(bool) granio.CanvasObject { return NewLabel("") }, func(id string, _ bool, object granio.CanvasObject) { object.(*Label).SetText(id) })
	require.False(t, tree.IsBranchOpen("branch"))
	item, ok := tree.AccessibilityVirtualItem("child")
	require.True(t, ok)
	require.Equal(t, "branch", item.Parent)
	require.Zero(t, item.Index)
	require.False(t, tree.IsBranchOpen("branch"))
	require.True(t, tree.AccessibilityVirtualAction("child", "select"))
	require.Equal(t, []string{"child"}, tree.AccessibilityVirtualSelection())
	require.False(t, tree.IsBranchOpen("branch"))
}
