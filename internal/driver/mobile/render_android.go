package mobile

import "github.com/Uzer007-admin/granio/v3/internal/driver/mobile/app"

const softwareRendererAvailable = true

func runMobileApp(f func(app.App), software bool) {
	app.MainWithSoftware(f, software)
}
