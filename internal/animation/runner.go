package animation

import (
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Uzer007-admin/granio/v3"
)

const initialAnimationsCapacity = 16

// Runner is the main driver for animations package
type Runner struct {
	// Visible is supplied by a driver on its UI thread, and can include hidden
	// ancestors, inactive tabs, and minimized windows. Nil checks Object.Visible.
	Visible func(granio.CanvasObject) bool
	// The mutex protects the animation lists and frameRunning. The frame reads
	// the active list outside the lock, so only its stopped flags can change then.
	animationMutex sync.RWMutex
	frameRunning   bool

	// animations is the list of animations that are being ticked in the current frame
	animations []*anim

	// pendingAnimations is animations that have been started but not yet picked up
	// by the runner goroutine to be ticked each frame
	pendingAnimations []*anim

	// nextFrameAnimations is the list of animations that will be ticked in the next frame.
	// It is accessed only by the runner goroutine and accumulates the continuing animations
	// during a tick that are not completed.
	// At the end of a full frame of animations, the nextFrameAnimations slice is swapped with
	// the current `animations` slice which is then cleared out, while holding the mutex.
	nextFrameAnimations []*anim

	runnerStarted atomic.Bool
	runnable      atomic.Bool
}

// Start will register the passed application and initiate its ticking.
func (r *Runner) Start(a *granio.Animation) {
	if a == nil || a.Tick == nil {
		return
	}
	r.animationMutex.Lock()
	defer r.animationMutex.Unlock()
	// Starting the same handle restarts it; pending copies must never accumulate.
	for _, previous := range r.animations {
		if previous.a == a {
			previous.setStopped()
		}
	}
	// Only the runner reads active slices outside the lock. Pending slices can
	// be compacted here, so repeated retargets before a paint use bounded space.
	r.pendingAnimations = slices.DeleteFunc(r.pendingAnimations, func(previous *anim) bool {
		return previous.a == a || previous.isStopped()
	})
	if r.pendingAnimations == nil {
		r.pendingAnimations = make([]*anim, 0, initialAnimationsCapacity)
	}
	r.pendingAnimations = append(r.pendingAnimations, newAnim(a))
	r.runnable.Store(true)
	r.runnerStarted.Store(true)
}

// Stop causes an animation to stop ticking (if it was still running) and removes it from the runner.
func (r *Runner) Stop(a *granio.Animation) {
	r.animationMutex.Lock()
	defer r.animationMutex.Unlock()

	for _, item := range r.animations {
		if item.a == a {
			item.setStopped()
		}
	}
	if !r.frameRunning {
		r.animations = slices.DeleteFunc(r.animations, (*anim).isStopped)
	}
	r.pendingAnimations = slices.DeleteFunc(r.pendingAnimations, func(item *anim) bool { return item.a == a })
	if len(r.animations) == 0 && len(r.pendingAnimations) == 0 {
		r.runnerStarted.Store(false)
		r.runnable.Store(false)
	}
}

// TickAnimations progresses all running animations by one tick.
// This will be called from the driver to update objects immediately before next paint.
func (r *Runner) TickAnimations() {
	if r.runnerStarted.Load() {
		r.runOneFrame()
	}
}

// HasAnimations is a cheap concurrent check for the event loop's idle policy.
func (r *Runner) HasAnimations() bool { return r.runnerStarted.Load() && r.runnable.Load() }

// VisitObjects visits scoped owners without copying the animation list.
func (r *Runner) VisitObjects(visit func(granio.CanvasObject)) {
	r.animationMutex.RLock()
	defer r.animationMutex.RUnlock()
	for _, list := range [][]*anim{r.animations, r.pendingAnimations} {
		for _, a := range list {
			if a.a.Object != nil && !a.isStopped() {
				visit(a.a.Object)
			}
		}
	}
}

