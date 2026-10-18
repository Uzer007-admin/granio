package cache

import (
	"image"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/stretchr/testify/assert"
)

func TestCacheTouchAfterIdle(t *testing.T) {
	testClearAll()
	t.Cleanup(testClearAll)
	clock := &timeMock{}
	clock.setTime(10, 0)
	previousClean := lastClean
	t.Cleanup(func() { lastClean = previousClean })
	lastClean = clock.now.Add(-cacheCleanCooldown)
	old := &dummyWidget{}
	Renderer(old)
	Clean(true)

	// Advance the real clock while keeping the last frame's cached sample.
	clock.now = clock.now.Add(2 * ValidDuration)
	fresh := &dummyWidget{}
	want := Renderer(fresh)
	SetFontMetrics("new after idle", 14, granio.TextStyle{}, nil, granio.NewSize(25, 11), 9)
	SetSvg("new after idle", nil, image.NewNRGBA(image.Rect(0, 0, 2, 2)), 2, 2)
	SetBlurKernel(3, []float32{1})
	Clean(true)

	got, found := CachedRenderer(fresh)
	assert.True(t, found, "a renderer used immediately before cleanup must survive")
	assert.Same(t, want, got)
	assert.False(t, IsRendered(old), "untouched expired renderers must still be evicted")
	size, baseline := GetFontMetrics("new after idle", 14, granio.TextStyle{}, nil)
	assert.Equal(t, granio.NewSize(25, 11), size)
	assert.Equal(t, float32(9), baseline)
	assert.NotNil(t, GetSvg("new after idle", nil, 2, 2))
	_, found = GetBlurKernel(3)
	assert.True(t, found)
}

func BenchmarkCacheTouch(b *testing.B) {
	item := &expiringCache{}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		item.setAlive()
	}
}
