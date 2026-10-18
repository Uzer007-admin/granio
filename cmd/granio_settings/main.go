package main

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
	"github.com/Uzer007-admin/granio/v2/cmd/granio_settings/settings"
	"github.com/Uzer007-admin/granio/v2/container"
)

func main() {
	s := settings.NewSettings()

	a := app.New()
	w := a.NewWindow("Fyne Settings")

	appearance := s.LoadAppearanceScreen(w)
	tabs := container.NewAppTabs(
		&container.TabItem{Text: "Appearance", Icon: s.AppearanceIcon(), Content: appearance},
	)
	tabs.SetTabLocation(container.TabLocationLeading)
	w.SetContent(tabs)

	w.Resize(granio.NewSize(520, 520)) //revive:disable-line:add-constant
	w.ShowAndRun()
}
