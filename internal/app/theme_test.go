package app_test

import (
	"testing"

	"github.com/Uzer007-admin/granio/v3/internal/app"
	"github.com/Uzer007-admin/granio/v3/test"
)

func TestApplySettings_BeforeContentSet(*testing.T) {
	a := test.NewApp()
	w := a.NewWindow("NoContent")
	defer w.Close()

	app.ApplySettings(a.Settings(), a)
}
