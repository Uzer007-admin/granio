//go:build !windows || android || ios || mobile || tinygo || noos || tamago

package theme

import "github.com/Uzer007-admin/granio/v3"

func loadSystemFonts() [4]granio.Resource { return [4]granio.Resource{} }
