package facades

import (
	"github.com/goravel/framework/contracts/event"

	"goravel/app/telego"
)

func Event() event.Instance {
	return telego.ObserveEvents(App().MakeEvent(), telego.DefaultStore())
}
