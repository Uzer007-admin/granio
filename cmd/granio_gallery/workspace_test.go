package main

import (
	"strings"
	"testing"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/container"
	"github.com/Uzer007-admin/granio/v3/driver/desktop"
	"github.com/Uzer007-admin/granio/v3/test"
	"github.com/Uzer007-admin/granio/v3/theme"
	"github.com/stretchr/testify/require"
)

func newWorkspaceTest(t *testing.T) *workspaceView {
	t.Helper()
	a := test.NewTempApp(t)
	base, err := theme.NewModern(theme.ModernOptions{Density: theme.DensityCompact})
	require.NoError(t, err)
	palette := workspaceTheme{Theme: base}
	a.Settings().SetTheme(palette)
	w := a.NewWindow("Workspace test")
	t.Cleanup(w.Close)
	v := newWorkspace(w, palette, nil)
	w.SetPadded(false)
	w.SetContent(v.root)
	w.Resize(granio.NewSize(1240, 848))
	return v
}

func TestWorkspacePanelButtonsAndTabs(t *testing.T) {
	v := newWorkspaceTest(t)

	require.True(t, v.documentPane.PanelOpen())
	require.True(t, v.sidebarPane.PanelOpen())
	test.Tap(v.toolbarPane)
	require.False(t, v.documentPane.PanelOpen())
	test.Tap(v.leftToggle)
	require.False(t, v.sidebarPane.PanelOpen())
	test.Tap(v.leftToggle)
	test.Tap(v.toolbarPane)
	require.True(t, v.sidebarPane.PanelOpen())
	require.True(t, v.documentPane.PanelOpen())

	test.Tap(v.docTabs[1])
	require.Equal(t, "ОТЧЁТ.md", v.docName.Text)
	require.True(t, v.docTabs[1].Selected)
	require.False(t, v.docTabs[0].Selected)
	test.Tap(v.newDialog)
	require.Equal(t, "Диалог 1", v.title.Text)
	before := len(v.feed.objects)
	v.entry.SetText("Контрольная заметка")
	v.addNote()
	require.Len(t, v.feed.objects, before+1)
	require.Empty(t, v.entry.Text)
}

func TestWorkspaceDocumentRowsFitViewportOnFirstResize(t *testing.T) {
	v := newWorkspaceTest(t)
	doc := v.docScroll.Content.(*workspaceFlow)
	v.w.Resize(granio.NewSize(825, 540))
	v.docScroll.ScrollToBottom()
	require.Positive(t, v.docScroll.Offset.Y)
	for _, width := range []float32{825, 1240, 920, 825} {
		v.w.Resize(granio.NewSize(width, 540))
		viewport := v.docScroll.Size()
		require.Equal(t, viewport.Width, doc.Size().Width)
		require.Equal(t, max(viewport.Height, doc.MinSize().Height), doc.Size().Height)
		last := float32(0)
		for _, object := range doc.objects {
			require.GreaterOrEqual(t, object.Position().Y, last)
			last = object.Position().Y + object.Size().Height
			if row, ok := object.(*workspaceBlock); ok {
				left := row.objects[0].(*container.Clip)
				right := row.objects[1].(*container.Clip)
				if row.Size().Width < 260 {
					require.GreaterOrEqual(t, right.Position().Y, left.Position().Y+left.Size().Height)
				} else {
					require.GreaterOrEqual(t, right.Position().X, left.Position().X+left.Size().Width)
				}
				for _, cell := range []*container.Clip{left, right} {
					require.GreaterOrEqual(t, cell.Size().Height, cell.Content.MinSize().Height)
					require.LessOrEqual(t, cell.Position().Y+cell.Size().Height, row.Size().Height)
					require.LessOrEqual(t, cell.Position().X+cell.Size().Width, row.Size().Width)
				}
			}
		}
		require.LessOrEqual(t, last, doc.Size().Height)
		require.LessOrEqual(t, v.docScroll.Offset.Y, doc.Size().Height-viewport.Height)
	}
}

func TestWorkspaceComposerGrowsAndSendsWithinBudget(t *testing.T) {
	v := newWorkspaceTest(t)
	minimum := v.entry.Size().Height
	v.entry.SetText(strings.Repeat("Перенос строки по ширине поля. ", 17))
	require.Greater(t, v.entry.Size().Height, minimum)
	v.entry.SetText(strings.Repeat("Отдельная строка\n", 45))
	require.Equal(t, float32(workspaceComposeMax-workspaceComposeTools), v.entry.Size().Height)
	require.Greater(t, v.entry.TextContentHeight(), v.entry.Size().Height)
	v.entry.SetText("")
	require.Equal(t, minimum, v.entry.Size().Height)

	test.Type(v.entry, "Привет")
	v.entry.KeyDown(&granio.KeyEvent{Name: desktop.KeyShiftLeft})
	v.entry.TypedKey(&granio.KeyEvent{Name: granio.KeyReturn})
	v.entry.KeyUp(&granio.KeyEvent{Name: desktop.KeyShiftLeft})
	test.Type(v.entry, "😀")
	require.Equal(t, "Привет\n😀", v.entry.Text)
	require.Equal(t, 8, v.usage.used)
	v.entry.TypedKey(&granio.KeyEvent{Name: granio.KeyReturn})
	require.Equal(t, []string{"Привет\n😀"}, v.threads[0].messages)
	require.Empty(t, v.entry.Text)
	require.Equal(t, 8, v.usage.used)

	v.entry.SetText("Незаконченный черновик")
	v.newConversation()
	require.Empty(t, v.entry.Text)
	require.Zero(t, v.usage.used)
	v.selectProject(0)
	require.Equal(t, "Незаконченный черновик", v.entry.Text)
	require.Len(t, v.feed.objects, 1)

	v.threads[0].used = workspaceChatLimit - 3
	v.entry.SetText("АБ😀Г")
	require.True(t, v.send.Disabled())
	v.entry.TypedKey(&granio.KeyEvent{Name: granio.KeyReturn})
	require.Len(t, v.threads[0].messages, 1)
	require.Equal(t, "АБ😀Г", v.entry.Text)
	v.entry.SetText("АБ😀")
	require.False(t, v.send.Disabled())
	test.Tap(v.send)
	require.Equal(t, workspaceChatLimit, v.threads[0].used)
	require.Len(t, v.threads[0].messages, 2)
}

