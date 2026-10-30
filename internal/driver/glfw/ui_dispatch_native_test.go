//go:build !no_glfw && !mobile && !wasm && !test_web_driver && !ci

package glfw

import (
	"image/color"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/stretchr/testify/require"
)

func TestNativeUIQueueYieldsToFrames(t *testing.T) {
	w := createWindowWithResizeCallback("UI queue fairness")
	defer w.Close()
	w.SetPadded(false)
	rectangle := canvas.NewRectangle(color.NRGBA{G: 255, A: 255})
	w.SetContent(rectangle)
	w.Resize(granio.NewSize(320, 240))
	runOnMain(func() { w.window.Show() })
	original := int(d.requestedFrameRate.Load())
	require.NoError(t, d.SetFrameRate(60))
	defer func() { _ = d.SetFrameRate(original) }()
	const count = 48
	var processed, intermediateFrames atomic.Int32
	animation := granio.NewAnimation(time.Second, func(float32) {
		if p := processed.Load(); p > 0 && p < count {
			intermediateFrames.Add(1)
		}
		rectangle.Refresh()
	})
	defer d.StopAnimation(animation)
	finished := make(chan struct{})
	started := time.Now()
	runOnMain(func() {
		d.StartAnimation(animation)
		// Publish a finite batch while the UI thread is occupied. Every callback
		// is short, but the batch exceeds several complete frame intervals.
		for range count {
			runOnMainWithWait(func() {
				time.Sleep(2 * time.Millisecond)
				processed.Add(1)
			}, false)
		}
		runOnMainWithWait(func() { close(finished) }, false)
	})
	select {
	case <-finished:
	case <-time.After(3 * time.Second):
		t.Fatal("UI queue failed to drain")
	}
	t.Logf("callbacks=%d intermediate_frames=%d elapsed=%s", processed.Load(), intermediateFrames.Load(), time.Since(started))
	require.EqualValues(t, count, processed.Load(), "accepted callbacks must not be discarded")
	require.Positive(t, intermediateFrames.Load(), "queued UI callbacks starved animation and presentation until the whole batch completed")
}

func TestNativeSynchronousMainDispatchReentrant(t *testing.T) {
	var order []int
	runOnMain(func() {
		order = append(order, 1)
		runOnMain(func() {
			order = append(order, 2)
		})
		order = append(order, 3)
	})
	require.Equal(t, []int{1, 2, 3}, order)
}

func TestNativeAsynchronousMainDispatchStaysDeferred(t *testing.T) {
	var order []int
	finished := make(chan struct{})
	runOnMain(func() {
		order = append(order, 1)
		runOnMainWithWait(func() {
			order = append(order, 3)
			close(finished)
		}, false)
		order = append(order, 2)
	})
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("asynchronous dispatch failed to complete")
	}
	require.Equal(t, []int{1, 2, 3}, order)
}
