package dialog

import (
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/container"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"
)

func Test_colorButton_Layout(t *testing.T) {
	test.NewTempApp(t)

	for name, tt := range map[string]struct {
		color   color.Color
		hovered bool
	}{
		"primary": {
			color: theme.Color(theme.ColorNamePrimary),
		},
		"primary_hovered": {
			color:   theme.Color(theme.ColorNamePrimary),
			hovered: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			color := newColorButton(tt.color, nil)

			if tt.hovered {
				color.MouseIn(nil)
			}

			window := test.NewTempWindow(t, container.NewCenter(color))
			window.Resize(color.MinSize().Max(granio.NewSize(50, 50)))

			test.AssertRendersToImage(t, "color/button_layout_"+name+".png", window.Canvas())
		})
	}
}
