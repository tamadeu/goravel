package config

import (
	"goravel/app/facades"
	"goravel/app/telego"
)

func init() {
	config := facades.Config()
	enabled := config.EnvBool("TELOGO_ENABLED", config.EnvString("APP_ENV", "production") == "local")
	telego.SetEnabled(enabled)
	config.Add("telego", map[string]any{
		"enabled": enabled,
	})
}
