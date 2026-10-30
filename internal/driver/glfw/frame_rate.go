package glfw

import (
	"fmt"
	"github.com/Uzer007-admin/granio/v2/driver/desktop"
)

var _ desktop.FrameRateDriver = (*gLDriver)(nil)

func (d *gLDriver) SetFrameRate(fps int) error {
	if fps < 0 || fps > 1000 {
		return fmt.Errorf("frame rate must be 0 (automatic) or between 1 and 1000, got %d", fps)
	}
	d.requestedFrameRate.Store(int32(fps))
	if fps == 0 {
		d.monitorFrameRateDirty.Store(true)
	}
	wakeMainLoop()
	return nil
}
func (d *gLDriver) FrameRate() int {
	if fps := d.requestedFrameRate.Load(); fps > 0 {
		return int(fps)
	}
	if fps := d.detectedFrameRate.Load(); fps > 0 {
		return int(fps)
	}
	return 60
}

// updateMonitorFrameRate coalesces native display changes. In particular, an
// idle application never queries display settings just because time has passed.
func (d *gLDriver) updateMonitorFrameRate(sample func() int) {
	if d.requestedFrameRate.Load() != 0 || !d.monitorFrameRateDirty.Swap(false) {
		return
	}
	rate := max(60, min(1000, sample()))
	d.detectedFrameRate.Store(int32(rate))
}

func (d *gLDriver) hasFrameWork() bool {
	if d.animation.HasAnimations() {
		return true
	}
	for _, win := range d.windows {
		w := win.(*window)
		if !w.closing && w.visible && w.frame.ready() && w.canvas.HasDirty() {
			return true
		}
	}
	return false
}
