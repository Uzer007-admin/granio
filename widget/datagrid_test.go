package widget

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/stretchr/testify/require"
)

type gridTestSource struct {
	ids    []string
	values map[DataGridCell]string
	fail   bool
	reads  int
}

func (s *gridTestSource) RowIDs() []string { return s.ids }
func (s *gridTestSource) Value(row, col string) string {
	s.reads++
	return s.values[DataGridCell{row, col}]
}
func (s *gridTestSource) SetValue(row, col, value string) error {
	if s.fail {
		return errors.New("запись отклонена")
	}
	s.values[DataGridCell{row, col}] = value
	return nil
}
func testGrid(t *testing.T) (*DataGrid, *gridTestSource) {
	t.Helper()
	s := &gridTestSource{ids: []string{"a", "b", "c"}, values: map[DataGridCell]string{{"a", "name"}: "Альфа", {"b", "name"}: "Бета", {"c", "name"}: "Гамма", {"a", "count"}: "12", {"b", "count"}: "2", {"c", "count"}: "8", {"a", "group"}: "A", {"b", "group"}: "B", {"c", "group"}: "A"}}
	g, err := NewDataGrid([]DataGridColumn{{ID: "name", Title: "Название", Editable: true, Width: 180, Validate: func(v string) error {
		if v == "" {
			return errors.New("нужно название")
		}
		return nil
	}}, {ID: "count", Title: "Число", Width: 100, Compare: func(a, b string) int { x, _ := strconv.Atoi(a); y, _ := strconv.Atoi(b); return x - y }}, {ID: "group", Title: "Группа", Width: 100}}, s)
	require.NoError(t, err)
	return g, s
}

func TestDataGridProjectionKeepsIdentityAndSourceOrder(t *testing.T) {
	newVisualTestApp(t)
	g, s := testGrid(t)
	g.SetSelection([]DataGridCell{{"a", "name"}})
	require.NoError(t, g.SetSort("count", false))
	id, _, _ := g.ViewRow(0)
	require.Equal(t, "b", id)
	require.Equal(t, []string{"a", "b", "c"}, s.ids)
	require.NoError(t, g.SetFilter(func(id string) bool { return id != "a" }))
	require.Equal(t, []DataGridCell{{"a", "name"}}, g.SelectedCells())
	require.Empty(t, g.Table.Selected())
	require.NoError(t, g.SetFilter(nil))
	require.Equal(t, []TableCellID{{2, 0}}, g.Table.Selected())
	require.NoError(t, g.SetGroupColumn("group"))
	_, name, header := g.ViewRow(0)
	require.True(t, header)
	require.Equal(t, "A", name)
	require.Equal(t, 5, g.ViewRowCount())
	s.ids = []string{"a", "a"}
	require.Error(t, g.Reload())
	require.Equal(t, 5, g.ViewRowCount())
	require.Error(t, g.SetSort("missing", false))
}

func TestDataGridEditingValidationFailureAndKeyboard(t *testing.T) {
	a := newVisualTestApp(t)
	g, s := testGrid(t)
	w := a.NewWindow("таблица")
	defer w.Close()
	w.SetContent(g)
	w.Resize(granio.NewSize(540, 300))
	cell := DataGridCell{"a", "name"}
	g.Table.Select(TableCellID{0, 0})
	g.Table.TypedKey(&granio.KeyEvent{Name: granio.KeyF2})
	_, draft, editing := g.EditingCell()
	require.True(t, editing)
	require.Equal(t, "Альфа", draft)
	require.Error(t, g.CommitEdit(""))
	require.Equal(t, "Альфа", s.Value("a", "name"))
	s.fail = true
	require.Error(t, g.CommitEdit("Новое"))
	require.Equal(t, "Альфа", s.Value("a", "name"))
	s.fail = false
	edited := 0
	g.OnEdited = func(got DataGridCell, value string) { require.Equal(t, cell, got); edited++ }
	require.NoError(t, g.CommitEdit("Новое"))
	require.Equal(t, 1, edited)
	require.Equal(t, "Новое", s.Value("a", "name"))
	require.ErrorIs(t, g.BeginEdit(DataGridCell{"a", "count"}), ErrDataGridReadOnly)
	require.NoError(t, g.BeginEdit(cell))
	require.NotNil(t, g.activeEditor)
	test.Type(g.activeEditor, "черновик")
	_, draft, _ = g.EditingCell()
	require.NoError(t, g.SetSort("count", false))
	_, after, _ := g.EditingCell()
	require.Equal(t, draft, after)
	g.CancelEdit()
	require.Equal(t, "Новое", s.Value("a", "name"))
}

