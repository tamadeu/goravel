package middleware

import (
	stdhttp "net/http"
	"strings"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
)

type Authenticate struct{}

func (m *Authenticate) Signature() string {
	return "Authenticate"
}

func (m *Authenticate) Handle(ctx http.Context) {
	token := strings.TrimSpace(ctx.Request().Header("Authorization"))
	if token == "" {
		_ = ctx.Response().Json(stdhttp.StatusUnauthorized, http.Json{"message": "Autenticação necessária"}).Abort()
		ctx.Request().Next()
		return
	}

	if _, err := facades.Auth(ctx).Parse(token); err != nil {
		_ = ctx.Response().Json(stdhttp.StatusUnauthorized, http.Json{"message": "Sessão inválida ou expirada"}).Abort()
		ctx.Request().Next()
		return
	}

	ctx.Request().Next()
}
