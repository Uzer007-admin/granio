package widget

import (
	"errors"
	"fmt"
	"image/color"
	"strings"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/data/binding"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/Uzer007-admin/granio/v2/internal/painter/software"
	"github.com/Uzer007-admin/granio/v2/layout"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/stretchr/testify/require"
)

func useModern(t *testing.T, a granio.App, options theme.ModernOptions) {
	t.Helper()
	m, err := theme.NewModern(options)
	require.NoError(t, err)
	a.Settings().SetTheme(m)
}

func newVisualTestApp(t *testing.T) granio.App {
	t.Helper()
	previous := granio.CurrentApp()
	a := test.NewApp()
	t.Cleanup(func() { a.Quit(); granio.SetCurrentApp(previous) })
	return a
}

func TestReadOnlyRejectsAllUserEditingAndKeepsBindingSelection(t *testing.T) {
	a := newVisualTestApp(t)
	data := binding.NewString()
	require.NoError(t, data.Set("Русский English 123"))
	e := NewEntryWithData(data)
	waitForBinding()
	e.Resize(granio.NewSize(420, 70))
	require.True(t, e.AccessibilitySetSelection(0, 7))
	e.UpdateComposition(desktop.CompositionEvent{Text: "候補", Cursor: 2})
	e.SetReadOnly(true)
	require.False(t, e.CompositionActive())
	before := e.Text
	_, start, end := e.AccessibilityTextSelection()
	changes := 0
	e.OnChanged = func(string) { changes++ }
	e.TypedRune('x')
	e.TypedText("замена")
	for _, key := range []granio.KeyName{granio.KeyBackspace, granio.KeyDelete, granio.KeyReturn, granio.KeyTab} {
		e.TypedKey(&granio.KeyEvent{Name: key})
	}
	a.Clipboard().SetContent("clipboard")
	e.TypedShortcut(&granio.ShortcutPaste{Clipboard: a.Clipboard()})
	e.TypedShortcut(&granio.ShortcutCut{Clipboard: a.Clipboard()})
	e.Undo()
	e.Redo()
	e.UpdateComposition(desktop.CompositionEvent{Text: "候補"})
	e.CommitComposition("確定")
	e.CommitTextEdit("native", 3, 2)
	require.False(t, e.ReplaceTextRange("native", 0, 1))
	require.False(t, e.StartCompositionRange(0, 1))
	require.False(t, e.AccessibilitySetText("automation"))
	require.Equal(t, before, e.Text)
	require.Zero(t, changes)
	_, afterStart, afterEnd := e.AccessibilityTextSelection()
	require.Equal(t, start, afterStart)
	require.Equal(t, end, afterEnd)
	e.TypedShortcut(&granio.ShortcutCopy{Clipboard: a.Clipboard()})
	require.Equal(t, "Русский", a.Clipboard().Content())
	require.True(t, e.AccessibilityState().ReadOnly)
	require.True(t, e.AccessibilitySetSelection(8, 15))
	e.SetText("programmatic")
	require.Equal(t, "programmatic", e.Text)
	e.SetReadOnly(false)
	e.CursorColumn = len([]rune(e.Text))
	e.TypedRune('!')
	require.Contains(t, e.Text, "!")

	rich := NewRichTextEntryFromMarkdown("**bold** Русский")
	rich.Resize(granio.NewSize(400, 100))
	rich.SetReadOnly(true)
	text := rich.Text
	rich.TypedRune('!')
	rich.TypedText("edit")
	rich.CommitTextEdit("edit", 0, 2)
	require.False(t, rich.ReplaceTextRange("edit", 0, 2))
	rich.Undo()
	rich.Redo()
	require.Equal(t, text, rich.Text)
}

