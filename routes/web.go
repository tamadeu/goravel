package routes

import (
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support"

	"goravel/app/facades"
	"goravel/app/http/controllers"
	"goravel/app/http/middleware"
)

func Web() {
	facades.Route().Get("/", func(ctx http.Context) http.Response {
		return ctx.Response().View().Make("welcome.tmpl", map[string]any{
			"version": support.Version,
		})
	})

	dashboard := func(ctx http.Context) http.Response {
		return ctx.Response().File("./public/dashboard/index.html")
	}
	facades.Route().Get("/dashboard", dashboard)
	facades.Route().Get("/dashboard/", dashboard)
	facades.Route().Get("/dashboard/connectors", dashboard)
	facades.Route().Static("dashboard/assets", "./public/dashboard/assets")
	facades.Route().Static("public", "./public")

	authController := controllers.NewAuthController()
	facades.Route().Post("/api/auth/register", authController.Register)
	facades.Route().Post("/api/auth/login", authController.Login)
	facades.Route().Middleware(&middleware.Authenticate{}).Get("/api/auth/me", authController.Me)
	facades.Route().Middleware(&middleware.Authenticate{}).Post("/api/auth/logout", authController.Logout)

	connectorController := controllers.NewConnectorController()
	facades.Route().Middleware(&middleware.Authenticate{}).Get("/api/connectors", connectorController.Index)
	facades.Route().Middleware(&middleware.Authenticate{}).Post("/api/connectors", connectorController.Store)

	userController := controllers.NewUserController()
	facades.Route().Get("/users", userController.Index)
	facades.Route().Get("/favicon.ico", func(ctx http.Context) http.Response {
		return ctx.Response().NoContent()
	})

	Telego()
}