func TestWorkspaceStartsEmptyAndOnlyPersistsAppearance(t *testing.T) {
	v := newWorkspaceTest(t)
	require.Equal(t, -1, v.project)
	require.Empty(t, v.threads)
	require.Empty(t, v.feed.objects)
	require.Empty(t, v.entry.Text)
	require.Zero(t, v.usage.used)
	require.True(t, v.send.Disabled())
	v.entry.SetText("Геометрический образец")
	v.addNote()
	require.Equal(t, "Диалог 1", v.title.Text)
	v.entry.SetText("Временный черновик")
	appearance := v.theme.appearance
	appearance.Groups[workspaceEditor] = workspaceOptics{100, 50, [4]int{0, 19, 32, 48}, false}
	v.saveAppearance()
	prefs := granio.CurrentApp().Preferences()
	saved := prefs.String(workspaceAppearanceKey)
	require.NotContains(t, saved, "Геометрический образец")
	require.NotContains(t, saved, "Временный черновик")
	require.Equal(t, *appearance, loadWorkspaceAppearance(prefs))
	require.Equal(t, uint8(0), v.composerGlass.Tint.A)
	require.Equal(t, float32(50), v.composerGlass.BlurRadius)
	require.Zero(t, v.composerGlass.EffectiveCornerRadii().TopLeft)
}

func TestWorkspaceAppearanceDialogLeavesComposerIntact(t *testing.T) {
	v := newWorkspaceTest(t)
	v.entry.SetText("Проверка переноса текста и фокуса")
	v.showAppearance()
	require.NotNil(t, v.appearanceUI)
	require.NotNil(t, v.w.Canvas().Overlays().Top())
	require.False(t, v.composerGlass.SoftwareBlur, "the visible settings preview has software blur priority")
	v.theme.appearance.Groups[workspaceControls] = workspaceOptics{30, 8, [4]int{0, 12, 7, 48}, false}
	v.theme.appearance.Motion = 2
	v.saveAppearance()
	require.Equal(t, "Проверка переноса текста и фокуса", v.entry.Text)
	require.True(t, theme.Motion(v.theme, granio.MotionNavigation).Disabled)
	require.Equal(t, &granio.CornerRadii{TopRight: 12, BottomRight: 7, BottomLeft: 48}, theme.Corners(v.theme, granio.ComponentButton, theme.VariantDark))
}

func TestWorkspaceComposerFloatsOverFeed(t *testing.T) {
	v := newWorkspaceTest(t)
	v.entry.SetText(strings.Repeat("Текст сообщения под плавающей панелью. ", 30))
	v.addNote()
	v.entry.SetText(strings.Repeat("Высота редактора при переносе строки. ", 20))
	widgetDriver := granio.CurrentApp().Driver()
	field := widgetDriver.AbsolutePositionForObject(v.entry)
	feed := widgetDriver.AbsolutePositionForObject(v.feedScroll)
	status := widgetDriver.AbsolutePositionForObject(v.usage.button)
	require.Less(t, field.Y+v.entry.Size().Height, feed.Y+v.feedScroll.Size().Height)
	require.Less(t, status.Y+v.usage.button.Size().Height, field.Y)
	require.Equal(t, v.conversationView.Size().Height-workspaceComposeBottom, v.feedScroll.Position().Y+v.feedScroll.Size().Height)
	v.feedScroll.ScrollToBottom()
	last := v.feed.objects[len(v.feed.objects)-1]
	lastPosition := widgetDriver.AbsolutePositionForObject(last)
	require.LessOrEqual(t, lastPosition.Y+last.Size().Height, status.Y)
	test.Tap(v.usage.button)
	require.NotNil(t, v.w.Canvas().Overlays().Top())
}

func TestWorkspaceDragsDocumentTabAndMovesPanel(t *testing.T) {
	v := newWorkspaceTest(t)
	tab := v.docTabs[2]
	pos := granio.CurrentApp().Driver().AbsolutePositionForObject(tab)
	test.Drag(v.w.Canvas(), pos.AddXY(10, 12), -2*v.documentTabWidth(v.tabStrip.Size().Width)+10, 0)
	require.Equal(t, []int{2, 0, 1}, v.docOrder)
	require.True(t, tab.Selected)
	require.Equal(t, float32(0), tab.frame.Position().X)
	v.moveDocuments()
	require.Equal(t, container.PanelLeading, v.documentPane.Side())
	require.Equal(t, float32(0), v.documentContent.Position().X)
	v.moveDocuments()
	require.Equal(t, container.PanelTrailing, v.documentPane.Side())
}
