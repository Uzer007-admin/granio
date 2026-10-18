//go:build !windows || !ci

package mobile

import (
	"image/color"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	fynecanvas "github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/driver/mobile"
	"github.com/Uzer007-admin/granio/v2/internal/async"
	"github.com/Uzer007-admin/granio/v2/internal/driver/common"
	_ "github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/Uzer007-admin/granio/v2/widget"

	"github.com/stretchr/testify/assert"
)

var d *driver

func TestMain(m *testing.M) {
	currentApp := granio.CurrentApp()
	tester := newTestMobileApp()
	d = tester.Driver().(*driver)
	d.queuedFuncs = async.NewUnboundedChan[func()]()
	granio.SetCurrentApp(tester)

	waitForStart := make(chan struct{})
	go func() {
		// Wait for app loop to be running (plus a moment in case of scheduling switches).
		<-waitForStart

		// Just like the GLFW tests, wait a short while for the driver to start
		time.Sleep(time.Millisecond * 100)

		ret := m.Run()
		granio.SetCurrentApp(currentApp)
		os.Exit(ret) //revive:disable-line:redundant-test-main-exit
	}()

	close(waitForStart) // Signal that execution can continue.
	tester.Run()
}

func Test_canvas_ChildMinSizeChangeAffectsAncestorsUpToRoot(t *testing.T) {
	c := newCanvas(granio.CurrentDevice()).(*canvas)
	leftObj1 := fynecanvas.NewRectangle(color.Black)
	leftObj1.SetMinSize(granio.NewSize(100, 50))
	leftObj2 := fynecanvas.NewRectangle(color.Black)
	leftObj2.SetMinSize(granio.NewSize(100, 50))
	leftCol := container.NewVBox(leftObj1, leftObj2)
	rightObj1 := fynecanvas.NewRectangle(color.Black)
	rightObj1.SetMinSize(granio.NewSize(100, 50))
	rightObj2 := fynecanvas.NewRectangle(color.Black)
	rightObj2.SetMinSize(granio.NewSize(100, 50))
	rightCol := container.NewVBox(rightObj1, rightObj2)
	content := container.NewHBox(leftCol, rightCol)
	c.SetContent(container.NewCenter(content))
	c.Resize(granio.NewSize(300, 300))

	oldContentSize := granio.NewSize(200+theme.Padding(), 100+theme.Padding())
	assert.Equal(t, oldContentSize, content.Size())

	leftObj1.SetMinSize(granio.NewSize(110, 60))
	c.EnsureMinSize()

	expectedContentSize := oldContentSize.Add(granio.NewSize(10, 10))
	assert.Equal(t, expectedContentSize, content.Size())
}

func Test_canvas_DraggingOutOfWidget(t *testing.T) {
	c := newCanvas(granio.CurrentDevice()).(*canvas)
	slider := widget.NewSlider(0.0, 100.0)
	c.SetContent(container.NewGridWithRows(2, slider, widget.NewLabel("Outside")))
	c.Resize(granio.NewSize(100, 200))

	assert.Zero(t, slider.Value)
	lastValue := slider.Value

	dragged := false
	c.tapDown(granio.NewPos(23, 13), 0)
	c.tapMove(granio.NewPos(30, 13), 0, func(wid granio.Draggable, ev *granio.DragEvent) {
		assert.Equal(t, slider, wid)
		wid.Dragged(ev)
		dragged = true
	})
	assert.True(t, dragged)
	assert.Greater(t, slider.Value, lastValue)
	lastValue = slider.Value

	dragged = false
	c.tapMove(granio.NewPos(40, 120), 0, func(wid granio.Draggable, ev *granio.DragEvent) {
		assert.Equal(t, slider, wid)
		wid.Dragged(ev)
		dragged = true
	})
	assert.True(t, dragged)
	assert.Greater(t, slider.Value, lastValue)
}

