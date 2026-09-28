package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	stdhttp "net/http"
	"net/url"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/contracts/http/client"

	"goravel/app/facades"
	"goravel/app/telego"
)

type TelegoController struct {
	store *telego.Store
}

type telegoCategory struct {
	Type  string
	Title string
}

type telegoHttpClientRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

var telegoCategories = map[string]telegoCategory{
	"requests":    {Type: "request", Title: "Requests"},
	"exceptions":  {Type: "exception", Title: "Exceptions"},
	"queries":     {Type: "query", Title: "Queries"},
	"http-client": {Type: "http", Title: "HTTP Client"},
	"cache":       {Type: "cache", Title: "Cache"},
	"events":      {Type: "event", Title: "Events"},
	"logs":        {Type: "log", Title: "Logs"},
	"traces":      {Type: "trace", Title: "Traces"},
}

func NewTelegoController() *TelegoController {
	return &TelegoController{
		store: telego.DefaultStore(),
	}
}

func (r *TelegoController) Root(ctx http.Context) http.Response {
	return ctx.Response().Redirect(stdhttp.StatusFound, "/telego/requests")
}

func (r *TelegoController) Index(ctx http.Context) http.Response {
	categorySlug := ctx.Request().Route("category")
	category, found := telegoCategories[categorySlug]
	if !found {
		return ctx.Response().String(stdhttp.StatusNotFound, "Telego category not found")
	}

	return ctx.Response().View().Make("telego_list.tmpl", map[string]any{
		"category": categorySlug,
		"title":    category.Title,
		"type":     category.Type,
	})
}

func (r *TelegoController) Detail(ctx http.Context) http.Response {
	categorySlug := ctx.Request().Route("category")
	category, categoryFound := telegoCategories[categorySlug]
	entry, entryFound := r.store.Get(ctx.Request().Route("id"))
	if !categoryFound || !entryFound || entry.Type != category.Type {
		return ctx.Response().String(stdhttp.StatusNotFound, "Telego entry not found")
	}

	return ctx.Response().View().Make("telego_detail.tmpl", map[string]any{
		"category": categorySlug,
		"entry":    entry,
		"title":    category.Title,
		"type":     category.Type,
	})
}

func (r *TelegoController) HttpClientTestPage(ctx http.Context) http.Response {
	return ctx.Response().View().Make("telego_http_client_test.tmpl")
}

func (r *TelegoController) HttpClientTest(ctx http.Context) http.Response {
	var input telegoHttpClientRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return ctx.Response().Json(stdhttp.StatusBadRequest, http.Json{"message": "Invalid request payload"})
	}

	parsedURL, err := url.ParseRequestURI(strings.TrimSpace(input.URL))
	if err != nil || parsedURL.Host == "" || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return ctx.Response().Json(stdhttp.StatusUnprocessableEntity, http.Json{"message": "URL must use http or https"})
	}

	method := strings.ToUpper(strings.TrimSpace(input.Method))
	request := facades.Http().WithContext(ctx.Context()).WithHeaders(input.Headers)
	started := time.Now()
	response, err := sendTelegoHttpRequest(request, method, parsedURL.String(), input.Body)
	duration := time.Since(started)
	if err != nil {
		_ = facades.Telemetry().ForceFlush(ctx.Context())
		entryID := r.latestHttpEntry(started)
		r.enrichHttpEntry(entryID, input, "", nil)
		return ctx.Response().Json(stdhttp.StatusBadGateway, http.Json{
			"message":     err.Error(),
			"duration_ms": float64(duration.Microseconds()) / 1000,
			"entry_id":    entryID,
		})
	}

	body, err := response.Body()
	if err != nil {
		return ctx.Response().Json(stdhttp.StatusBadGateway, http.Json{"message": err.Error()})
	}
	if len(body) > 256*1024 {
		body = body[:256*1024] + "\n... [response truncated]"
	}

	_ = facades.Telemetry().ForceFlush(ctx.Context())
	entryID := r.latestHttpEntry(started)
	r.enrichHttpEntry(entryID, input, body, response.Headers())

	return ctx.Response().Success().Json(http.Json{
		"body":         body,
		"content_type": response.Header("Content-Type"),
		"duration_ms":  float64(duration.Microseconds()) / 1000,
		"entry_id":     entryID,
		"headers":      telego.Redact(map[string][]string(response.Headers())),
		"status":       response.Status(),
		"url":          parsedURL.String(),
	})
}

