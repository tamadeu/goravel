package bootstrap

import (
	"strings"

	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/contracts/foundation/configuration"
	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/foundation"
	telemetryhttp "github.com/goravel/framework/telemetry/instrumentation/http"

	"goravel/app/http/middleware"
	"goravel/config"
	"goravel/routes"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithMigrations(Migrations).
		WithRouting(func() {
			routes.Web()
			routes.Grpc()
		}).
		WithProviders(Providers).
		WithConfig(config.Boot).
		WithMiddleware(func(handler configuration.Middleware) {
			handler.Append(
				telemetryhttp.Telemetry(telemetryhttp.WithFilter(func(ctx http.Context) bool {
					path := ctx.Request().Path()
					return !strings.HasPrefix(path, "/telego") && path != "/favicon.ico"
				})),
				&middleware.Telego{},
			)
		}).
		Create()
}