func Test_canvas_Focusable(t *testing.T) {
	// Discarding log output for tests
	// The following method logs an error:
	// wid.Tapped(ev) on lines 148 and 159
	// c.Focus(content) on line 168
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	c := newCanvas(granio.CurrentDevice()).(*canvas)
	content := newFocusableEntry(c)
	c.SetContent(content)
	content.Resize(granio.NewSize(25, 25))

	pos := granio.NewPos(10, 10)
	d.DoFromGoroutine(func() {
		c.tapDown(pos, 0)
		c.tapUp(pos, 0, func(wid granio.Tappable, ev *granio.PointEvent) {
			wid.Tapped(ev)
		}, nil, nil, nil)
	}, true)
	waitAndCheck(tapDoubleDelay/time.Millisecond+150, func() {
		assert.Equal(t, 1, content.focusedTimes)
		assert.Equal(t, 0, content.unfocusedTimes)
	})

	d.DoFromGoroutine(func() {
		c.tapDown(pos, 1)
		c.tapUp(pos, 1, func(wid granio.Tappable, ev *granio.PointEvent) {
			wid.Tapped(ev)
		}, nil, nil, nil)
	}, true)
	waitAndCheck(tapDoubleDelay/time.Millisecond+150, func() {
		assert.Equal(t, 1, content.focusedTimes)
		assert.Equal(t, 0, content.unfocusedTimes)
	})

	d.DoFromGoroutine(func() {
		c.Focus(content)
		assert.Equal(t, 1, content.focusedTimes)
		assert.Equal(t, 0, content.unfocusedTimes)

		c.Unfocus()
		assert.Equal(t, 1, content.focusedTimes)
		assert.Equal(t, 1, content.unfocusedTimes)

		content.Disable()
		c.Focus(content)
		assert.Equal(t, 1, content.focusedTimes)
		assert.Equal(t, 1, content.unfocusedTimes)

		c.tapDown(granio.NewPos(10, 10), 2)
		assert.Equal(t, 1, content.focusedTimes)
		assert.Equal(t, 1, content.unfocusedTimes)
	}, true)
}

func Test_canvas_InteractiveArea(t *testing.T) {
	dev := &device{
		safeTop:    17,
		safeLeft:   42,
		safeBottom: 71,
		safeRight:  24,
	}
	scale := dev.SystemScaleForWindow(nil)

	c := newCanvas(dev)
	c.SetContent(fynecanvas.NewRectangle(color.Black))
	canvasSize := granio.NewSize(600, 800)
	c.(*canvas).Resize(canvasSize)

	t.Run("for canvas with size", func(t *testing.T) {
		pos, size := c.InteractiveArea()
		assert.Equal(t, granio.NewPos(float32(dev.safeLeft)/scale, float32(dev.safeTop)/scale), pos)
		assert.Equal(t, canvasSize.SubtractWidthHeight(float32(dev.safeLeft+dev.safeRight)/scale, float32(dev.safeTop+dev.safeBottom)/scale), size)
	})

	t.Run("when canvas size has changed", func(t *testing.T) {
		changedCanvasSize := granio.NewSize(800, 600)
		c.(*canvas).Resize(changedCanvasSize)
		pos, size := c.InteractiveArea()
		assert.Equal(t, granio.NewPos(float32(dev.safeLeft)/scale, float32(dev.safeTop)/scale), pos)
		assert.Equal(t, changedCanvasSize.SubtractWidthHeight(float32(dev.safeLeft+dev.safeRight)/scale, float32(dev.safeTop+dev.safeBottom)/scale), size)
	})

	t.Run("when canvas got hovering window head", func(t *testing.T) {
		c.(*canvas).Resize(canvasSize)
		hoveringMenu := fynecanvas.NewRectangle(color.Black)
		hoveringMenu.SetMinSize(granio.NewSize(17, 17))
		windowHead := container.NewHBox(hoveringMenu) // a single object in window head is considered to be the hovering menu
		c.(*canvas).setWindowHead(windowHead)
		pos, size := c.InteractiveArea()
		assert.Equal(t, granio.NewPos(float32(dev.safeLeft)/scale, float32(dev.safeTop)/scale), pos)
		assert.Equal(t, canvasSize.SubtractWidthHeight(float32(dev.safeLeft+dev.safeRight)/scale, float32(dev.safeTop+dev.safeBottom)/scale), size)
	})

	t.Run("when canvas got displacing window head", func(t *testing.T) {
		c.(*canvas).Resize(canvasSize)
		menu := fynecanvas.NewRectangle(color.Black)
		menu.SetMinSize(granio.NewSize(17, 17))
		title := fynecanvas.NewRectangle(color.Black)
		title.SetMinSize(granio.NewSize(42, 20))
		expectedOffset := 20 + 2*theme.Padding()
		windowHead := container.NewHBox(menu, title) // two or more objects in window head are considered to be the window title bar
		c.(*canvas).setWindowHead(windowHead)
		pos, size := c.InteractiveArea()
		assert.Equal(t, granio.NewPos(float32(dev.safeLeft)/scale, float32(dev.safeTop)/scale+expectedOffset), pos)
		expectedWidth := canvasSize.Width - float32(dev.safeLeft+dev.safeRight)/scale
		expectedHeight := canvasSize.Height - float32(dev.safeTop+dev.safeBottom)/scale - expectedOffset
		assert.InDelta(t, expectedWidth, size.Width, 0.001)
		assert.InDelta(t, expectedHeight, size.Height, 0.001)
	})
}

