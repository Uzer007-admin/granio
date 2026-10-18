package cache

import (
	"image"
	"strings"
	"sync"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/stretchr/testify/require"
)

func TestCacheLimitsEvictLeastRecentlyUsed(t *testing.T) {
	previous := GetLimits()
	t.Cleanup(func() { _ = SetLimits(previous); ResetThemeCaches() })
	ResetThemeCaches()
	limits := Limits{TextEntries: 2, TextBytes: 1024, SVGEntries: 2, SVGBytes: 1024}
	require.NoError(t, SetLimits(limits))
	setText := func(s string) { SetFontMetrics(s, 12, granio.TextStyle{}, nil, granio.NewSize(10, 12), 10) }
	setText("a")
	setText("b")
	GetFontMetrics("a", 12, granio.TextStyle{}, nil)
	setText("c")
	missing, _ := GetFontMetrics("b", 12, granio.TextStyle{}, nil)
	require.True(t, missing.IsZero(), "recently unused text survived eviction")
	setText("a")
	require.Equal(t, 2, GetUsage().TextEntries, "replacement counted twice")
	setText(strings.Repeat("x", 2048))
	require.Equal(t, 2, GetUsage().TextEntries, "oversized text was retained")

	for _, key := range []string{"a", "b"} {
		SetSvg(key, nil, image.NewNRGBA(image.Rect(0, 0, 10, 10)), 10, 10)
	}
	require.NotNil(t, GetSvg("a", nil, 10, 10))
	SetSvg("c", nil, image.NewNRGBA(image.Rect(0, 0, 10, 10)), 10, 10)
	require.Nil(t, GetSvg("b", nil, 10, 10))
	SetSvg("large", nil, image.NewNRGBA(image.Rect(0, 0, 100, 100)), 100, 100)
	require.Nil(t, GetSvg("large", nil, 100, 100))
	usage := GetUsage()
	require.LessOrEqual(t, usage.TextBytes, limits.TextBytes)
	require.LessOrEqual(t, usage.SVGBytes, limits.SVGBytes)
	require.Error(t, SetLimits(Limits{TextEntries: -1}))
	require.Equal(t, limits, GetLimits())
	require.NoError(t, SetLimits(Limits{}))
	require.Equal(t, Usage{}, GetUsage(), "disabling caches did not release retained entries")
}

func TestCacheLimitsConcurrentReadsAndWrites(t *testing.T) {
	previous := GetLimits()
	t.Cleanup(func() { _ = SetLimits(previous); ResetThemeCaches() })
	require.NoError(t, SetLimits(Limits{TextEntries: 16, TextBytes: 4096, SVGEntries: 8, SVGBytes: 4096}))
	var workers sync.WaitGroup
	for worker := 0; worker < 6; worker++ {
		workers.Go(func() {
			for i := 0; i < 1000; i++ {
				key := string(rune('a' + i%50))
				SetFontMetrics(key, 12, granio.TextStyle{}, nil, granio.NewSize(10, 12), 10)
				GetFontMetrics(key, 12, granio.TextStyle{}, nil)
				SetSvg(key, nil, image.NewNRGBA(image.Rect(0, 0, 2, 2)), 2, 2)
				GetSvg(key, nil, 2, 2)
			}
		})
	}
	workers.Wait()
	usage := GetUsage()
	require.LessOrEqual(t, usage.TextEntries, 16)
	require.LessOrEqual(t, usage.TextBytes, int64(4096))
	require.LessOrEqual(t, usage.SVGEntries, 8)
	require.LessOrEqual(t, usage.SVGBytes, int64(4096))
}
