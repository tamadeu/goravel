package services

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClickHouseSourceMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		query, err := io.ReadAll(request.Body)
		require.NoError(t, err)
		username, password, ok := request.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "default", username)
		assert.Equal(t, "secret&password", password)
		assert.Equal(t, "analytics & reports", request.URL.Query().Get("param_source_database"))
		if strings.Contains(string(query), "system.tables") {
			_, _ = response.Write([]byte(`{"data":[{"schema":"analytics & reports","name":"orders"}]}`))
			return
		}
		assert.Contains(t, string(query), "{source_table:String}")
		assert.NotContains(t, string(query), "orders'; DROP TABLE users")
		assert.Equal(t, "orders'; DROP TABLE users", request.URL.Query().Get("param_source_table"))
		_, _ = response.Write([]byte(`{"data":[{"name":"id","type":"UInt64"},{"name":"email","type":"Nullable(String)"},{"name":"tags","type":"Array(Nullable(String))"}]}`))
	}))
	defer server.Close()
	config := clickHouseTestConfig(t, server)
	tables, err := ListClickHouseTables(context.Background(), config)
	require.NoError(t, err)
	assert.Equal(t, []SourceTable{{Schema: config.Database, Name: "orders"}}, tables)
	columns, err := ListClickHouseColumns(context.Background(), config, config.Database, "orders'; DROP TABLE users")
	require.NoError(t, err)
	assert.Equal(t, []SourceColumn{{Name: "id", Type: "UInt64"}, {Name: "email", Type: "Nullable(String)", Nullable: true}, {Name: "tags", Type: "Array(Nullable(String))"}}, columns)
}

func TestSourceMetadataRejectsOtherSchema(t *testing.T) {
	_, err := ListPostgreSQLColumns(context.Background(), PostgreSQLConfig{Schema: "public"}, "private", "users")
	require.Error(t, err)
	_, err = ListClickHouseColumns(context.Background(), ClickHouseConfig{Database: "default"}, "other_database", "users")
	require.Error(t, err)
}

func TestClickHouseEmptySourceMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		_, _ = response.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()
	config := clickHouseTestConfig(t, server)
	tables, err := ListClickHouseTables(context.Background(), config)
	require.NoError(t, err)
	assert.Empty(t, tables)
	columns, err := ListClickHouseColumns(context.Background(), config, config.Database, "missing")
	require.NoError(t, err)
	assert.Empty(t, columns)
}
