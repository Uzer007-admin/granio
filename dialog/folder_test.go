package dialog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	intWidget "github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/lang"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/storage"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/widget"
)

func TestShowFolderOpen(t *testing.T) {
	var chosen granio.ListableURI
	var openErr error
	win := test.NewTempWindow(t, widget.NewLabel("OpenDir"))
	d := NewFolderOpen(func(file granio.ListableURI, err error) {
		chosen = file
		openErr = err
	}, win)
	dir, err := storage.ListerForURI(storage.NewFileURI("testdata"))
	if err != nil {
		t.Error("Failed to open testdata dir", err)
	}
	d.SetLocation(dir)
	d.Show()

	popup := win.Canvas().Overlays().Top().(*intWidget.OverlayContainer).Content.(*widget.PopUp)
	defer win.Canvas().Overlays().Remove(popup)
	assert.NotNil(t, popup)

	// padded container -> content container
	ui := popup.Content.(*granio.Container).Objects[0].(*granio.Container)
	title := ui.Objects[1].(*granio.Container).Objects[0].(*widget.Label)
	assert.Equal(t, lang.L("Open")+" "+lang.L("Folder"), title.Text)

	nameLabel := ui.Objects[2].(*granio.Container).Objects[0].(*container.Scroll).Content.(*widget.Label)
	buttons := ui.Objects[2].(*granio.Container).Objects[1].(*granio.Container)
	open := buttons.Objects[1].(*widget.Button)

	files := ui.Objects[0].(*container.Split).Trailing.(*granio.Container).Objects[0].(*container.Scroll).Content.(*granio.Container).Objects[0].(*widget.GridWrap)
	assert.NotEmpty(t, d.dialog.data)

	item := test.TempWidgetRenderer(t, files).Objects()[0].(*container.Scroll).Content.(*granio.Container).Objects[0]
	fileName := test.TempWidgetRenderer(t, item.(granio.Widget)).Objects()[1].(*fileDialogItem).name
	assert.Equal(t, lang.L("(Parent)"), fileName)
	assert.False(t, open.Disabled())

	var target granio.URI
	id := -1
	for i, uri := range d.dialog.data {
		ok, _ := storage.CanList(uri)
		if ok {
			target = uri
			id = i
		} else {
			t.Error("Folder dialog should not list files")
		}
	}

	require.NotNil(t, target, "Failed to find folder in testdata")
	d.dialog.files.(*widget.GridWrap).Select(id)
	assert.Equal(t, target.Name(), nameLabel.Text)
	assert.False(t, open.Disabled())

	test.Tap(open)
	assert.Nil(t, win.Canvas().Overlays().Top())
	assert.NoError(t, openErr)

	assert.Equal(t, target.String(), chosen.String())
}