func Test_canvas_PixelCoordinateAtPosition(t *testing.T) {
	c := newCanvas(granio.CurrentDevice()).(*canvas)

	pos := granio.NewPos(4, 4)
	c.scale = 2.5
	x, y := c.PixelCoordinateForPosition(pos)
	assert.Equal(t, 10, x)
	assert.Equal(t, 10, y)
}

func Test_canvas_ResizeWithModalPopUpOverlay(t *testing.T) {
	c := newCanvas(granio.CurrentDevice()).(*canvas)

	c.SetContent(widget.NewLabel("Content"))

	popup := widget.NewModalPopUp(widget.NewLabel("PopUp"), c)
	popupBgSize := granio.NewSize(200, 200)
	popup.Show()
	popup.Resize(popupBgSize)

	canvasSize := granio.NewSize(600, 700)
	c.Resize(canvasSize)

	// get popup content padding dynamically
	popupContentPadding := popup.MinSize().Subtract(popup.Content.MinSize())

	// TODO modal non-zero again
	assert.Equal(t, popupBgSize.Subtract(popupContentPadding), popup.Content.Size())
	assert.Equal(t, canvasSize, c.Overlays().Top().Size())
}

func Test_canvas_Tappable(t *testing.T) {
	content := &touchableLabel{Label: widget.NewLabel("Hi\nHi\nHi"), ids: make(map[int]bool)}
	content.ExtendBaseWidget(content)
	c := newCanvas(granio.CurrentDevice()).(*canvas)
	c.SetContent(content)
	c.Resize(granio.NewSize(36, 24))
	content.Resize(granio.NewSize(24, 24))

	c.tapDown(granio.NewPos(15, 15), 0)
	assert.True(t, content.down)

	c.tapUp(granio.NewPos(15, 15), 0, func(granio.Tappable, *granio.PointEvent) {
	}, func(granio.SecondaryTappable, *granio.PointEvent) {
	}, func(granio.DoubleTappable, *granio.PointEvent) {
	}, func(granio.Draggable, *granio.DragEvent) {
	})
	assert.True(t, content.up)

	c.tapDown(granio.NewPos(15, 15), 0)
	c.tapMove(granio.NewPos(35, 15), 0, func(wid granio.Draggable, ev *granio.DragEvent) {
		wid.Dragged(ev)
	})
	assert.True(t, content.cancel)
}

func Test_canvas_TouchID(t *testing.T) {
	content := &touchableLabel{Label: widget.NewLabel("Hi\nHi\nHi"), ids: make(map[int]bool)}
	content.ExtendBaseWidget(content)
	c := newCanvas(granio.CurrentDevice()).(*canvas)
	c.SetContent(content)
	c.Resize(granio.NewSize(36, 24))
	content.Resize(granio.NewSize(24, 24))

	c.tapDown(granio.NewPos(15, 15), 0)
	c.tapUp(granio.NewPos(15, 15), 0, func(granio.Tappable, *granio.PointEvent) {
	}, func(granio.SecondaryTappable, *granio.PointEvent) {
	}, func(granio.DoubleTappable, *granio.PointEvent) {
	}, func(granio.Draggable, *granio.DragEvent) {
	})
	assert.True(t, content.ids[0])

	c.tapDown(granio.NewPos(15, 15), 1)
	c.tapUp(granio.NewPos(15, 15), 1, func(granio.Tappable, *granio.PointEvent) {
	}, func(granio.SecondaryTappable, *granio.PointEvent) {
	}, func(granio.DoubleTappable, *granio.PointEvent) {
	}, func(granio.Draggable, *granio.DragEvent) {
	})
	assert.True(t, content.ids[1])
}

