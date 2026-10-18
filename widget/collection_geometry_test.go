package widget

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/cache"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/stretchr/testify/require"
)

func TestCollectionGeometryMatchesLinearLayout(t *testing.T) {
	const count = 10000
	rng := rand.New(rand.NewSource(42))
	overrides := make(map[int]float32)
	for i := 0; i < 200; i++ {
		overrides[rng.Intn(count)] = float32(1 + rng.Intn(100))
	}
	overrides[-1], overrides[count] = 999, 999
	var geometry collectionGeometry
	geometry.prepare(overrides, count, 20, 4)
	positions := make([]float32, count+1)
	for i := 0; i < count; i++ {
		height := float32(20)
		if custom, ok := overrides[i]; ok {
			height = custom
		}
		positions[i+1] = positions[i] + height + 4
		require.Equal(t, positions[i+1], geometry.position(i+1))
	}
	for i := 0; i < 1000; i++ {
		offset := float32(rng.Intn(int(positions[count]) + 100))
		item, end := count-1, count
		for id := 0; id < count; id++ {
			if positions[id+1] > offset {
				item = id
				break
			}
		}
		for id := 0; id < count; id++ {
			if positions[id] >= offset {
				end = id
				break
			}
		}
		require.Equal(t, item, geometry.itemAt(offset))
		require.Equal(t, end, geometry.endAt(offset))
	}
	overrides[0] = 50
	geometry.dirty = true
	geometry.prepare(overrides, count, 20, 4)
	require.Equal(t, float32(54), geometry.position(1))
}

func TestTreeViewportReusesGeometryAndSeparators(t *testing.T) {
	test.NewApp()
	ids := make([]string, 20000)
	for i := range ids {
		ids[i] = strconv.Itoa(i)
	}
	walks := 0
	tree := NewTree(func(id string) []string {
		if id == "" {
			return ids
		}
		return nil
	},
		func(id string) bool { walks++; return id == "" },
		func(bool) granio.CanvasObject { return NewLabel("row") },
		func(id string, _ bool, object granio.CanvasObject) { object.(*Label).SetText(id) })
	tree.Resize(granio.NewSize(240, 200))
	renderer := cache.Renderer(tree).(*treeRenderer)
	content := cache.Renderer(renderer.content).(*treeContentRenderer)
	walks = 0
	for i := 0; i < 10; i++ {
		tree.offset.Y = float32(10000 + i*100)
		content.Layout(renderer.content.Size())
		require.Less(t, len(content.separators), 20, "separator cache grew with scroll offset")
	}
	require.Zero(t, walks, "scrolling traversed the data source again")
	ids = ids[:100]
	tree.Refresh()
	require.Len(t, content.rows, 100, "Refresh did not invalidate the flattened rows")
}

func BenchmarkCollectionGeometryViewport(b *testing.B) {
	for _, count := range []int{1000, 100000, 1000000} {
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			var geometry collectionGeometry
			geometry.prepare(map[int]float32{1: 40, count / 2: 60, count - 1: 80}, count, 20, 4)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				offset := float32(i%count) * 24
				_ = geometry.itemAt(offset)
				_ = geometry.endAt(offset + 600)
			}
		})
	}
}
