//go:build !ci && !test

package painter_test

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/internal/painter"
	"github.com/stretchr/testify/assert"
)

func TestHangul(t *testing.T) {
	got := painter.CachedFontFace(granio.TextStyle{}, nil, nil)
	f := got.Fonts.ResolveFace('안')
	gid, ok := f.Cmap.Lookup('안')
	assert.True(t, ok)
	assert.NotZero(t, gid)
}