func Test_canvas_Tapped(t *testing.T) {
	tapped := false
	altTapped := false
	buttonTap := false
	var pointEvent *granio.PointEvent
	var tappedObj granio.Tappable
	button := widget.NewButton("Test", func() {
		buttonTap = true
	})
	c := newCanvas(granio.CurrentDevice()).(*canvas)
	c.SetContent(button)
	c.Resize(granio.NewSize(36, 24))
	button.Move(granio.NewPos(3, 3))

	tapPos := granio.NewPos(6, 6)
	c.tapDown(tapPos, 0)
	c.tapUp(tapPos, 0, func(wid granio.Tappable, ev *granio.PointEvent) {
		tapped = true
		tappedObj = wid
		pointEvent = ev
		wid.Tapped(ev)
	}, func(wid granio.SecondaryTappable, ev *granio.PointEvent) {
		altTapped = true
		wid.TappedSecondary(ev)
	}, func(wid granio.DoubleTappable, ev *granio.PointEvent) {
		wid.DoubleTapped(ev)
	}, func(granio.Draggable, *granio.DragEvent) {
	})

	assert.True(t, tapped, "tap primary")
	assert.False(t, altTapped, "don't tap secondary")
	assert.True(t, buttonTap, "button should be tapped")
	assert.Equal(t, button, tappedObj)
	if assert.NotNil(t, pointEvent) {
		assert.Equal(t, granio.NewPos(6, 6), pointEvent.AbsolutePosition)
		assert.Equal(t, granio.NewPos(3, 3), pointEvent.Position)
	}
}

func Test_canvas_TappedAndDoubleTapped(t *testing.T) {
	tapped := 0
	but := newDoubleTappableButton()
	but.OnTapped = func() {
		tapped = 1
	}
	but.onDoubleTap = func() {
		tapped = 2
	}

	c := newCanvas(granio.CurrentDevice()).(*canvas)
	c.SetContent(container.NewStack(but))
	c.Resize(granio.NewSize(36, 24))

	simulateTap(c)
	waitAndCheck(700, func() {
		assert.Equal(t, 1, tapped)
	})

	simulateTap(c)
	simulateTap(c)
	waitAndCheck(700, func() {
		assert.Equal(t, 2, tapped)
	})
}

func Test_canvas_TappedMulti(t *testing.T) {
	buttonTap := false
	button := widget.NewButton("Test", func() {
		buttonTap = true
	})
	c := newCanvas(granio.CurrentDevice()).(*canvas)
	c.SetContent(button)
	c.Resize(granio.NewSize(36, 24))
	button.Move(granio.NewPos(3, 3))

	tapPos := granio.NewPos(6, 6)
	c.tapDown(tapPos, 0)
	c.tapUp(tapPos, 1, func(wid granio.Tappable, ev *granio.PointEvent) { // different tapID
		wid.Tapped(ev)
	}, func(granio.SecondaryTappable, *granio.PointEvent) {
	}, func(wid granio.DoubleTappable, ev *granio.PointEvent) {
		wid.DoubleTapped(ev)
	}, func(granio.Draggable, *granio.DragEvent) {
	})

	assert.False(t, buttonTap, "button should not be tapped")
}

