//go:build !android

package mobile

import "github.com/Uzer007-admin/granio/v3/internal/driver/mobile/app"

const softwareRendererAvailable = false

func runMobileApp(f func(app.App), _ bool) { app.Main(f) }
