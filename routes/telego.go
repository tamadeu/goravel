package routes

import (
	"goravel/app/facades"
	"goravel/app/http/controllers"
)

func Telego() {
	if !facades.Config().GetBool("telego.enabled") {
		return
	}

	controller := controllers.NewTelegoController()
	facades.Route().Get("/telego", controller.Root)
	facades.Route().Get("/telego/api/entries", controller.Entries)
	facades.Route().Get("/telego/api/entries/{id}", controller.Show)
	facades.Route().Delete("/telego/api/entries", controller.Clear)
	facades.Route().Post("/telego/api/pause", controller.Pause)
	facades.Route().Post("/telego/api/http-client/test", controller.HttpClientTest)
	facades.Route().Static("telego/assets", "./public/telego")
	facades.Route().Get("/telego/http-client/test", controller.HttpClientTestPage)
	facades.Route().Get("/telego/{category}", controller.Index)
	facades.Route().Get("/telego/{category}/{id}", controller.Detail)
}
