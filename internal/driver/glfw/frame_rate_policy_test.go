package glfw

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAutomaticFrameRateSamplesOnlyInvalidatedConfiguration(t *testing.T) {
	driver := &gLDriver{}
	queries, rate := 0, 144
	sample := func() int { queries++; return rate }
	require.Equal(t, 60, driver.FrameRate())
	require.NoError(t, driver.SetFrameRate(0))
	driver.updateMonitorFrameRate(sample)
	require.Equal(t, 144, driver.FrameRate())
	for i := 0; i < 10000; i++ {
		driver.updateMonitorFrameRate(sample)
	}
	require.Equal(t, 1, queries, "idle/event-loop iterations must not enumerate monitor modes")

	rate = 240
	for i := 0; i < 10; i++ {
		driver.monitorFrameRateDirty.Store(true)
	}
	driver.updateMonitorFrameRate(sample)
	require.Equal(t, 2, queries, "a burst of native display events requires one query")
	require.Equal(t, 240, driver.FrameRate())

	// A disconnected fast monitor must allow the automatic rate to decrease.
	rate = 60
	driver.monitorFrameRateDirty.Store(true)
	driver.updateMonitorFrameRate(sample)
	require.Equal(t, 60, driver.FrameRate())
	require.Equal(t, 3, queries)
}

func TestFixedFrameRateDefersMonitorQueriesUntilAutomatic(t *testing.T) {
	driver := &gLDriver{}
	queries := 0
	sample := func() int { queries++; return 165 }
	require.NoError(t, driver.SetFrameRate(60))
	driver.monitorFrameRateDirty.Store(true)
	for i := 0; i < 10000; i++ {
		driver.updateMonitorFrameRate(sample)
	}
	require.Zero(t, queries)
	require.Equal(t, 60, driver.FrameRate())
	require.NoError(t, driver.SetFrameRate(0))
	driver.updateMonitorFrameRate(sample)
	require.Equal(t, 165, driver.FrameRate())
	require.Equal(t, 1, queries)

	require.NoError(t, driver.SetFrameRate(0))
	driver.updateMonitorFrameRate(sample)
	require.Equal(t, 2, queries, "explicit automatic selection refreshes the cached value")
}

func TestAutomaticFrameRateBoundsAndIdleAllocations(t *testing.T) {
	for _, rate := range []int{-1, 0, 50, 60, 144, 1000, 1001} {
		driver := &gLDriver{}
		driver.monitorFrameRateDirty.Store(true)
		driver.updateMonitorFrameRate(func() int { return rate })
		require.Equal(t, max(60, min(1000, rate)), driver.FrameRate())
	}
	driver := &gLDriver{}
	sample := func() int { t.Fatal("an idle driver queried display modes"); return 0 }
	require.Zero(t, testing.AllocsPerRun(1000, func() { driver.updateMonitorFrameRate(sample) }))
}