func (r *Runner) runOneFrame() {
	r.animationMutex.Lock()
	if r.frameRunning {
		r.animationMutex.Unlock()
		return
	}
	r.frameRunning = true
	// Include all animations started before this frame. A Tick callback can
	// start more animations, which remain pending for the next frame.
	r.animations = append(r.animations, r.pendingAnimations...)
	clear(r.pendingAnimations)
	r.pendingAnimations = r.pendingAnimations[:0]
	oldList := r.animations
	r.runnable.Store(false)
	r.animationMutex.Unlock()
	now := time.Now()
	disabled := false
	if app := granio.CurrentApp(); app != nil {
		if settings := app.Settings(); settings != nil {
			disabled = !settings.ShowAnimations()
			if policy, ok := settings.Theme().(granio.MotionTheme); ok {
				disabled = disabled || policy.Motion(granio.MotionHover).Disabled
			}
		}
	}
	for _, a := range oldList {
		if !a.isStopped() && r.tickAnimation(a, now, disabled) && !a.isStopped() {
			r.nextFrameAnimations = append(r.nextFrameAnimations, a)
		}
	}

	r.animationMutex.Lock()
	// Stop/Clear may have run from a later Tick callback after an earlier item
	// was scheduled for the next frame. Release those references immediately.
	tmp := r.animations
	clear(tmp)
	r.animations = slices.DeleteFunc(r.nextFrameAnimations, (*anim).isStopped)
	r.nextFrameAnimations = tmp[:0]
	r.frameRunning = false
	if len(r.pendingAnimations) > 0 {
		r.runnable.Store(true)
	}
	r.runnerStarted.Store(len(r.animations) != 0 || len(r.pendingAnimations) != 0)
	r.animationMutex.Unlock()
}

// tickAnimation will process a frame of animation and return true if this should continue animating
func (r *Runner) tickAnimation(a *anim, now time.Time, disabled bool) bool {
	duration := a.a.Duration
	disabled = disabled || a.a.Disabled
	if obj, ok := a.a.Object.(interface{ Theme() granio.Theme }); ok {
		if policy, ok := obj.Theme().(granio.MotionTheme); ok {
			disabled = disabled || policy.Motion(granio.MotionHover).Disabled
		}
	}
	if disabled {
		end := float32(1)
		if a.a.AutoReverse && a.reverse {
			end = 0
		}
		a.a.Tick(end)
		a.finish()
		return false
	}
	if obj := a.a.Object; obj != nil {
		visible := obj.Visible()
		if visible && r.Visible != nil {
			visible = r.Visible(obj)
		}
		if !visible {
			if a.pausedAt.IsZero() {
				a.pausedAt = now
			}
			return true
		}
		if !a.pausedAt.IsZero() {
			elapsed := now.Sub(a.pausedAt)
			a.start = a.start.Add(elapsed)
			a.pinTime = a.pinTime.Add(elapsed)
			a.pausedAt = time.Time{}
		}
	}
	r.runnable.Store(true)

	if duration != a.lastDuration {
		a.pinProgress = a.progressFraction(now, a.lastDuration)
		a.pinTime = now
		a.lastDuration = duration
	}

	if !now.Before(a.start.Add(duration)) {
		if a.reverse {
			a.a.Tick(0.0)
			if a.repeatsLeft == 0 {
				a.finish()
				return false
			}
			a.reverse = false
		} else {
			a.a.Tick(1.0)
			if a.a.AutoReverse {
				a.reverse = true
			}
		}
		if !a.reverse {
			if a.repeatsLeft == 0 {
				a.finish()
				return false
			}
			if a.repeatsLeft > 0 {
				a.repeatsLeft--
			}
		}

		a.resetCycle(now)
		return true
	}

	val := a.progressFraction(now, duration)
	curve := a.a.Curve
	if curve == nil {
		curve = granio.AnimationEaseInOut
	}
	if a.reverse {
		a.a.Tick(curve(1 - val))
	} else {
		a.a.Tick(curve(val))
	}

	return true
}

// Clear cancels handles on driver shutdown, including paused scopes.
func (r *Runner) Clear() {
	r.animationMutex.Lock()
	defer r.animationMutex.Unlock()
	for _, list := range [][]*anim{r.animations, r.pendingAnimations} {
		for _, a := range list {
			a.setStopped()
		}
	}
	// An in-flight tick may still read the old active slice. Mark its entries
	// stopped without mutating its storage; runOneFrame discards them on return.
	r.animations, r.pendingAnimations = nil, nil
	r.runnerStarted.Store(false)
	r.runnable.Store(false)
}
