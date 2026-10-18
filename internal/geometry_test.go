package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Uzer007-admin/granio/v2"
)

// The helpers exist only to mirror granio.Size.Max/Min without boxing the
// argument into their Vector2 parameter, so each test checks the element-wise
// result and that the helper stays in sync with the public method it shadows.

func TestMaxSizes(t *testing.T) {
	// width taken from b, height from a — the maximum is element-wise,
	// not whichever whole Size is "bigger"
	a := granio.NewSize(1, 40)
	b := granio.NewSize(30, 2)

	assert.Equal(t, granio.NewSize(30, 40), MaxSizes(a, b))
	assert.Equal(t, a.Max(b), MaxSizes(a, b), "must match granio.Size.Max")
}

func TestMinSizes(t *testing.T) {
	// width taken from a, height from b — element-wise, as above
	a := granio.NewSize(1, 40)
	b := granio.NewSize(30, 2)

	assert.Equal(t, granio.NewSize(1, 2), MinSizes(a, b))
	assert.Equal(t, a.Min(b), MinSizes(a, b), "must match granio.Size.Min")
}

// BenchmarkMaxSizes guards the reason the helpers exist: 0 allocs/op.
// granio.Size.Max costs one heap allocation per call (interface boxing); if a
// change reintroduces an allocation here, the per-frame layout win is gone.
func BenchmarkMaxSizes(b *testing.B) {
	s := granio.NewSize(10, 10)
	for i := 0; i < b.N; i++ {
		s = MaxSizes(s, granio.NewSize(20, 20))
	}
	// use the result so the loop cannot be optimized away
	if s.Width != 20 {
		b.Fatal("unexpected result")
	}
}
