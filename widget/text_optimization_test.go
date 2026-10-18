package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRuneFitWithZeroEstimatedCharacterWidth(t *testing.T) {
	measure := func(text []rune) granio.Size { return granio.NewSize(float32(len(text))*10, 12) }
	require.Equal(t, 2, howManyRunesFit([]rune("abcd"), 25, 0, measure))
	require.Equal(t, 4, howManyRunesFit([]rune("abcd"), 25, 0,
		func([]rune) granio.Size { return granio.NewSize(0, 12) }))
	// Combining and empty text must also terminate without invalid slice bounds.
	require.Zero(t, howManyRunesFit(nil, 25, 0, measure))
}
