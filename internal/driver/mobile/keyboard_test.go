package mobile

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal/driver/mobile/event/key"
	"github.com/Uzer007-admin/granio/v2/internal/goos"
	"github.com/Uzer007-admin/granio/v2/widget"
)

// defaultModifier returns the modifier the desktop driver detects the built-in
// shortcuts with, i.e. granio.KeyModifierShortcutDefault (Command on macOS,
// Control everywhere else).
func defaultModifier() key.Modifiers {
	if runtime.GOOS == goos.Darwin {
		return key.ModMeta
	}
	return key.ModControl
}

// wordModifier returns the modifier widget.Entry registers its word selection
// and word deletion shortcuts with (Alt on macOS, Control everywhere else).
func wordModifier() key.Modifiers {
	if runtime.GOOS == goos.Darwin {
		return key.ModAlt
	}
	return key.ModControl
}

func TestShortcutForKey(t *testing.T) {
	assert.IsType(t, &granio.ShortcutSelectAll{}, shortcutForKey(granio.KeyA, granio.KeyModifierShortcutDefault))
	assert.IsType(t, &granio.ShortcutUndo{}, shortcutForKey(granio.KeyZ, granio.KeyModifierShortcutDefault))
	assert.IsType(t, &granio.ShortcutRedo{}, shortcutForKey(granio.KeyY, granio.KeyModifierShortcutDefault))
	assert.IsType(t, &granio.ShortcutCopy{}, shortcutForKey(granio.KeyC, granio.KeyModifierShortcutDefault))
	assert.IsType(t, &granio.ShortcutCut{}, shortcutForKey(granio.KeyX, granio.KeyModifierShortcutDefault))
	assert.IsType(t, &granio.ShortcutPaste{}, shortcutForKey(granio.KeyV, granio.KeyModifierShortcutDefault))
	assert.IsType(t, &granio.ShortcutCopy{}, shortcutForKey(granio.KeyInsert, granio.KeyModifierShortcutDefault))
	assert.IsType(t, &granio.ShortcutPaste{}, shortcutForKey(granio.KeyInsert, granio.KeyModifierShift))
	assert.IsType(t, &granio.ShortcutCut{}, shortcutForKey(granio.KeyDelete, granio.KeyModifierShift))

	// no modifier or shift-only modifiers must not create a shortcut
	assert.Nil(t, shortcutForKey(granio.KeyA, 0))
	assert.Nil(t, shortcutForKey(granio.KeyLeft, granio.KeyModifierShift))
	assert.Nil(t, shortcutForKey(desktop.KeyShiftLeft, granio.KeyModifierShift))

	// everything else becomes a custom shortcut, which widgets match by name
	bs := shortcutForKey(granio.KeyBackspace, granio.KeyModifierShortcutDefault)
	if assert.NotNil(t, bs) {
		want := &desktop.CustomShortcut{KeyName: granio.KeyBackspace, Modifier: granio.KeyModifierShortcutDefault}
		assert.Equal(t, want.ShortcutName(), bs.ShortcutName())
	}
	left := shortcutForKey(granio.KeyLeft, granio.KeyModifierShortcutDefault|granio.KeyModifierShift)
	if assert.NotNil(t, left) {
		want := &desktop.CustomShortcut{KeyName: granio.KeyLeft, Modifier: granio.KeyModifierShortcutDefault | granio.KeyModifierShift}
		assert.Equal(t, want.ShortcutName(), left.ShortcutName())
	}
}

func keyTestEntry(t *testing.T, text string) (*canvas, *widget.Entry) {
	t.Helper()
	w := d.CreateWindow("keyboard test")
	entry := widget.NewEntry()
	entry.SetText(text)
	w.SetContent(entry)
	w.Show()
	w.Resize(granio.NewSize(300, 100))
	c := w.Canvas().(*canvas)

	d.DoFromGoroutine(func() {
		c.Focus(entry)
	}, true)

	return c, entry
}

func TestMobileDriverKeyboardShortcuts(t *testing.T) {
	t.Run("default modifier + A selects all text", func(t *testing.T) {
		c, entry := keyTestEntry(t, "hello world")

		d.DoFromGoroutine(func() {
			d.typeDownCanvas(c, -1, key.CodeA, defaultModifier())
		}, true)

		assert.Equal(t, "hello world", entry.SelectedText())
	})

	t.Run("word modifier + Backspace deletes the preceding word", func(t *testing.T) {
		c, entry := keyTestEntry(t, "hello world")
		entry.CursorColumn = len(entry.Text) // place the cursor at the end

		d.DoFromGoroutine(func() {
			d.typeDownCanvas(c, -1, key.CodeDeleteBackspace, wordModifier())
		}, true)

		assert.Equal(t, "hello ", entry.Text)
	})

	t.Run("word modifier + shift + Left selects the preceding word", func(t *testing.T) {
		c, entry := keyTestEntry(t, "hello world foo")
		entry.CursorColumn = len(entry.Text)

		d.DoFromGoroutine(func() {
			d.typeDownCanvas(c, -1, key.CodeLeftArrow, wordModifier()|key.ModShift)
		}, true)

		assert.Equal(t, "foo", entry.SelectedText())
	})

	t.Run("Shift+Left selects by character while shift is held", func(t *testing.T) {
		c, entry := keyTestEntry(t, "hello")
		entry.CursorColumn = len(entry.Text)

		d.DoFromGoroutine(func() {
			d.typeDownCanvas(c, -1, key.CodeLeftShift, key.ModShift) // KeyDown(LeftShift) -> selectKeyDown
			d.typeDownCanvas(c, -1, key.CodeLeftArrow, key.ModShift)
		}, true)

		assert.Equal(t, "o", entry.SelectedText())
	})

	t.Run("releasing shift stops extending the selection", func(t *testing.T) {
		c, entry := keyTestEntry(t, "hello")
		entry.CursorColumn = len(entry.Text)

		d.DoFromGoroutine(func() {
			d.typeDownCanvas(c, -1, key.CodeLeftShift, key.ModShift)
			d.typeDownCanvas(c, -1, key.CodeLeftArrow, key.ModShift)
			d.typeUpCanvas(c, -1, key.CodeLeftShift, key.ModShift) // KeyUp(LeftShift)
			// the first Left collapses the active selection, the second moves the cursor
			d.typeDownCanvas(c, -1, key.CodeLeftArrow, 0)
			d.typeDownCanvas(c, -1, key.CodeLeftArrow, 0)
		}, true)

		assert.Equal(t, "", entry.SelectedText())
		assert.Equal(t, 3, entry.CursorColumn)
	})

	t.Run("plain typing still inserts the rune", func(t *testing.T) {
		c, entry := keyTestEntry(t, "he")
		entry.CursorColumn = len(entry.Text)

		d.DoFromGoroutine(func() {
			d.typeDownCanvas(c, 'l', key.CodeL, 0)
		}, true)

		assert.Equal(t, "hel", entry.Text)
	})
}

func TestKeyModifiers(t *testing.T) {
	assert.Equal(t, granio.KeyModifierShift, keyModifiers(key.ModShift))
	assert.Equal(t, granio.KeyModifierControl, keyModifiers(key.ModControl))
	assert.Equal(t, granio.KeyModifierAlt, keyModifiers(key.ModAlt))
	assert.Equal(t, granio.KeyModifierSuper, keyModifiers(key.ModMeta))
	assert.Equal(t, granio.KeyModifier(0), keyModifiers(0))
	assert.Equal(t, granio.KeyModifierControl|granio.KeyModifierShift, keyModifiers(key.ModControl|key.ModShift))
}
