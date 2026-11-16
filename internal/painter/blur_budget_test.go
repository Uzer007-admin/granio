package painter

import (
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v3/canvas"
	"github.com/stretchr/testify/require"
)

func TestSoftwareBlurBudgetLimitsAndDegrades(t *testing.T) {
	var budget SoftwareBlurBudget
	owner := &canvas.Material{SoftwareBlur: true}
	other := &canvas.Material{SoftwareBlur: true}
	now := time.Now()
	require.Zero(t, budget.Begin(owner, MaxSoftwareBlurPixels, 1, 4, now))
	require.Zero(t, budget.Begin(owner, 800, 481, 4, now))
	for _, expected := range []uint8{4, 4, 4, 8, 8, 8, 0} {
		budget.ResetFrame()
		require.Equal(t, expected, budget.Begin(owner, 800, 480, 4, now))
		require.Zero(t, budget.Begin(other, 10, 10, 4, now))
		if expected > 0 {
			budget.End(3 * time.Millisecond)
		}
		now = now.Add(16 * time.Millisecond)
	}
	budget.ResetFrame()
	require.Equal(t, uint8(4), budget.Begin(owner, 800, 480, 4, now.Add(2*time.Second)))
}
