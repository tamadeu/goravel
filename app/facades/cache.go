package facades

import (
	"github.com/goravel/framework/contracts/cache"

	"goravel/app/telego"
)

func Cache() cache.Cache {
	return telego.ObserveCache(App().MakeCache(), telego.DefaultStore())
}