func TestSemanticTypographySharesDrawMeasureAndHitTest(t *testing.T) {
	a := newVisualTestApp(t)
	for _, fontPolicy := range []theme.FontPolicy{theme.FontBundled, theme.FontSystem} {
		useModern(t, a, theme.ModernOptions{TextScale: 2.25, FontPolicy: fontPolicy})
		for _, role := range []granio.TextRole{granio.TextRoleBody, granio.TextRoleCaption, granio.TextRoleTitle, granio.TextRoleCode} {
			label := NewLabel("AV Яя 012345 /данные/каталог.txt")
			label.TextRole, label.Selectable = role, true
			label.Resize(granio.NewSize(1800, 140))
			label.Refresh()
			segment := label.provider.Segments[0].(*TextSegment)
			text := segment.Visual().(*canvas.Text)
			expected, _ := a.Driver().RenderedTextSize(label.Text, text.TextSize, text.TextStyle, text.FontSource)
			require.Equal(t, expected, segment.measure(label.Text))
			pad := label.Theme().Size(theme.SizeNameInnerPadding)
			runes := []rune(label.Text)
			for i := 0; i <= len(runes); i++ {
				x := segment.measure(string(runes[:i])).Width
				require.InDelta(t, x+pad, label.provider.lineSizeToColumn(i, 0, text.TextSize, pad).Width, .001)
				require.Equal(t, i, label.selection.cursorColAt(0, granio.NewPos(x+pad, 0)))
			}
		}
	}
	text := NewRichText(&TextSegment{Text: "Русский\nEnglish\n123", Style: RichTextStyle{TextRole: granio.TextRoleBody}})
	text.Resize(granio.NewSize(800, 300))
	text.Refresh()
	text.ensureRowGeometry()
	y, height := text.rowGeometry(0)
	next, _ := text.rowGeometry(1)
	require.EqualValues(t, 0, y)
	require.InDelta(t, height+9, next, .001, "semantic spacing must reach caret geometry")
	legacy := &TextSegment{Text: "legacy", Style: RichTextStyle{SizeName: theme.SizeNameCaptionText, TextStyle: granio.TextStyle{Bold: true}}}
	require.Nil(t, legacy.typography().Font, "legacy explicit font resolution must remain unchanged")
}

func TestModernFocusErrorReadOnlyContoursAndGeometry(t *testing.T) {
	a := newVisualTestApp(t)
	useModern(t, a, theme.ModernOptions{DisableMotion: true})
	e := NewEntry()
	e.SetText("Данные")
	e.Validator = func(string) error { return errors.New("Ошибка проверки") }
	e.AlwaysShowValidationError = true
	e.Resize(granio.NewSize(320, 70))
	r := cache.Renderer(e).(*entryRenderer)
	r.Refresh()
	min, textPos := e.MinSize(), e.content.Position()
	e.SetReadOnly(true)
	e.FocusGained()
	r.Refresh()
	require.True(t, r.focusBorder.Visible())
	require.Equal(t, e.Theme().Color(theme.ColorNameError, a.Settings().ThemeVariant()), r.border.StrokeColor)
	require.Equal(t, granio.NewPos(4, 4), r.border.Position())
	require.Equal(t, granio.NewPos(1, 1), r.focusBorder.Position())
	require.Equal(t, min, e.MinSize())
	require.Equal(t, textPos, e.content.Position())
	e.Disable()
	r.Refresh()
	require.False(t, r.focusBorder.Visible())
	e.Enable()
	b := NewButton("Кнопка", nil)
	b.Resize(granio.NewSize(200, 50))
	bmin := b.MinSize()
	for _, event := range []func(){func() { b.MouseIn(&desktop.MouseEvent{}) }, func() { b.MouseDown(&desktop.MouseEvent{Button: desktop.MouseButtonPrimary}) }, b.FocusGained, b.MouseOut, b.Disable, b.Enable} {
		event()
		require.Equal(t, bmin, b.MinSize())
	}
	icon := NewButtonWithIcon("", theme.SearchIcon(), nil)
	icon.AccessibleName = "Поиск"
	require.Equal(t, "Поиск", icon.AccessibilityLabel())
	require.GreaterOrEqual(t, icon.MinSize().Height, float32(40))

	busy := NewProgressBarInfinite()
	busy.Resize(granio.NewSize(200, 40))
	busy.Start()
	view := cache.Renderer(busy).(*infProgressRenderer)
	view.Refresh()
	require.True(t, busy.Running())
	require.False(t, view.wasRunning)
	require.Greater(t, view.bar.Size().Width, float32(0))
}

