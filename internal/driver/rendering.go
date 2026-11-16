package driver

import (
	"os"
	"strings"

	"github.com/Uzer007-admin/granio/v3/driver"
)

func ResolveRenderMode(request driver.RenderMode, softwareAvailable bool) driver.RenderMode {
	mode := RenderPreference(request, softwareAvailable)
	if mode == driver.RenderAuto {
		return driver.RenderOpenGL
	}
	return mode
}

// RenderPreference preserves auto mode for drivers that can fall back at startup.
func RenderPreference(request driver.RenderMode, softwareAvailable bool) driver.RenderMode {
	if request != driver.RenderAuto {
		return request
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("GRANIO_RENDERER"))) {
	case "opengl":
		return driver.RenderOpenGL
	case "software":
		if softwareAvailable {
			return driver.RenderSoftware
		}
	}
	return driver.RenderAuto
}
