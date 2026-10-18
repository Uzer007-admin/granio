package theme

import (
	"testing"

	"github.com/Uzer007-admin/granio/v2"
)

func BenchmarkTheme_current(b *testing.B) {
	granio.CurrentApp().Settings().SetTheme(LightTheme())

	for n := 0; n < b.N; n++ {
		Current()
	}
}