func TestModernFormWrapsLabelsAndErrorsWithoutReplacingInput(t *testing.T) {
	a := newVisualTestApp(t)
	useModern(t, a, theme.ModernOptions{TextScale: 2.25, DisableMotion: true})
	e := NewEntry()
	e.SetText("value")
	e.dirty = true
	e.Validator = func(string) error {
		return errors.New(strings.Repeat("Подробная ошибка проверки введённых данных. ", 4))
	}
	item := NewFormItem(strings.Repeat("Длинная русская подпись ", 3), e)
	f := NewForm(item, NewFormItem("Следующее поле", NewEntry()))
	f.Resize(granio.NewSize(620, 900))
	f.Refresh()
	f.itemGrid.Resize(granio.NewSize(620, 850))
	require.NotNil(t, item.helperWrapped)
	require.Greater(t, item.helperWrapped.MinSize().Height, e.Theme().Size(theme.SizeNameCaptionText)*2)
	label := f.itemGrid.Objects[0].(*RichText)
	require.Greater(t, label.MinSize().Height, e.Theme().Size(theme.SizeNameText)*2)
	first, second := f.itemGrid.Objects[1], f.itemGrid.Objects[3]
	require.GreaterOrEqual(t, second.Position().Y, first.Position().Y+first.Size().Height)
	require.Same(t, e, f.unwrapItemWidget(first))
	a.Settings().SetTheme(theme.DefaultTheme())
	f.Refresh()
	require.Same(t, e, f.unwrapItemWidget(f.itemGrid.Objects[1]))
	require.Equal(t, "value", e.Text)
}

func TestModernDPIAndTextScaleStockMatrix(t *testing.T) {
	a := newVisualTestApp(t)
	for _, dpi := range []float32{1, 1.5, 2} {
		for _, textScale := range []float32{1, 1.5, 2, 2.25} {
			t.Run(fmt.Sprintf("dpi-%g-text-%g", dpi, textScale), func(t *testing.T) {
				useModern(t, a, theme.ModernOptions{TextScale: textScale, DisableMotion: true})
				entry := NewEntry()
				entry.SetText("Русский English 012345")
				entry.AccessibilitySetSelection(0, 7)
				button := NewButton("Сохранить", nil)
				check := NewCheck("Состояние", nil)
				selectBox := NewSelect([]string{"Первый", "Второй"}, nil)
				selectBox.SetSelected("Второй")
				radio := newRadioItem("Вариант", nil)
				content := granio.NewContainerWithLayout(layout.NewVBoxLayout(), entry, button, check, selectBox, radio, NewSlider(0, 100), NewProgressBar(), NewLabel("/очень/длинный/путь/"))
				c := test.NewCanvasWithPainter(software.NewPainter())
				c.SetScale(dpi)
				c.SetPadded(false)
				c.SetContent(content)
				c.Resize(granio.NewSize(700, 600))
				image := canvas.CaptureOpaque(c, color.White)
				require.EqualValues(t, 700*dpi, image.Bounds().Dx())
				require.EqualValues(t, 600*dpi, image.Bounds().Dy())
				for _, object := range []granio.CanvasObject{entry, button, check, selectBox, radio} {
					require.GreaterOrEqual(t, object.MinSize().Height, float32(40))
					require.GreaterOrEqual(t, object.Size().Height, object.MinSize().Height)
				}
				require.EqualValues(t, 14*textScale, a.Settings().Theme().Size(theme.SizeNameText))
				require.EqualValues(t, 20, a.Settings().Theme().Size(theme.SizeNameInlineIcon))
				_, start, end := entry.AccessibilityTextSelection()
				require.Equal(t, 0, start)
				require.Equal(t, 7, end)
				require.Equal(t, "Второй", selectBox.Selected)
			})
		}
	}
}

