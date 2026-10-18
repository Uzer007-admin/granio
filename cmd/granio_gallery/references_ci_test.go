//go:build ci

package main

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/app"
)

// Real app settings defer cache invalidation to the event loop. The exporter
// runs without that loop and must also update already cached icon colours.
func TestReferenceExporterIconsWithAppSettings(t *testing.T) {
	previous := granio.CurrentApp()
	a := app.NewWithID("io.fyne.reference-icons-test")
	t.Cleanup(func() { a.Quit(); granio.SetCurrentApp(previous) })
	checkReferenceSceneIcons(t, a)
}