func TestDataGridVirtualizesLargeSource(t *testing.T) {
	a := newVisualTestApp(t)
	s := &gridTestSource{values: map[DataGridCell]string{}}
	for i := 0; i < 10000; i++ {
		id := fmt.Sprintf("%05d", i)
		s.ids = append(s.ids, id)
		s.values[DataGridCell{id, "name"}] = id
	}
	g, err := NewDataGrid([]DataGridColumn{{ID: "name", Title: "Строка", Width: 180}}, s)
	require.NoError(t, err)
	w := a.NewWindow("viewport")
	defer w.Close()
	w.SetContent(g)
	w.Resize(granio.NewSize(500, 320))
	count := 0
	original := g.Table.CreateCell
	g.Table.CreateCell = func() granio.CanvasObject { count++; return original() }
	g.Resize(granio.NewSize(500, 320))
	g.Table.ScrollToBottom()
	g.Refresh()
	require.Less(t, count, 500)
	before := s.reads
	require.NoError(t, g.SetSort("name", true))
	require.Less(t, s.reads-before, 12000)
	id, _, _ := g.ViewRow(0)
	require.Equal(t, "09999", id)
}

func TestDataGridEditingKeepsDraftAndSearchFocus(t *testing.T) {
	a := newVisualTestApp(t)
	g, _ := testGrid(t)
	w := a.NewWindow("Фильтр и редактор")
	defer w.Close()
	search := NewEntry()
	w.SetContent(granio.NewContainer(g, search))
	g.Resize(granio.NewSize(540, 300))
	require.NoError(t, g.BeginEdit(DataGridCell{"a", "name"}))
	require.NotNil(t, g.activeEditor)
	old := g.activeEditor
	old.SetText("Черновик")
	old.AccessibilitySetSelection(2, 5)
	w.Canvas().Focus(search)
	require.NoError(t, g.SetSort("count", true))
	require.Same(t, search, w.Canvas().Focused())
	_, draft, editing := g.EditingCell()
	require.True(t, editing)
	require.Equal(t, "Черновик", draft)
	require.NoError(t, g.SetFilter(func(id string) bool { return id != "a" }))
	require.Same(t, search, w.Canvas().Focused())
	require.NoError(t, g.SetFilter(nil))
	_, draft, _ = g.EditingCell()
	require.Equal(t, "Черновик", draft)
	require.Same(t, search, w.Canvas().Focused())
	require.NotNil(t, g.activeEditor)
	g.activeEditor.TypedKey(&granio.KeyEvent{Name: granio.KeyEscape})
	_, _, editing = g.EditingCell()
	require.False(t, editing)
}

func TestDataGridScrollKeepsDraftWithoutHiddenFocus(t *testing.T) {
	a := newVisualTestApp(t)
	s := &gridTestSource{values: map[DataGridCell]string{}}
	for i := 0; i < 300; i++ {
		id := fmt.Sprintf("%03d", i)
		s.ids = append(s.ids, id)
		s.values[DataGridCell{id, "name"}] = id
	}
	g, err := NewDataGrid([]DataGridColumn{{ID: "name", Title: "Название", Width: 200, Editable: true}}, s)
	require.NoError(t, err)
	w := a.NewWindow("Прокрутка редактора")
	defer w.Close()
	w.SetContent(g)
	w.Resize(granio.NewSize(500, 300))
	require.NoError(t, g.BeginEdit(DataGridCell{"000", "name"}))
	require.NotNil(t, g.activeEditor)
	g.activeEditor.SetText("Черновик")
	g.activeEditor.AccessibilitySetSelection(2, 5)
	g.Table.ScrollToBottom()
	require.Same(t, g.Table, w.Canvas().Focused())
	_, draft, editing := g.EditingCell()
	require.True(t, editing)
	require.Equal(t, "Черновик", draft)
	g.Table.ScrollToTop()
	require.NotNil(t, g.activeEditor)
	require.Equal(t, "Черновик", g.activeEditor.Text)
	_, start, end := g.activeEditor.AccessibilityTextSelection()
	require.Equal(t, 2, start)
	require.Equal(t, 5, end)
	require.Equal(t, "000", s.Value("000", "name"))
}
