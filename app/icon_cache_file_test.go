package app

import (
	_ "embed"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v2"
)

//go:embed testdata/fyne.png
var iconData []byte

func TestCachedIcon_PATH(t *testing.T) {
	SetMetadata(granio.AppMetadata{})
	a := &fyneApp{uniqueID: "icontest"}
	assert.Equal(t, "", a.cachedIconPath())

	a.SetIcon(granio.NewStaticResource("dummy", iconData))
	path := a.cachedIconPath()
	if path == "" {
		t.Error("cache path not constructed")
		return
	}

	defer os.Remove(path)

	info, err := os.Stat(path)
	assert.NoError(t, err)
	assert.Equal(t, "icon.png", info.Name())

	assert.NoError(t, err)
}
