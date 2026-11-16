//go:build freebsd || linux || darwin || openbsd

package app

import (
	"github.com/Uzer007-admin/granio/v3/internal/driver/mobile/event/lifecycle"
	"github.com/Uzer007-admin/granio/v3/internal/driver/mobile/event/size"
	"github.com/Uzer007-admin/granio/v3/internal/driver/mobile/gl"
)

func init() {
	theApp.glctx, theApp.worker = gl.NewContext()
}

func screenOrientation(width, height int) size.Orientation {
	if width > height {
		return size.OrientationLandscape
	}

	return size.OrientationPortrait
}

func (a *app) sendLifecycle(to lifecycle.Stage) {
	if a.lifecycleStage == to {
		return
	}
	a.events.In() <- lifecycle.Event{
		From:        a.lifecycleStage,
		To:          to,
		DrawContext: a.glctx,
	}
	a.lifecycleStage = to
}
