package controllers

import (
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/foundation"
	httpmocks "github.com/goravel/framework/mocks/http"
	frameworkmock "github.com/goravel/framework/testing/mock"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"

	"goravel/app/models"
)

func TestClickHouseConnectorDefaults(t *testing.T) {
	for _, test := range []struct {
		mode string
		port int
	}{
		{mode: "", port: 8123},
		{mode: "verify-full", port: 8443},
	} {
		t.Run(test.mode, func(t *testing.T) {
			input := connectorRequest{Name: " ClickHouse ", Kind: " ClickHouse ", Host: " localhost ", Database: " default ", SSLMode: test.mode}
			normalizeConnectorInput(&input)
			assert.Equal(t, "clickhouse", input.Kind)
			assert.Equal(t, test.port, *input.Port)
			assert.Equal(t, "default", input.Username)
			assert.Equal(t, 10, input.ConnectTimeout)
			assert.Empty(t, input.Schema)
			assert.Empty(t, validateConnector(input))
		})
	}
}

func TestClickHouseConnectorValidation(t *testing.T) {
	for _, test := range []struct {
		name   string
		modify func(*connectorRequest)
	}{
		{name: "SSL mode", modify: func(input *connectorRequest) { input.SSLMode = "prefer" }},
		{name: "host URL", modify: func(input *connectorRequest) { input.Host = "https://localhost" }},
		{name: "host port", modify: func(input *connectorRequest) { input.Host = "localhost:8123" }},
		{name: "missing database", modify: func(input *connectorRequest) { input.Database = "" }},
		{name: "timeout", modify: func(input *connectorRequest) { input.ConnectTimeout = 61 }},
		{name: "certificate pair", modify: func(input *connectorRequest) { input.SSLMode = "verify-full"; input.SSLCertificate = "certificate" }},
		{name: "HTTP certificate", modify: func(input *connectorRequest) { input.SSLRootCert = "certificate" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := connectorRequest{Name: "ClickHouse", Kind: "clickhouse", Host: "localhost", Database: "default"}
			normalizeConnectorInput(&input)
			test.modify(&input)
			assert.NotEmpty(t, validateConnector(input))
		})
	}
}

func TestPostgreSQLConnectorDefaultsUnchanged(t *testing.T) {
	input := connectorRequest{Name: "PostgreSQL", Kind: "postgresql", Host: "localhost", Database: "analytics", Username: "postgres"}
	normalizeConnectorInput(&input)
	assert.Equal(t, "public", input.Schema)
	assert.Equal(t, "prefer", input.SSLMode)
	assert.Equal(t, 10, input.ConnectTimeout)
	assert.Empty(t, validateConnector(input))
}

func TestConnectorControllerPurpose(t *testing.T) {
	assert.Equal(t, "source", NewConnectorController().purpose)
	assert.Equal(t, "destination", NewDestinationController().purpose)
}

func TestOwnedConnectorPurposeScope(t *testing.T) {
	for _, purpose := range []string{"source", "destination"} {
		for _, found := range []bool{true, false} {
			name := purpose + "/not-found"
			if found {
				name = purpose + "/found"
			}
			t.Run(name, func(t *testing.T) {
				previousApp := foundation.App
				t.Cleanup(func() { foundation.App = previousApp })
				factory := frameworkmock.Factory()
				ormMock, query := factory.Orm(), factory.OrmQuery()
				ctx, request := httpmocks.NewContext(t), httpmocks.NewContextRequest(t)
				t.Cleanup(func() {
					ormMock.AssertExpectations(t)
					query.AssertExpectations(t)
					factory.App().AssertExpectations(t)
				})
				ctx.EXPECT().Request().Return(request).Once()
				request.EXPECT().Route("id").Return("42").Once()
				ormMock.EXPECT().Query().Return(query).Once()
				query.EXPECT().Where("id", uint64(42)).Return(query).Once()
				query.EXPECT().Where("user_id", uint(7)).Return(query).Once()
				query.EXPECT().Where("purpose", purpose).Return(query).Once()
				want := models.Connector{UserID: 7, Purpose: purpose}
				query.EXPECT().FindOrFail(testifymock.IsType(&models.Connector{})).
					RunAndReturn(func(dest any, conditions ...any) error {
						if !found {
							return assert.AnError
						}
						*dest.(*models.Connector) = want
						return nil
					}).Once()
				var expectedResponse contractshttp.Response
				if !found {
					response := httpmocks.NewContextResponse(t)
					result := httpmocks.NewAbortableResponse(t)
					ctx.EXPECT().Response().Return(response).Once()
					response.EXPECT().Json(404, contractshttp.Json{"message": "Conector não encontrado"}).Return(result).Once()
					expectedResponse = result
				}
				controller := NewConnectorController()
				if purpose == "destination" {
					controller = NewDestinationController()
				}
				got, response := controller.ownedConnector(ctx, uint(7))
				assert.Equal(t, expectedResponse, response)
				if found {
					assert.Equal(t, want, got)
				}
			})
		}
	}
}
