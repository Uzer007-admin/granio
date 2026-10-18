//go:build !no_glfw && !mobile

package glfw

import (
	"io"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/lang"

	"github.com/stretchr/testify/assert"
)

func Test_Menu_Empty(t *testing.T) {
	// Discarding log output for tests
	// The following method logs an error:
	// bar := buildMenuOverlay(granio.NewMainMenu(), w.window)
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	w := createWindow("Menu Test")
	bar := buildMenuOverlay(granio.NewMainMenu(), w.window)
	assert.Nil(t, bar) // no bar but does not crash
}

func Test_Menu_AddsQuit(t *testing.T) {
	w := createWindow("Menu Test")
	mainMenu := granio.NewMainMenu(granio.NewMenu("File"))
	bar := buildMenuOverlay(mainMenu, w.window)
	assert.NotNil(t, bar)
	assert.Equal(t, 1, len(mainMenu.Items))
	assert.Equal(t, 2, len(mainMenu.Items[0].Items)) // separator+quit inserted
	assert.True(t, mainMenu.Items[0].Items[1].IsQuit)
}

func Test_Menu_LeaveQuit(t *testing.T) {
	w := createWindow("Menu Test")
	quitFunc := func() {}
	mainMenu := granio.NewMainMenu(granio.NewMenu("File", granio.NewMenuItem(lang.L("Quit"), quitFunc)))
	bar := buildMenuOverlay(mainMenu, w.window)
	assert.NotNil(t, bar)
	assert.Equal(t, 1, len(mainMenu.Items[0].Items)) // no separator added
	assert.Equal(t, reflect.ValueOf(quitFunc).Pointer(), reflect.ValueOf(mainMenu.Items[0].Items[0].Action).Pointer())
}

func Test_Menu_LeaveQuit_AddAction(t *testing.T) {
	w := createWindow("Menu Test")
	mainMenu := granio.NewMainMenu(granio.NewMenu("File", granio.NewMenuItem(lang.L("Quit"), nil)))
	bar := buildMenuOverlay(mainMenu, w.window)
	assert.NotNil(t, bar)
	assert.Equal(t, 1, len(mainMenu.Items[0].Items))    // no separator added
	assert.NotNil(t, mainMenu.Items[0].Items[0].Action) // quit action was added
}

func Test_Menu_CustomQuit(t *testing.T) {
	w := createWindow("Menu Test")

	quitFunc := func() {}
	quitItem := granio.NewMenuItem("Beenden", quitFunc)
	quitItem.IsQuit = true

	mainMenu := granio.NewMainMenu(granio.NewMenu("File", quitItem))
	bar := buildMenuOverlay(mainMenu, w.window)

	assert.NotNil(t, bar)
	assert.Equal(t, 1, len(mainMenu.Items[0].Items)) // no separator added
	assert.Equal(t, reflect.ValueOf(quitFunc).Pointer(), reflect.ValueOf(mainMenu.Items[0].Items[0].Action).Pointer())
}

func Test_Menu_CustomQuit_NoAction(t *testing.T) {
	w := createWindow("Menu Test")
	quitItem := granio.NewMenuItem("Beenden", nil)
	quitItem.IsQuit = true
	mainMenu := granio.NewMainMenu(granio.NewMenu("File", quitItem))
	bar := buildMenuOverlay(mainMenu, w.window)

	assert.NotNil(t, bar)
	assert.Equal(t, 1, len(mainMenu.Items[0].Items))    // no separator added
	assert.NotNil(t, mainMenu.Items[0].Items[0].Action) // quit action was added
}
