//go:build !wasm && !test_web_driver

package glfw

import (
	glfw "github.com/Uzer007-admin/glfw/v3.5/glfw"
	"github.com/Uzer007-admin/granio/v3"
	"github.com/Uzer007-admin/granio/v3/internal/driver"
)

// One visible-tree walk per active frame, shared by all scoped animations.
// Inactive tabs/hidden ancestors are excluded; virtualized rows are not created.
func (d *gLDriver) prepareAnimationVisibility() {
	owners := d.animationOwners
	if owners == nil {
		owners = make(map[granio.CanvasObject]bool)
		d.animationOwners = owners
	}
	clear(owners)
	d.animation.VisitObjects(func(obj granio.CanvasObject) { owners[obj] = false })
	remaining := len(owners)
	if remaining == 0 {
		return
	}
	if d.animation.Visible == nil {
		d.animation.Visible = func(obj granio.CanvasObject) bool { return d.animationOwners[obj] }
	}
	mark := func(obj granio.CanvasObject, _ granio.Position, _ granio.Position, _ granio.Size) bool {
		if found, requested := owners[obj]; requested && !found {
			owners[obj] = true
			remaining--
		}
		return remaining == 0
	}
	for _, win := range d.windows {
		w, ok := win.(*window)
		if !ok || !w.visible || w.isClosing() || w.viewport == nil || w.viewport.GetAttrib(glfw.Iconified) != 0 {
			continue
		}
		if driver.WalkVisibleObjectTree(w.canvas.Content(), mark, nil) {
			return
		}
		for _, overlay := range w.canvas.Overlays().List() {
			if driver.WalkVisibleObjectTree(overlay, mark, nil) {
				return
			}
		}
	}
}
