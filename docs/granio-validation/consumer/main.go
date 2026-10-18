package main

import (
	"fmt"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/widget"
)

type rows struct{}
func (rows) RowIDs() []string { return []string{"first", "second"} }
func (rows) Value(row, _ string) string { return row }

func main() {
	a := app.NewWithID("io.granio.consumer-check")
	w := a.NewWindow("Проверка установки Granio")
	grid, err := widget.NewDataGrid([]widget.DataGridColumn{{ID:"name", Title:"Название", Width:240}}, rows{})
	if err != nil { panic(err) }
	if err := grid.SetSort("name",true); err != nil { panic(err) }
	w.SetContent(grid)
	w.Resize(granio.NewSize(480,320))
	go func() {
		time.Sleep(700*time.Millisecond)
		granio.Do(func(){
			capture:=w.Canvas().Capture()
			if capture.Bounds().Empty(){panic("пустой кадр")}
			fmt.Println("Независимый потребитель Granio:",capture.Bounds())
			w.Close();a.Quit()
		})
	}()
	w.ShowAndRun()
}
