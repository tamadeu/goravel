package routes

import (
	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/http/controllers"
	"goravel/app/http/middleware"
)

func Web() {
	dashboard := func(ctx http.Context) http.Response {
		return ctx.Response().File("./public/dashboard/index.html")
	}
	facades.Route().Get("/", dashboard)
	facades.Route().Get("/login", dashboard)
	facades.Route().Get("/connectors", dashboard)
	facades.Route().Get("/destinations", dashboard)
	facades.Route().Get("/pipelines", dashboard)
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
	facades.Route().Middleware(&middleware.Authenticate{}).Put("/api/connectors/{id}", connectorController.Update)
	facades.Route().Middleware(&middleware.Authenticate{}).Delete("/api/connectors/{id}", connectorController.Destroy)
	facades.Route().Middleware(&middleware.Authenticate{}).Post("/api/connectors/{id}/test", connectorController.Test)
	facades.Route().Middleware(&middleware.Authenticate{}).Get("/api/connectors/{id}/tables", connectorController.Tables)
	facades.Route().Middleware(&middleware.Authenticate{}).Get("/api/connectors/{id}/columns", connectorController.Columns)

	destinationController := controllers.NewDestinationController()
	facades.Route().Middleware(&middleware.Authenticate{}).Get("/api/destinations", destinationController.Index)
	facades.Route().Middleware(&middleware.Authenticate{}).Post("/api/destinations", destinationController.Store)
	facades.Route().Middleware(&middleware.Authenticate{}).Put("/api/destinations/{id}", destinationController.Update)
	facades.Route().Middleware(&middleware.Authenticate{}).Delete("/api/destinations/{id}", destinationController.Destroy)
	facades.Route().Middleware(&middleware.Authenticate{}).Post("/api/destinations/{id}/test", destinationController.Test)

	pipelineController := controllers.NewPipelineController()
	facades.Route().Middleware(&middleware.Authenticate{}).Get("/api/pipelines", pipelineController.Index)
	facades.Route().Middleware(&middleware.Authenticate{}).Post("/api/pipelines", pipelineController.Store)
	facades.Route().Middleware(&middleware.Authenticate{}).Put("/api/pipelines/{id}", pipelineController.Update)
	facades.Route().Middleware(&middleware.Authenticate{}).Delete("/api/pipelines/{id}", pipelineController.Destroy)

	userController := controllers.NewUserController()
	facades.Route().Get("/users", userController.Index)
	facades.Route().Get("/favicon.ico", func(ctx http.Context) http.Response {
		return ctx.Response().NoContent()
	})

	Telego()
}
