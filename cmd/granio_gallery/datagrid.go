package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/widget"
)

type galleryDataRow struct {
	name, status string
	size         int
}
type galleryDataSource struct {
	ids     []string
	indexes map[string]int
	rows    []galleryDataRow
}

func (s *galleryDataSource) RowIDs() []string { return s.ids }
func (s *galleryDataSource) Value(row, col string) string {
	i, ok := s.indexes[row]
	if !ok {
		return ""
	}
	r := s.rows[i]
	switch col {
	case "name":
		return r.name
	case "status":
		return r.status
	case "size":
		return strconv.Itoa(r.size)
	}
	return ""
}
func (s *galleryDataSource) SetValue(row, col, value string) error {
	i, ok := s.indexes[row]
	if !ok {
		return widget.ErrDataGridCell
	}
	switch col {
	case "name":
		s.rows[i].name = value
	case "status":
		s.rows[i].status = value
	default:
		return widget.ErrDataGridReadOnly
	}
	return nil
}
func galleryDataGrid() granio.CanvasObject {
	source := &galleryDataSource{indexes: make(map[string]int, 10000)}
	statuses := []string{"Готово", "Ожидает", "Проверка"}
	for i := 0; i < 10000; i++ {
		id := fmt.Sprintf("row-%05d", i)
		source.indexes[id] = i
		source.ids = append(source.ids, id)
		source.rows = append(source.rows, galleryDataRow{name: fmt.Sprintf("Документ %05d · Русский", i), status: statuses[i%len(statuses)], size: 1024 + i*173})
	}
	grid, err := widget.NewDataGrid([]widget.DataGridColumn{
		{ID: "name", Title: "Название", Width: 310, Editable: true, Validate: func(v string) error {
			if strings.TrimSpace(v) == "" {
				return errors.New("Название не должно быть пустым")
			}
			return nil
		}},
		{ID: "status", Title: "Статус", Width: 150, Editable: true, Validate: func(v string) error {
			for _, status := range statuses {
				if v == status {
					return nil
				}
			}
			return errors.New("Допустимо: Готово, Ожидает, Проверка")
		}},
		{ID: "size", Title: "Размер, байт", Width: 140, Compare: func(a, b string) int {
			x, _ := strconv.Atoi(a)
			y, _ := strconv.Atoi(b)
			if x < y {
				return -1
			}
			if x > y {
				return 1
			}
			return 0
		}},
	}, source)
	if err != nil {
		return widget.NewLabel(err.Error())
	}
	grid.Table.StickyColumnCount = 1
	status := widget.NewLabel("10 000 строк · сортировка заголовком · редактирование F2 или двойным нажатием")
	status.Wrapping = granio.TextWrapWord
	grid.OnEditError = func(err error) { status.SetText(err.Error()) }
	grid.OnEdited = func(cell widget.DataGridCell, value string) {
		status.SetText("Сохранено: " + cell.RowID + " / " + cell.ColumnID + " = " + value)
	}
	filter := widget.NewEntry()
	filter.PlaceHolder = "Фильтр названий и статусов"
	filter.OnChanged = func(value string) {
		query := strings.ToLower(strings.TrimSpace(value))
		_ = grid.SetFilter(func(id string) bool {
			return strings.Contains(strings.ToLower(source.Value(id, "name")+" "+source.Value(id, "status")), query)
		})
		status.SetText(fmt.Sprintf("Строк представления с группами: %d", grid.ViewRowCount()))
	}
	group := widget.NewSelect([]string{"Без группировки", "По статусу"}, func(value string) {
		column := ""
		if value == "По статусу" {
			column = "status"
		}
		_ = grid.SetGroupColumn(column)
	})
	group.SetSelected("Без группировки")
	edit := widget.NewButton("Редактировать", func() {
		cells := grid.SelectedCells()
		if len(cells) == 0 {
			status.SetText("Выберите ячейку названия или статуса")
			return
		}
		if err := grid.BeginEdit(cells[0]); err != nil {
			status.SetText(err.Error())
		}
	})
	cancel := widget.NewButton("Отменить правку", grid.CancelEdit)
	return container.NewBorder(container.NewVBox(filter, container.NewHBox(group, edit, cancel), status), nil, nil, nil, grid)
}
