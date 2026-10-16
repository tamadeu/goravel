package services

import (
	"context"
	"encoding/pem"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func clickHouseTestConfig(t *testing.T, server *httptest.Server) ClickHouseConfig {
	t.Helper()
	endpoint, err := url.Parse(server.URL)
	require.NoError(t, err)
	host, port, err := net.SplitHostPort(endpoint.Host)
	require.NoError(t, err)
	portNumber, err := strconv.Atoi(port)
	require.NoError(t, err)
	return ClickHouseConfig{Host: host, Port: portNumber, Database: "analytics & reports", Username: "default", Password: "secret&password", SSLMode: "disable", ConnectTimeout: 1}
}

func TestClickHouseConnectionMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		assert.Equal(t, http.MethodPost, request.Method)
		assert.Equal(t, "analytics & reports", request.URL.Query().Get("database"))
		assert.Empty(t, request.URL.Query().Get("password"))
		username, password, ok := request.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "default", username)
		assert.Equal(t, "secret&password", password)
		query, err := io.ReadAll(request.Body)
		assert.NoError(t, err)
		assert.Equal(t, "SELECT currentDatabase() AS database, currentUser() AS user, version() AS version FORMAT JSON", string(query))
		_, _ = response.Write([]byte(`{"data":[{"database":"analytics","user":"default","version":"26.9.1"}]}`))
	}))
	defer server.Close()
	result, err := TestClickHouseConnection(context.Background(), clickHouseTestConfig(t, server))
	require.NoError(t, err)
	assert.Equal(t, "analytics", result.Database)
	assert.Equal(t, "default", result.User)
	assert.Equal(t, "26.9.1", result.Version)
}

func TestClickHouseConnectionRejectsResponses(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
	}{
		{name: "authentication", status: http.StatusUnauthorized, body: "secret&password"},
		{name: "redirect", status: http.StatusTemporaryRedirect, body: "secret&password"},
		{name: "invalid JSON", status: http.StatusOK, body: "invalid"},
		{name: "empty rows", status: http.StatusOK, body: `{"data":[]}`},
		{name: "missing metadata", status: http.StatusOK, body: `{"data":[{}]}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				response.Header().Set("Location", "/redirect")
				response.WriteHeader(test.status)
				_, _ = response.Write([]byte(test.body))
			}))
			defer server.Close()
			_, err := TestClickHouseConnection(context.Background(), clickHouseTestConfig(t, server))
			require.Error(t, err)
			assert.NotContains(t, err.Error(), "secret&password")
		})
	}
}

func TestClickHouseConnectionTLS(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		_, _ = response.Write([]byte(`{"data":[{"database":"default","user":"default","version":"26.9"}]}`))
	}))
	defer server.Close()
	config := clickHouseTestConfig(t, server)
	config.SSLMode = "verify-full"
	_, err := TestClickHouseConnection(context.Background(), config)
	require.Error(t, err)
	config.SSLRootCert = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}))
	_, err = TestClickHouseConnection(context.Background(), config)
	require.NoError(t, err)
}

func TestClickHouseConnectionTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		_, _ = io.Copy(io.Discard, request.Body)
		<-request.Context().Done()
	}))
	defer server.Close()
	_, err := TestClickHouseConnection(context.Background(), clickHouseTestConfig(t, server))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tempo limite")
}

func TestClickHouseTransportValidation(t *testing.T) {
	for _, test := range []struct {
		name   string
		modify func(*ClickHouseConfig)
	}{
		{name: "SSL mode", modify: func(config *ClickHouseConfig) { config.SSLMode = "require" }},
		{name: "CA", modify: func(config *ClickHouseConfig) { config.SSLMode = "verify-full"; config.SSLRootCert = "invalid" }},
		{name: "client certificate", modify: func(config *ClickHouseConfig) { config.SSLMode = "verify-full"; config.SSLCertificate = "invalid" }},
		{name: "HTTP certificates", modify: func(config *ClickHouseConfig) { config.SSLRootCert = "invalid" }},
		{name: "host URL", modify: func(config *ClickHouseConfig) { config.Host = "https://example.com" }},
		{name: "timeout", modify: func(config *ClickHouseConfig) { config.ConnectTimeout = 0 }},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := ClickHouseConfig{Host: "localhost", Port: 8123, Database: "default", Username: "default", SSLMode: "disable", ConnectTimeout: 10}
			test.modify(&config)
			_, err := clickHouseTransport(config)
			require.Error(t, err)
		})
	}
}
