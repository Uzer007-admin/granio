package mobile

import (
	"io"
	"log"
	"testing"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/stretchr/testify/assert"
)

func Test_device(t *testing.T) {
	dev := &device{}

	assert.Equal(t, true, dev.IsMobile())
	assert.Equal(t, false, dev.IsBrowser())
	assert.Equal(t, false, dev.HasKeyboard())
}

func TestDeviceVirtualKeyboardWithoutApp(t *testing.T) {
	previousApp, previousLog := granio.CurrentApp(), log.Writer()
	t.Cleanup(func() {
		granio.SetCurrentApp(previousApp)
		log.SetOutput(previousLog)
	})
	granio.SetCurrentApp(nil)
	log.SetOutput(io.Discard)

	dev := &device{}
	assert.NotPanics(t, dev.ShowVirtualKeyboard)
	assert.NotPanics(t, dev.HideVirtualKeyboard)
}
