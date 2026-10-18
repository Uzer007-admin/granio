//go:build !no_glfw && !wasm && !test_web_driver && !mobile

package glfw

import (
	"image/color"
	"os"
	"testing"
	"time"

	native "github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/data/binding"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"
	"github.com/stretchr/testify/require"
)

func requireCompletionTest(t *testing.T) {
	if os.Getenv("FYNE_TEST_TZ_FUNCTIONAL") != "1" {
		t.Skip("short native visual completion checks")
	}
}

func TestNativeModernScopeVisibilityAndEscape(t *testing.T) {
	requireCompletionTest(t)
	w := newVisualTestWindow("Modern scope and Escape")
	defer runOnMain(w.Close)
	m, err := theme.NewModern(theme.ModernOptions{})
	require.NoError(t, err)
	var original granio.Theme
	e := widget.NewEntry()
	owner := canvas.NewRectangle(color.White)
	parent := container.NewStack(owner)
	runOnMain(func() {
		original = granio.CurrentApp().Settings().Theme()
		granio.CurrentApp().Settings().SetTheme(m)
		e.SetText("Русский English")
		w.SetContent(container.NewVBox(e, parent))
		w.Resize(granio.NewSize(350, 220))
		w.Show()
	})
	defer runOnMain(func() { granio.CurrentApp().Settings().SetTheme(original) })
	var first, hidden, resumed, minimized int
	iconified := false
	runOnMain(func() {
		ticks := 0
		a := granio.NewAnimation(time.Hour, func(float32) { ticks++ })
		a.Object = owner
		d.animation.Start(a)
		d.prepareAnimationVisibility()
		d.animation.TickAnimations()
		first = ticks
		parent.Hide()
		d.prepareAnimationVisibility()
		d.animation.TickAnimations()
		hidden = ticks
		parent.Show()
		d.prepareAnimationVisibility()
		d.animation.TickAnimations()
		resumed = ticks
		w.Hide()
		d.prepareAnimationVisibility()
		d.animation.TickAnimations()
		minimized = ticks
		w.Show()
		w.viewport.Iconify()
		iconified = w.viewport.GetAttrib(native.Iconified) != 0
		if iconified {
			d.prepareAnimationVisibility()
			d.animation.TickAnimations()
			minimized = ticks
		}
		w.viewport.Restore()
		d.animation.Stop(a)
		w.canvas.Focus(e)
		e.AccessibilitySetSelection(0, 7)
		popup := widget.NewPopUp(widget.NewLabel("Popup"), w.canvas)
		popup.Show()
		w.processKeyPressed(granio.KeyEscape, granio.KeyEscape, 0, press, 0)
	})
	require.Equal(t, 1, first)
	require.Equal(t, first, hidden)
	require.Equal(t, 2, resumed)
	require.Equal(t, resumed, minimized)
	var focus granio.Focusable
	var cursor, start, end int
	runOnMain(func() { focus = w.canvas.Focused(); cursor, start, end = e.AccessibilityTextSelection() })
	require.Same(t, e, focus)
	require.Equal(t, 0, start)
	require.Equal(t, 7, end)
	t.Logf("hidden ancestor/window paused; native minimized flag observed=%t; restored selection=%d,%d,%d", iconified, cursor, start, end)
}

