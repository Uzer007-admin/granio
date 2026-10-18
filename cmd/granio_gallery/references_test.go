package main

import (
	"image/color"
	"testing"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/canvas"
	"github.com/Uzer007-admin/granio/v2/driver/software"
	"github.com/Uzer007-admin/granio/v2/internal/driver"
	"github.com/Uzer007-admin/granio/v2/test"
	"github.com/Uzer007-admin/granio/v2/theme"
	"github.com/stretchr/testify/require"
)

func TestReferenceSceneIconsFollowVariant(t *testing.T) {
	previous := granio.CurrentApp()
	a := test.NewApp()
	t.Cleanup(func() { a.Quit(); granio.SetCurrentApp(previous) })
	checkReferenceSceneIcons(t, a)
}

func checkReferenceSceneIcons(t *testing.T, a granio.App) {
	t.Helper()
	anchor := a.NewWindow("reference test")
	defer anchor.Close()
	modern, err := theme.NewModern(theme.ModernOptions{DisableMotion: true})
	require.NoError(t, err)
	for _, variant := range []granio.ThemeVariant{theme.VariantLight, theme.VariantDark, theme.VariantLight} {
		setReferenceTheme(a, fixedModernReferenceTheme{fixedReferenceTheme{Theme: modern, variant: variant}})
		c := software.NewCanvas()
		c.SetPadded(false)
		c.SetContent(galleryReferenceScene())
		c.Resize(granio.NewSize(780, 1800))
		c.Capture()
		icons := 0
		driver.WalkVisibleObjectTree(c.Content(), func(obj granio.CanvasObject, _, _ granio.Position, _ granio.Size) bool {
			img, ok := obj.(*canvas.Image)
			if !ok || img.Image == nil {
				return false
			}
			res, ok := img.Resource.(granio.ThemedResource)
			if !ok {
				return false
			}
			icons++
			want := color.NRGBAModel.Convert(modern.Color(res.ThemeColorName(), variant)).(color.NRGBA)
			bounds := img.Image.Bounds()
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					got := color.NRGBAModel.Convert(img.Image.At(x, y)).(color.NRGBA)
					if got.A == 255 {
						require.Equal(t, want, got, "variant %d, icon %s", variant, res.Name())
						return false
					}
				}
			}
			return false
		}, nil)
		require.GreaterOrEqual(t, icons, 4)
	}
}
