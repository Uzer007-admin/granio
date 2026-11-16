package container

import (
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/Uzer007-admin/granio/v3/test"
	"github.com/stretchr/testify/require"
)

type sidePanelAnimationDriver struct {
	granio.Driver
	started, stopped *granio.Animation
}

func (d *sidePanelAnimationDriver) StartAnimation(a *granio.Animation) { d.started = a }
func (d *sidePanelAnimationDriver) StopAnimation(a *granio.Animation)  { d.stopped = a }

type sidePanelAnimationApp struct {
	granio.App
	driver *sidePanelAnimationDriver
}

func (a sidePanelAnimationApp) Driver() granio.Driver { return a.driver }

func TestSidePanelReversalKeepsPanelChildSize(t *testing.T) {
	app := test.NewTempApp(t)
	driver := &sidePanelAnimationDriver{Driver: app.Driver()}
	granio.SetCurrentApp(sidePanelAnimationApp{App: app, driver: driver})
	t.Cleanup(func() { granio.SetCurrentApp(app) })
	content, panel := canvas.NewRectangle(color.Black), canvas.NewRectangle(color.White)
	split := NewSidePanel(content, panel, PanelTrailing, 200)
	split.Resize(granio.NewSize(700, 450))
	view := test.TempWidgetRenderer(t, split).(*sidePanelRenderer)
	require.Equal(t, float32(200), view.clip.Size().Width)

	split.SetPanelOpen(false)
	first := driver.started
	require.NotNil(t, first)
	first.Tick(0.5)
	require.Equal(t, float32(100), view.clip.Size().Width)
	require.Equal(t, float32(200), panel.Size().Width)
	require.Equal(t, float32(600), content.Size().Width)

	split.SetPanelOpen(true)
	require.Same(t, first, driver.stopped)
	require.NotSame(t, first, driver.started)
	driver.started.Tick(0)
	require.Equal(t, float32(100), view.clip.Size().Width)
	driver.started.Tick(1)
	driver.started.OnFinished()
	require.Equal(t, float32(200), view.clip.Size().Width)
	require.Nil(t, split.transition)

	split.SetPanelOpen(false)
	driver.started.Tick(1)
	driver.started.OnFinished()
	require.False(t, view.clip.Visible())
	require.Equal(t, float32(700), content.Size().Width)
}