func TestNativeVisualHundredLifecycleAndResizeCycles(t *testing.T) {
	requireCompletionTest(t)
	w := newVisualTestWindow("100 short functional cycles")
	defer runOnMain(w.Close)
	entry := widget.NewEntry()
	var initialGeneration uint64
	runOnMain(func() {
		entry.SetText("Данные сохраняются")
		w.SetPadded(false)
		w.SetContent(container.NewVBox(entry, widget.NewButton("Действие", nil)))
		_, err := w.ConfigureVisuals(desktop.VisualOptions{PrepareAlpha: true})
		if err != nil {
			t.Error(err)
		}
		w.Resize(granio.NewSize(330, 220))
		w.Show()
		initialGeneration = w.VisualCapabilities().Generation
	})
	backdrops := []desktop.WindowBackdrop{desktop.BackdropOpaque, desktop.BackdropTransparent, desktop.BackdropSystemBlur, desktop.BackdropMica, desktop.BackdropMicaAlt, desktop.BackdropDesktopAcrylic}
	maximized, iconified := 0, 0
	for i := 0; i < 100; i++ {
		var configureErr error
		var generation uint64
		var text string
		runOnMain(func() {
			w.Hide()
			_, configureErr = w.ConfigureVisuals(desktop.VisualOptions{Backdrop: backdrops[i%len(backdrops)], PrepareAlpha: true})
			w.Resize(granio.NewSize(330+float32(i%5)*3, 220+float32(i%3)*2))
			w.Show()
			w.viewport.Maximize()
			if w.viewport.GetAttrib(native.Maximized) != 0 {
				maximized++
			}
			w.viewport.Restore()
			w.viewport.Iconify()
			if w.viewport.GetAttrib(native.Iconified) != 0 {
				iconified++
			}
			w.viewport.Restore()
			generation = w.VisualCapabilities().Generation
			text = entry.Text
		})
		require.NoError(t, configureErr)
		require.Equal(t, initialGeneration, generation, "visual changes must not recreate native surface")
		require.Equal(t, "Данные сохраняются", text)
	}
	frame := visualFrame(w)
	require.NotEmpty(t, frame.Bounds())
	var status desktop.VisualStatus
	runOnMain(func() { status = w.VisualStatus() })
	require.True(t, status.SurfaceReady)
	require.Greater(t, status.PreparedFrames, uint32(0))
	t.Logf("100 show/hide/resize/backdrop/maximize/minimize cycles; surface generation=%d; native maximize observed=%d; native minimize observed=%d; final=%+v", initialGeneration, maximized, iconified, status)
}

func TestNativeModernLiveTextScalePreservesState(t *testing.T) {
	requireCompletionTest(t)
	w := newVisualTestWindow("Live TextScale in Fyne")
	defer runOnMain(w.Close)
	data := binding.NewString()
	require.NoError(t, data.Set("Русский English 012345\nВторая строка"))
	var e *widget.Entry
	var original granio.Theme
	runOnMain(func() {
		original = granio.CurrentApp().Settings().Theme()
		e = widget.NewMultiLineEntry()
		e.Bind(data)
		w.SetContent(e)
		w.Resize(granio.NewSize(440, 260))
		w.Show()
	})
	defer runOnMain(func() { granio.CurrentApp().Settings().SetTheme(original) })
	require.Eventually(t, func() bool { ready := false; runOnMain(func() { ready = e.Text != "" }); return ready }, time.Second, 5*time.Millisecond)
	runOnMain(func() { w.canvas.Focus(e); e.AccessibilitySetSelection(0, 7) })
	for _, scale := range []float32{1, 1.5, 2, 2.25, 1} {
		m, err := theme.NewModern(theme.ModernOptions{TextScale: scale, DisableMotion: true})
		require.NoError(t, err)
		var text string
		var start, end int
		var focused granio.Focusable
		runOnMain(func() {
			granio.CurrentApp().Settings().SetTheme(m)
			text = e.Text
			_, start, end = e.AccessibilityTextSelection()
			focused = w.canvas.Focused()
		})
		require.Equal(t, "Русский English 012345\nВторая строка", text)
		require.Equal(t, 0, start)
		require.Equal(t, 7, end)
		require.Same(t, e, focused)
		frame := visualFrame(w)
		require.NotEmpty(t, frame.Bounds())
	}
	require.NoError(t, data.Set("Binding после переключений"))
	require.Eventually(t, func() bool {
		updated := false
		runOnMain(func() { updated = e.Text == "Binding после переключений" })
		return updated
	}, time.Second, 5*time.Millisecond)
	t.Log("native live TextScale 100/150/200/225/100: text, focus, selection and binding retained")
}
