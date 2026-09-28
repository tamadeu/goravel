package feature

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/app/telego"
	"goravel/tests"
)

type TelegoTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestTelegoTestSuite(t *testing.T) {
	suite.Run(t, new(TelegoTestSuite))
}

func (s *TelegoTestSuite) SetupTest() {
	telego.SetEnabled(true)
	telego.DefaultStore().SetPaused(false)
	telego.DefaultStore().Clear()
}

func (s *TelegoTestSuite) TearDownTest() {
	telego.DefaultStore().Clear()
}

func (s *TelegoTestSuite) TestDashboard() {
	response, err := s.Http(s.T()).Get("/telego")
	s.Require().NoError(err)
	response.AssertFound().AssertHeader("Location", "/telego/requests")

	response, err = s.Http(s.T()).Get("/telego/requests")
	s.Require().NoError(err)
	response.AssertOk().AssertSee([]string{"Telego", "Requests", "/telego/assets/telego.css"})
}

func (s *TelegoTestSuite) TestDedicatedEntryPage() {
	entry := telego.DefaultStore().Add(telego.Entry{Type: "request", Name: "GET /users", Method: "GET", URI: "/users"})

	response, err := s.Http(s.T()).Get("/telego/requests/" + entry.ID)
	s.Require().NoError(err)
	response.AssertOk().AssertSee([]string{"GET /users", "Overview", "/telego/assets/telego-detail.js"})

	response, err = s.Http(s.T()).Get("/telego/queries/" + entry.ID)
	s.Require().NoError(err)
	response.AssertNotFound()
}

func (s *TelegoTestSuite) TestEntriesAPI() {
	telego.DefaultStore().Add(telego.Entry{Type: "cache", Name: "cache.GET", URI: "profile:1"})

	response, err := s.Http(s.T()).Get("/telego/api/entries?type=cache")

	s.Require().NoError(err)
	response.AssertOk().AssertJson(map[string]any{
		"paused": false,
		"stats": map[string]any{
			"all":   float64(1),
			"cache": float64(1),
		},
	})
}

func (s *TelegoTestSuite) TestHttpClientLabCapturesRequest() {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("X-Test-Source", "telego")
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"source":"http-client-lab","items":[1,2,3]}`))
	}))
	defer server.Close()

	page, err := s.Http(s.T()).Get("/telego/http-client/test")
	s.Require().NoError(err)
	page.AssertOk().AssertSee([]string{"Request lab", "/telego/assets/telego-http-client.js"})

	response, err := s.Http(s.T()).Post("/telego/api/http-client/test", strings.NewReader(`{
		"method":"GET",
		"url":"`+server.URL+`/profile",
		"headers":{"Accept":"application/json","Authorization":"Bearer secret"}
	}`))
	s.Require().NoError(err)
	response.AssertOk()

	var result struct {
		Body    string `json:"body"`
		EntryID string `json:"entry_id"`
		Status  int    `json:"status"`
	}
	s.Require().NoError(response.Bind(&result))
	s.Equal(http.StatusOK, result.Status)
	s.Contains(result.Body, "http-client-lab")
	s.NotEmpty(result.EntryID)

	entries := telego.DefaultStore().List(telego.Filters{Type: "http", Limit: 10})
	s.Require().Len(entries, 1)
	s.Equal("GET", entries[0].Method)
	s.Contains(entries[0].URI, "/profile")
	attributes := entries[0].Payload["attributes"].(map[string]any)
	requestHeaders := attributes["telego.request.headers"].(map[string]any)
	s.Equal("[REDACTED]", requestHeaders["Authorization"])
	responseBody := attributes["telego.response.body"].(map[string]any)
	s.Equal("http-client-lab", responseBody["source"])
	s.Len(responseBody["items"], 3)
	responseHeaders := attributes["telego.response.headers"].(map[string]any)
	s.Contains(responseHeaders, "X-Test-Source")
}

func (s *TelegoTestSuite) TestCapturesApplicationLog() {
	facades.Log().With(map[string]any{"request_id": "req-1"}).Info("telego integration log")
	s.Require().NoError(facades.Telemetry().ForceFlush(context.Background()))

	entries := telego.DefaultStore().List(telego.Filters{Type: "log", Limit: 10})
	s.Require().Len(entries, 1)
	s.Equal("telego integration log", entries[0].Name)
	s.Equal("INFO", entries[0].Method)
}

func (s *TelegoTestSuite) TestCapturesApplicationRequestAndExcludesDashboard() {
	response, err := s.Http(s.T()).Get("/users")
	s.Require().NoError(err)
	response.AssertOk()
	s.Require().NoError(facades.Telemetry().ForceFlush(context.Background()))

	requests := telego.DefaultStore().List(telego.Filters{Type: "request", Limit: 10})
	s.Require().Len(requests, 1)
	s.Equal("GET", requests[0].Method)
	s.Equal("/users", requests[0].URI)

	response, err = s.Http(s.T()).Get("/telego/requests")
	s.Require().NoError(err)
	response.AssertOk()
	s.Require().NoError(facades.Telemetry().ForceFlush(context.Background()))
	s.Len(telego.DefaultStore().List(telego.Filters{Type: "request", Limit: 10}), 1)
}

func (s *TelegoTestSuite) TestFaviconDoesNotGeneratePanic() {
	response, err := s.Http(s.T()).Get("/favicon.ico")

	s.Require().NoError(err)
	response.AssertNoContent()
	s.Require().NoError(facades.Telemetry().ForceFlush(context.Background()))
	s.Empty(telego.DefaultStore().List(telego.Filters{Type: "log", Limit: 10}))
}
