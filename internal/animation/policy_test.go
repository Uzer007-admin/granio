package animation

import (
	"image/color"
	"testing"
	"time"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/stretchr/testify/require"
)

func TestRunnerRetargetsAreBoundedAndStopCancelsCompletion(t *testing.T) {
	test.NewApp()
	r := &Runner{}
	finished, ticks := 0, 0
	a := granio.NewAnimation(time.Second, func(float32) { ticks++ })
	a.OnFinished = func() { finished++ }
	for i := 0; i < 1000; i++ {
		r.Start(a)
	}
	require.LessOrEqual(t, len(r.animations)+len(r.pendingAnimations), 2)
	r.Stop(a)
	r.TickAnimations()
	require.Zero(t, ticks)
	require.Zero(t, finished)
	require.False(t, r.HasAnimations())
}

func TestRunnerCompletionAndDisabledFinalStateAreSingle(t *testing.T) {
	a := test.NewApp()
	m, err := theme.NewModern(theme.ModernOptions{DisableMotion: true})
	require.NoError(t, err)
	a.Settings().SetTheme(m)
	for _, repeats := range []int{0, granio.AnimationRepeatForever} {
		r := &Runner{}
		values, finished := []float32{}, 0
		animation := granio.NewAnimation(time.Hour, func(value float32) { values = append(values, value) })
		animation.RepeatCount = repeats
		animation.OnFinished = func() { finished++ }
		r.Start(animation)
		for i := 0; i < 4; i++ {
			r.TickAnimations()
		}
		require.Equal(t, []float32{1}, values)
		require.Equal(t, 1, finished)
		require.False(t, r.HasAnimations())
	}
	a.Settings().SetTheme(theme.DefaultTheme())
	r := &Runner{}
	finished := 0
	animation := granio.NewAnimation(time.Second, func(float32) {})
	animation.OnFinished = func() { finished++ }
	r.Start(animation)
	r.animations[0].start = time.Now().Add(-2 * time.Second)
	r.TickAnimations()
	r.TickAnimations()
	require.Equal(t, 1, finished)
}

func TestRunnerPausesHiddenOrMinimizedScopeWithoutTimeJump(t *testing.T) {
	test.NewApp()
	owner := canvas.NewRectangle(color.White)
	visible, ticks := true, 0
	r := &Runner{Visible: func(granio.CanvasObject) bool { return visible }}
	animation := granio.NewAnimation(time.Second, func(float32) { ticks++ })
	animation.Object = owner
	r.Start(animation)
	r.TickAnimations()
	require.Equal(t, 1, ticks)
	visible = false // driver supplies false for minimized or hidden ancestors
	r.TickAnimations()
	require.False(t, r.HasAnimations())
	require.Equal(t, 1, ticks)
	scoped := r.animations[0]
	scoped.pausedAt = scoped.pausedAt.Add(-time.Hour)
	start := scoped.start
	visible = true
	r.TickAnimations()
	require.True(t, r.HasAnimations())
	require.Equal(t, 2, ticks)
	require.GreaterOrEqual(t, scoped.start.Sub(start), time.Hour)
	owner.Hide()
	r.TickAnimations()
	require.Equal(t, 2, ticks)
	r.Clear()
	require.False(t, r.HasAnimations())
	require.Empty(t, r.animations)
}

func TestRunnerCanClearInsideTick(t *testing.T) {
	test.NewApp()
	r := &Runner{}
	r.Start(granio.NewAnimation(time.Second, func(float32) {}))
	r.Start(granio.NewAnimation(time.Second, func(float32) { r.Clear() }))
	r.TickAnimations() // picks up the second (pending) handle
	r.TickAnimations()
	require.Empty(t, r.animations)
	require.False(t, r.HasAnimations())
}

func TestRunnerStopInsideFinalTickDoesNotComplete(t *testing.T) {
	test.NewApp()
	r := &Runner{}
	finished := 0
	var a *granio.Animation
	a = granio.NewAnimation(0, func(float32) { r.Stop(a) })
	a.OnFinished = func() { finished++ }
	r.Start(a)
	r.TickAnimations()
	require.Zero(t, finished)
	require.Empty(t, r.animations)
}
