package middleware

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goravel/framework/contracts/http"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"goravel/app/telego"
)

type Telego struct{}

func (m *Telego) Signature() string {
	return "telego"
}

func (m *Telego) Handle(ctx http.Context) {
	if strings.HasPrefix(ctx.Request().Path(), "/telego") {
		ctx.Request().Next()
		return
	}

	span := trace.SpanFromContext(ctx.Context())
	defer func() {
		if recovered := recover(); recovered != nil {
			span.RecordError(fmt.Errorf("panic: %v", recovered), trace.WithStackTrace(true))
			span.SetStatus(codes.Error, "Internal Server Error")
			panic(recovered)
		}
	}()

	ctx.Request().Next()

	span.SetAttributes(
		attribute.String("telego.handler", ctx.Request().Info().Handler),
		attribute.String("telego.client_ip", ctx.Request().Ip()),
		attribute.String("telego.request.headers", telego.JSON(map[string][]string(ctx.Request().Headers()), 16*1024)),
		attribute.String("telego.request.payload", telego.JSON(ctx.Request().All(), 16*1024)),
		attribute.String("telego.response.headers", telego.JSON(map[string][]string(ctx.Response().Origin().Header()), 16*1024)),
		attribute.String("telego.response.body", responseBody(ctx)),
	)
}

func responseBody(ctx http.Context) string {
	bodyBuffer := ctx.Response().Origin().Body()
	if bodyBuffer == nil {
		return ""
	}

	body := bodyBuffer.Bytes()
	if len(body) == 0 {
		return ""
	}
	if len(body) > 16*1024 {
		body = body[:16*1024]
	}

	var payload any
	if json.Unmarshal(body, &payload) == nil {
		return telego.JSON(payload, 16*1024)
	}

	return string(body)
}
