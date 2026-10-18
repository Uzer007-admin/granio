package widget

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/widget"
	"github.com/Uzer007-admin/granio/v2/layout"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/stretchr/testify/require"
)

func TestModernPopupEscapeRestoresEntryFocusAndSelection(t *testing.T) {
	a := newVisualTestApp(t)
	useModern(t, a, theme.ModernOptions{DisableMotion: true})
	e := NewEntry()
	e.SetText("Русский English")
	window := a.NewWindow("Popup focus")
	window.SetContent(granio.NewContainerWithLayout(layout.NewVBoxLayout(), e))
	window.Resize(granio.NewSize(500, 350))
	c := window.Canvas()
	c.Focus(e)
	e.AccessibilitySetSelection(0, 7)
	popup := NewPopUp(NewButton("Действие", nil), c)
	popup.Show()
	require.False(t, e.focused)
	require.True(t, c.Overlays().Top().(*widget.OverlayContainer).HandleEscape())
	require.Nil(t, c.Overlays().Top())
	require.Same(t, e, c.Focused())
	require.True(t, e.focused)
	_, start, end := e.AccessibilityTextSelection()
	require.Equal(t, 0, start)
	require.Equal(t, 7, end)
	menu := NewPopUpMenu(granio.NewMenu("", granio.NewMenuItem("Команда", func() {})), c)
	menu.Show()
	menu.TypedKey(&granio.KeyEvent{Name: granio.KeyEscape})
	require.Same(t, e, c.Focused())
	require.True(t, e.focused)
	modal := NewModalPopUp(NewLabel("Модальная панель"), c)
	modal.Show()
	require.False(t, modal.HandleEscape())
	closed := 0
	modal.OnEscape = func() { closed++; modal.Hide() }
	require.True(t, modal.HandleEscape())
	require.Equal(t, 1, closed)
	require.Same(t, e, c.Focused())
}