func TestModernLiveProfilesPreserveEditorAndVirtualCollections(t *testing.T) {
	a := newVisualTestApp(t)
	useModern(t, a, theme.ModernOptions{DisableMotion: true})
	data := binding.NewString()
	require.NoError(t, data.Set(strings.Repeat("Русский English 012345\n", 30)))
	e := NewMultiLineEntry()
	e.Bind(data)
	waitForBinding()
	e.Resize(granio.NewSize(450, 180))
	window := a.NewWindow("Live profiles")
	window.SetContent(e)
	window.Resize(granio.NewSize(450, 180))
	window.Canvas().Focus(e)
	e.AccessibilitySetSelection(2, 18)
	e.scroller.ScrollToOffset(granio.NewPos(0, 80))
	offset := e.scroller.Offset
	listCreated, tableCreated, treeCreated := 0, 0, 0
	l := NewList(func() int { return 10000 }, func() granio.CanvasObject { listCreated++; return NewLabel("") }, func(id int, obj granio.CanvasObject) { obj.(*Label).SetText(fmt.Sprint(id)) })
	table := NewTable(func() (int, int) { return 10000, 3 }, func() granio.CanvasObject { tableCreated++; return NewLabel("") }, func(id TableCellID, obj granio.CanvasObject) { obj.(*Label).SetText(fmt.Sprint(id.Row)) })
	children := make([]string, 10000)
	for i := range children {
		children[i] = fmt.Sprint(i)
	}
	tree := NewTree(func(id string) []string {
		if id == "" {
			return children
		}
		return nil
	}, func(id string) bool { return id == "" }, func(bool) granio.CanvasObject { treeCreated++; return NewLabel("") }, func(id string, _ bool, obj granio.CanvasObject) { obj.(*Label).SetText(id) })
	for _, obj := range []granio.CanvasObject{l, table, tree} {
		obj.Resize(granio.NewSize(300, 200))
	}
	l.Select(7)
	table.Select(TableCellID{Row: 8, Col: 1})
	tree.Select("9")
	text := e.Text
	for _, scale := range []float32{1.5, 2, 2.25, 1} {
		for _, density := range []theme.Density{theme.DensityCompact, theme.DensityComfortable} {
			useModern(t, a, theme.ModernOptions{TextScale: scale, Density: density, HighContrast: density == theme.DensityCompact, DisableMotion: true})
			e.Refresh()
			l.Refresh()
			table.Refresh()
			tree.Refresh()
			require.Equal(t, text, e.Text)
			_, start, end := e.AccessibilityTextSelection()
			require.Equal(t, 2, start)
			require.Equal(t, 18, end)
			require.Same(t, e, window.Canvas().Focused())
			require.Equal(t, offset, e.scroller.Offset)
			require.Equal(t, []int{7}, l.Selected())
			require.Equal(t, []TableCellID{{Row: 8, Col: 1}}, table.Selected())
			require.Equal(t, []string{"9"}, tree.Selected())
		}
	}
	require.Less(t, listCreated, 120)
	require.Less(t, tableCreated, 240)
	require.Less(t, treeCreated, 120)
	t.Logf("live profiles kept viewport %v and selection; created list/table/tree children=%d/%d/%d for 10000 rows", offset, listCreated, tableCreated, treeCreated)
	require.NoError(t, data.Set("Новые данные binding"))
	waitForBinding()
	require.Equal(t, "Новые данные binding", e.Text)
}
