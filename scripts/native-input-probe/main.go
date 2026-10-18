//go:build input_probe

// A normal Fyne application for external IME, screen-reader and compositor tests.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"os"
	"sync"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func main() {
	seconds := flag.Int("seconds", 90, "probe lifetime")
	dropTarget := flag.Bool("drop-target", false, "solid marker for an independent compositor drag")
	flag.Parse()
	a := app.NewWithID("org.fyne.native-input-probe")
	modern, err := theme.NewModern(theme.ModernOptions{})
	if err != nil {
		panic(err)
	}
	a.Settings().SetTheme(modern)
	w := a.NewWindow("Fyne native input acceptance")
	var mu sync.Mutex
	emit := func(data interface{}) { mu.Lock(); defer mu.Unlock(); _ = json.NewEncoder(os.Stdout).Encode(data) }
	e := widget.NewEntry()
	e.PlaceHolder = "CJK input"
	e.OnChanged = func(text string) { emit(map[string]interface{}{"event": "commit", "text": text}) }
	password := widget.NewPasswordEntry()
	password.PlaceHolder = "Private input"
	password.SetText("secret 日本 😀") // Nonempty fixture: redaction must not pass vacuously.
	list := widget.NewList(func() int { return 100000 }, func() granio.CanvasObject { return widget.NewLabel("") }, func(id widget.ListItemID, object granio.CanvasObject) {
		object.(*widget.Label).SetText(fmt.Sprintf("Item %d 日本", id))
	})
	list.SelectionMode = widget.SelectionMultiple
	list.SetSelection([]int{90000})
	button := widget.NewButton("Probe action", func() { emit(map[string]interface{}{"event": "action"}) })
	undo := widget.NewButton("Undo input", func() { e.Undo(); emit(map[string]interface{}{"event": "undo", "text": e.Text}) })
	w.SetContent(container.NewBorder(container.NewVBox(widget.NewLabel("Native IME and screen reader fixture"), e, password, button, undo), nil, nil, nil, list))
	if *dropTarget {
		w.SetContent(canvas.NewRectangle(color.NRGBA{R: 31, G: 80, B: 130, A: 255}))
	}
	w.Resize(granio.NewSize(600, 450))
	if drop, ok := w.(desktop.DataDropWindow); ok {
		_ = drop.SetOnDataDropped([]string{granio.MIMETextPlain, granio.MIMETextHTML, granio.MIMETextURIList, "application/x-fyne-drag-test"}, func(pos granio.Position, data map[string][]byte) {
			emit(map[string]interface{}{"event": "drop", "position": pos, "data": data})
		})
	}
	w.Show()
	if !*dropTarget {
		w.Canvas().Focus(e)
	}
	emit(map[string]interface{}{"event": "ready"})
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		until := time.NewTimer(time.Duration(*seconds) * time.Second)
		defer until.Stop()
		last := ""
		for {
			select {
			case <-ticker.C:
				granio.Do(func() {
					state := e.TextInputState()
					snapshot := fmt.Sprintf("%q/%t/%d/%d", e.Text, e.CompositionActive(), state.SelectionStart, state.SelectionEnd)
					if snapshot != last {
						last = snapshot
						emit(map[string]interface{}{"event": "state", "text": e.Text, "composition": e.CompositionActive(), "start": state.SelectionStart, "end": state.SelectionEnd})
					}
				})
			case <-until.C:
				granio.Do(func() { w.Close(); a.Quit() })
				return
			}
		}
	}()
	a.Run()
}
