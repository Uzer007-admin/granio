//go:build !mobile

package widget

import (
	"fmt"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/test"

	"github.com/stretchr/testify/assert"
)

func TestTable_Hovered(t *testing.T) {
	test.NewTempApp(t)

	table := NewTable(
		func() (int, int) { return 2, 2 },
		func() granio.CanvasObject {
			return NewLabel("placeholder")
		},
		func(id TableCellID, c granio.CanvasObject) {
			c.(*Label).SetText(fmt.Sprintf("Cell %d, %d", id.Row, id.Col))
		},
	)

	w := test.NewWindow(table)
	defer w.Close()
	w.Resize(granio.NewSize(180, 180))

	test.MoveMouse(w.Canvas(), granio.NewPos(35, 50))
	test.MoveMouse(w.Canvas(), granio.NewPos(35, 100))

	assert.Nil(t, table.hoveredCell)

	test.AssertRendersToMarkup(t, "table/desktop/hovered_out.xml", w.Canvas())

	table.Length = func() (int, int) { return 3, 5 }
	table.Refresh()

	w.SetContent(table)
	w.Resize(granio.NewSize(180, 180))
	test.MoveMouse(w.Canvas(), granio.NewPos(35, 58))

	assert.Equal(t, 0, table.hoveredCell.Col)
	assert.Equal(t, 1, table.hoveredCell.Row)

	test.AssertRendersToMarkup(t, "table/desktop/hovered.xml", w.Canvas())
}
