package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTreeTableColumnsAlignAcrossDepths(t *testing.T) {
	test.NewApp()
	data := map[string][]string{"": {"parent", "last"}, "parent": {"child"}}
	table := NewTreeTable(func(id string) []string { return data[id] }, func(id string) bool { _, ok := data[id]; return ok }, 3,
		func(int) granio.CanvasObject { return NewLabel("cell") },
		func(id string, _ int, cell granio.CanvasObject) { cell.(*Label).SetText(id) })
	table.SetColumnWidth(0, 180)
	table.SetColumnWidth(1, 100)
	table.SetColumnWidth(2, 140)
	table.Resize(granio.NewSize(450, 300))
	table.OpenBranch("parent")
	content := cache.Renderer(table.scroller.Content.(*treeContent)).(*treeContentRenderer)
	for _, id := range []string{"parent", "child", "last"} {
		var node *treeNode
		if branch := content.branches[id]; branch != nil {
			node = branch.treeNode
		} else {
			node = content.leaves[id].treeNode
		}
		row := node.content.(*treeTableRow)
		// Content coordinates include the expander and depth indentation.
		require.Equal(t, float32(180), row.Position().X+row.cells[1].Position().X, id)
		require.Equal(t, float32(280), row.Position().X+row.cells[2].Position().X, id)
		require.Equal(t, id, row.cells[1].(*Label).Text)
	}
	table.SelectionMode = SelectionMultiple
	table.SelectRange("parent", "last")
	require.Len(t, table.Selected(), 3)
	table.CloseBranch("parent")
	require.Len(t, content.rows, 2)
}