func Test_canvas_TappedSecondary(t *testing.T) {
	var pointEvent *granio.PointEvent
	var altTappedObj granio.SecondaryTappable
	obj := &tappableLabel{}
	obj.ExtendBaseWidget(obj)
	c := newCanvas(granio.CurrentDevice()).(*canvas)
	c.SetContent(obj)
	c.Resize(granio.NewSize(36, 24))
	obj.Move(granio.NewPos(3, 3))

	tapPos := granio.NewPos(6, 6)
	c.tapDown(tapPos, 0)
	time.Sleep(310 * time.Millisecond)
	c.tapUp(tapPos, 0, func(wid granio.Tappable, ev *granio.PointEvent) {
		obj.tap = true
		wid.Tapped(ev)
	}, func(wid granio.SecondaryTappable, ev *granio.PointEvent) {
		obj.altTap = true
		altTappedObj = wid
		pointEvent = ev
		wid.TappedSecondary(ev)
	}, func(wid granio.DoubleTappable, ev *granio.PointEvent) {
		wid.DoubleTapped(ev)
	}, func(granio.Draggable, *granio.DragEvent) {
	})

	assert.False(t, obj.tap, "don't tap primary")
	assert.True(t, obj.altTap, "tap secondary")
	assert.Equal(t, obj, altTappedObj)
	if assert.NotNil(t, pointEvent) {
		assert.Equal(t, granio.NewPos(6, 6), pointEvent.AbsolutePosition)
		assert.Equal(t, granio.NewPos(3, 3), pointEvent.Position)
	}
}

type touchableLabel struct {
	*widget.Label
	down, up, cancel bool
	ids              map[int]bool
}

func (t *touchableLabel) TouchDown(ev *mobile.TouchEvent) {
	t.down = true
	t.ids[ev.ID] = true
}

func (t *touchableLabel) TouchUp(ev *mobile.TouchEvent) {
	t.up = true
	t.ids[ev.ID] = true
}

func (t *touchableLabel) TouchCancel(ev *mobile.TouchEvent) {
	t.cancel = true
	t.ids[ev.ID] = true
}

type tappableLabel struct {
	widget.Label
	tap, altTap bool
}

func (t *tappableLabel) Tapped(_ *granio.PointEvent) {
	t.tap = true
}

func (t *tappableLabel) TappedSecondary(_ *granio.PointEvent) {
	t.altTap = true
}

type focusableEntry struct {
	widget.Entry
	c granio.Canvas

	focusedTimes   int
	unfocusedTimes int
}

func newFocusableEntry(c granio.Canvas) *focusableEntry {
	entry := &focusableEntry{c: c}
	entry.ExtendBaseWidget(entry)
	return entry
}

func (f *focusableEntry) FocusGained() {
	f.focusedTimes++
	f.Entry.FocusGained()
}

func (f *focusableEntry) FocusLost() {
	f.unfocusedTimes++
	f.Entry.FocusLost()
}

func (f *focusableEntry) Tapped(*granio.PointEvent) {
	f.c.Focus(f)
}

type doubleTappableButton struct {
	widget.Button

	onDoubleTap func()
}

func (t *doubleTappableButton) DoubleTapped(_ *granio.PointEvent) {
	t.onDoubleTap()
}

func newDoubleTappableButton() *doubleTappableButton {
	but := &doubleTappableButton{}
	but.ExtendBaseWidget(but)

	return but
}

func simulateTap(c *canvas) {
	c.tapDown(granio.NewPos(15, 15), 0)
	time.Sleep(50 * time.Millisecond)
	c.tapUp(granio.NewPos(15, 15), 0, func(wid granio.Tappable, ev *granio.PointEvent) {
		wid.Tapped(ev)
	}, func(granio.SecondaryTappable, *granio.PointEvent) {
	}, func(wid granio.DoubleTappable, ev *granio.PointEvent) {
		wid.DoubleTapped(ev)
	}, func(granio.Draggable, *granio.DragEvent) {
	})
}

type mobileApp struct {
	granio.App
	driver granio.Driver
}

func (a *mobileApp) Driver() granio.Driver {
	return a.driver
}

func (*mobileApp) Run() {
	// This is an incomplete driver loop - our CI does not currently support booting the mobile graphics
	// TODO replace with a full mobileApp.Run() once that is resolved
	async.SetMainGoroutine()

	for fn := range d.queuedFuncs.Out() {
		fn()
	}
}

func newTestMobileApp() granio.App {
	return &mobileApp{
		App:    granio.CurrentApp(),
		driver: NewGoMobileDriver(),
	}
}

func waitAndCheck(msWait time.Duration, fn func()) {
	waitForCheck := common.DonePool.Get()

	go func() {
		defer common.DonePool.Put(waitForCheck)

		time.Sleep(msWait * time.Millisecond)
		d.DoFromGoroutine(func() {
			fn()

			waitForCheck <- struct{}{}
		}, true)
	}()
	<-waitForCheck
}
