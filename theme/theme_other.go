//go:build !hints

package theme

import (
	"image/color"

	"github.com/Uzer007-admin/granio/v3"
)

var (
	fallbackColor = color.Transparent
	fallbackIcon  = &granio.StaticResource{}
)
