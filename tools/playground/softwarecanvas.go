package playground

import (
	"github.com/Uzer007-admin/granio/v2/driver/software"
)

// NewSoftwareCanvas creates a new canvas in memory that can render without hardware support
func NewSoftwareCanvas() software.WindowlessCanvas {
	return software.NewCanvas()
}
