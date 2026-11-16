//go:build android || ios || mobile

package app

import "github.com/Uzer007-admin/granio/v3"

func rootCacheDir(a granio.App) string {
	return a.(*fyneApp).storageRoot()
}