func sendTelegoHttpRequest(request client.Request, method, targetURL, body string) (client.Response, error) {
	payload := func() *bytes.Reader { return bytes.NewReader([]byte(body)) }

	switch method {
	case stdhttp.MethodGet:
		return request.Get(targetURL)
	case stdhttp.MethodPost:
		return request.Post(targetURL, payload())
	case stdhttp.MethodPut:
		return request.Put(targetURL, payload())
	case stdhttp.MethodPatch:
		return request.Patch(targetURL, payload())
	case stdhttp.MethodDelete:
		return request.Delete(targetURL, payload())
	case stdhttp.MethodHead:
		return request.Head(targetURL)
	case stdhttp.MethodOptions:
		return request.Options(targetURL)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}
}

func (r *TelegoController) latestHttpEntry(started time.Time) string {
	for _, entry := range r.store.List(telego.Filters{Type: "http", Limit: 10}) {
		if !entry.Timestamp.Before(started.Add(-time.Second)) {
			return entry.ID
		}
	}

	return ""
}

func (r *TelegoController) enrichHttpEntry(entryID string, input telegoHttpClientRequest, responseBody string, responseHeaders stdhttp.Header) {
	if entryID == "" {
		return
	}

	r.store.Update(entryID, func(entry *telego.Entry) {
		payload := cloneMap(entry.Payload)
		attributes, _ := payload["attributes"].(map[string]any)
		attributes = cloneMap(attributes)
		attributes["telego.request.payload"] = capturedBody(input.Body)
		attributes["telego.request.headers"] = telego.Redact(input.Headers)
		if responseHeaders != nil {
			attributes["telego.response.body"] = capturedBody(responseBody)
			attributes["telego.response.headers"] = telego.Redact(map[string][]string(responseHeaders))
		}
		payload["attributes"] = attributes
		entry.Payload = payload
	})
}

func capturedBody(body string) any {
	if strings.TrimSpace(body) == "" {
		return map[string]any{}
	}

	var value any
	if json.Unmarshal([]byte(body), &value) == nil {
		return value
	}

	return body
}

func cloneMap(source map[string]any) map[string]any {
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}

	return cloned
}

func (r *TelegoController) Entries(ctx http.Context) http.Response {
	entries := r.store.List(telego.Filters{
		Type:   ctx.Request().Query("type"),
		Search: ctx.Request().Query("search"),
		Limit:  ctx.Request().QueryInt("limit", 100),
	})

	return ctx.Response().Success().Json(http.Json{
		"data":   entries,
		"stats":  r.store.Stats(),
		"paused": r.store.Paused(),
	})
}

func (r *TelegoController) Show(ctx http.Context) http.Response {
	entry, found := r.store.Get(ctx.Request().Route("id"))
	if !found {
		return ctx.Response().Json(stdhttp.StatusNotFound, http.Json{"message": "Entry not found"})
	}

	return ctx.Response().Success().Json(http.Json{"data": entry})
}

func (r *TelegoController) Clear(ctx http.Context) http.Response {
	r.store.Clear()

	return ctx.Response().NoContent()
}

func (r *TelegoController) Pause(ctx http.Context) http.Response {
	paused := ctx.Request().InputBool("paused", !r.store.Paused())
	r.store.SetPaused(paused)

	return ctx.Response().Success().Json(http.Json{"paused": paused})
}
