package theme

import "github.com/Uzer007-admin/granio/v3"

// Optional looks for an optional interface on a theme or its wrapper.
func Optional[T any](th granio.Theme) (T, bool) {
	for th != nil {
		if value, ok := th.(T); ok {
			return value, true
		}
		wrapper, ok := th.(interface{ UnderlyingTheme() granio.Theme })
		if !ok {
			break
		}
		th = wrapper.UnderlyingTheme()
	}
	var zero T
	return zero, false
}
