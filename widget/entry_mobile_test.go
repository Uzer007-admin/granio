//go:build mobile

package widget_test

import (
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/mobile"
	"github.com/Uzer007-admin/granio/v2/test"

	"github.com/stretchr/testify/assert"
)

func TestEntry_Select_TripleTap_Mobile(t *testing.T) {
	e, _ := setupSelection(t, false)
	e.MultiLine = true

	// double tap selects a word and records the timestamp used for triple detection
	test.DoubleTap(e)
	time.Sleep(20 * time.Millisecond)

	// the third touch triggers the triple-tap select of the whole row
	pos := granio.NewPos(1, 1)
	e.TouchDown(&mobile.TouchEvent{PointEvent: granio.PointEvent{Position: pos}})
	e.TouchUp(&mobile.TouchEvent{PointEvent: granio.PointEvent{Position: pos}})
	assert.Equal(t, "Testing", e.SelectedText())
}

func TestEntry_Select_TapAfterDoubleTap_Mobile(t *testing.T) {
	e, _ := setupSelection(t, false)
	e.MultiLine = true

	// a double tap selects a word - the driver sends DoubleTapped instead of Tapped
	test.DoubleTap(e)
	assert.NotEmpty(t, e.SelectedText())

	// a later single tap should just move the cursor, not keep the selection alive
	time.Sleep(granio.CurrentApp().Driver().DoubleTapDelay() + 50*time.Millisecond)
	pos := granio.NewPos(1, 1)
	e.TouchDown(&mobile.TouchEvent{PointEvent: granio.PointEvent{Position: pos}})
	e.TouchUp(&mobile.TouchEvent{PointEvent: granio.PointEvent{Position: pos}})
	test.Tap(e)
	assert.Equal(t, "", e.SelectedText())
}
