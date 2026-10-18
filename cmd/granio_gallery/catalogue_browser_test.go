package main

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/require"
)

func TestCatalogueSearchHistoryAndCachedPages(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	w := a.NewWindow("Каталог")
	defer w.Close()
	b := newGalleryCatalogue(w, "data-grid")
	w.SetContent(b.root)
	w.Resize(granio.NewSize(1060, 760))
	require.Equal(t, "data-grid", b.samples[b.current].id)
	require.Equal(t, []widget.ListItemID{1}, b.list.Selected())
	page := b.pages[b.current]
	b.search.SetText("таблица сортировка")
	require.Len(t, b.filtered, 1)
	require.Same(t, page, b.pages[b.current])
	b.search.SetText("несуществующий пример")
	require.Empty(t, b.filtered)
	require.Empty(t, b.code.Text)
	b.search.SetText("Wayland")
	require.Equal(t, "system", b.samples[b.current].id)
	b.move(-1)
	require.Equal(t, "data-grid", b.samples[b.current].id)
	require.Empty(t, b.search.Text)
	require.Same(t, page, b.pages[b.current])
	b.move(1)
	require.Equal(t, "system", b.samples[b.current].id)
	require.NotEmpty(t, b.code.Text)
	for i := 0; i < 150; i++ {
		b.navigate(i%2, true)
	}
	require.LessOrEqual(t, len(b.history), 128)
}
