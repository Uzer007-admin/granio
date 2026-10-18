//go:build !no_glfw && !mobile && !wasm && !test_web_driver

package glfw

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/stretchr/testify/require"
	"sync/atomic"
	"testing"
	"time"
)

func TestNativeFrameRates(t *testing.T) {
	original := int(d.requestedFrameRate.Load())
	t.Cleanup(func() { _ = d.SetFrameRate(original) })
	for _, rate := range []int{120, 144, 160, 180, 240, 280, 300, 360} {
		require.NoError(t, d.SetFrameRate(rate))
		require.Equal(t, rate, d.FrameRate())
		var ticks atomic.Int32
		animation := granio.NewAnimation(2*time.Second, func(float32) { ticks.Add(1) })
		d.StartAnimation(animation)
		time.Sleep(400 * time.Millisecond)
		d.StopAnimation(animation)
		// A 60Hz loop cannot meet this count; allow scheduling variation under race.
		require.GreaterOrEqual(t, ticks.Load(), int32(rate*3/10), "requested %dHz", rate)
		t.Logf("target=%dHz observed=%d animation ticks in 400ms", rate, ticks.Load())
	}
	require.Error(t, d.SetFrameRate(-1))
	require.Error(t, d.SetFrameRate(1001))
}

func TestQueuedWorkWakesNativeIdle(t *testing.T) {
	deadline := time.Now().Add(time.Second)
	for !loopWaiting.Load() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	start := time.Now()
	runOnMain(func() {})
	require.Less(t, time.Since(start), 100*time.Millisecond, "queued work waited for the 250ms idle deadline")
}
