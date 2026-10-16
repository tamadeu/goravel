package controllers

import (
	"encoding/json"
	"strings"
	"testing"

	contractshttp "github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/foundation"
	httpmocks "github.com/goravel/framework/mocks/http"
	frameworkmock "github.com/goravel/framework/testing/mock"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"goravel/app/models"
	"goravel/app/services"
)

func TestValidatePipeline(t *testing.T) {
	for _, test := range []struct {
		name   string
		modify func(*pipelineRequest)
	}{
		{name: "short name", modify: func(input *pipelineRequest) { input.Name = "a" }},
		{name: "long name", modify: func(input *pipelineRequest) { input.Name = strings.Repeat("a", 121) }},
		{name: "long description", modify: func(input *pipelineRequest) { input.Description = strings.Repeat("a", 2001) }},
		{name: "missing source", modify: func(input *pipelineRequest) { input.SourceID = 0 }},
		{name: "missing destination", modify: func(input *pipelineRequest) { input.DestinationID = 0 }},
		{name: "same endpoint", modify: func(input *pipelineRequest) { input.DestinationID = input.SourceID }},
		{name: "missing table", modify: func(input *pipelineRequest) { input.SourceTable = "" }},
		{name: "unsafe destination table", modify: func(input *pipelineRequest) { input.DestinationTable = "orders; DROP TABLE users" }},
		{name: "no columns", modify: func(input *pipelineRequest) { input.Columns = nil }},
		{name: "duplicate columns", modify: func(input *pipelineRequest) { input.Columns = []string{"id", "id"} }},
		{name: "unknown frequency", modify: func(input *pipelineRequest) { input.Frequency = "invalid" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			input := pipelineRequest{Name: "Analytics", SourceID: 11, DestinationID: 12, SourceSchema: "public", SourceTable: "orders", DestinationTable: "orders_copy", Columns: []string{"id"}, Frequency: "manual"}
			test.modify(&input)
			assert.NotEmpty(t, validatePipeline(input))
		})
	}
	assert.Empty(t, validatePipeline(pipelineRequest{Name: "Analytics", SourceID: 11, DestinationID: 12, SourceSchema: "public", SourceTable: "orders", DestinationTable: "orders_copy", Columns: []string{"id"}, Frequency: "manual"}))
}

func TestPipelineColumnSelection(t *testing.T) {
	columns := []services.SourceColumn{{Name: "id"}, {Name: "total"}}
	assert.Empty(t, validatePipelineColumns([]string{"id"}, columns))
	assert.NotEmpty(t, validatePipelineColumns([]string{"missing"}, columns))
	assert.NotEmpty(t, validatePipelineColumns([]string{"id"}, nil))
}

func TestLegacyPipelineWizardDefaults(t *testing.T) {
	response := presentPipeline(models.Pipeline{}, models.Connector{}, models.Connector{})
	assert.Equal(t, "manual", response.Frequency)
	assert.NotNil(t, response.Columns)
	assert.Empty(t, response.Columns)
}

func TestResolvePipelineEndpointsScope(t *testing.T) {
	for _, failure := range []string{"", "source", "destination"} {
		t.Run("failure="+failure, func(t *testing.T) {
			previousApp := foundation.App
			t.Cleanup(func() { foundation.App = previousApp })
			factory := frameworkmock.Factory()
			ormMock, query := factory.Orm(), factory.OrmQuery()
			t.Cleanup(func() {
				ormMock.AssertExpectations(t)
				query.AssertExpectations(t)
				factory.App().AssertExpectations(t)
			})
			queryCount := 2
			if failure == "source" {
				queryCount = 1
			}
			ormMock.EXPECT().Query().Return(query).Times(queryCount)
			query.EXPECT().Where("user_id", uint(7)).Return(query).Times(queryCount)
			query.EXPECT().Where("id", uint(11)).Return(query).Once()
			query.EXPECT().Where("purpose", "source").Return(query).Once()
			query.EXPECT().FindOrFail(testifymock.IsType(&models.Connector{})).RunAndReturn(func(target any, conditions ...any) error {
				if failure == "source" {
					return assert.AnError
				}
				connector := target.(*models.Connector)
				connector.ID, connector.UserID, connector.Purpose = 11, 7, "source"
				return nil
			}).Once()
			if failure != "source" {
				query.EXPECT().Where("id", uint(12)).Return(query).Once()
				query.EXPECT().Where("purpose", "destination").Return(query).Once()
				query.EXPECT().FindOrFail(testifymock.IsType(&models.Connector{})).RunAndReturn(func(target any, conditions ...any) error {
					if failure == "destination" {
						return assert.AnError
					}
					connector := target.(*models.Connector)
					connector.ID, connector.UserID, connector.Purpose = 12, 7, "destination"
					return nil
				}).Once()
			}
			source, destination, message := resolvePipelineEndpoints(7, pipelineRequest{SourceID: 11, DestinationID: 12})
			if failure == "" {
				assert.Empty(t, message)
				assert.Equal(t, uint(11), source.ID)
				assert.Equal(t, uint(12), destination.ID)
			} else {
				assert.NotEmpty(t, message)
			}
		})
	}
}

func TestOwnedPipelineScope(t *testing.T) {
	for _, found := range []bool{true, false} {
		name := "not-found"
		if found {
			name = "found"
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
			query.EXPECT().FindOrFail(testifymock.IsType(&models.Pipeline{})).RunAndReturn(func(target any, conditions ...any) error {
				if !found {
					return assert.AnError
				}
				pipeline := target.(*models.Pipeline)
				pipeline.ID, pipeline.UserID = 42, 7
				return nil
			}).Once()
			var expectedResponse contractshttp.Response
			if !found {
				response := httpmocks.NewContextResponse(t)
				result := httpmocks.NewAbortableResponse(t)
				ctx.EXPECT().Response().Return(response).Once()
				response.EXPECT().Json(404, contractshttp.Json{"message": "Pipeline não encontrado"}).Return(result).Once()
				expectedResponse = result
			}
			pipeline, response := ownedPipeline(ctx, 7)
			assert.Equal(t, expectedResponse, response)
			if found {
				assert.Equal(t, uint(42), pipeline.ID)
			}
		})
	}
}

func TestPresentPipelineExcludesCredentials(t *testing.T) {
	endpoint := models.Connector{Name: "Analytics", Kind: "postgresql", PasswordEncrypted: "private-secret", Host: "private-host", Username: "private-user"}
	data, err := json.Marshal(presentPipeline(models.Pipeline{Name: "Sync"}, endpoint, endpoint))
	require.NoError(t, err)
	assert.NotContains(t, string(data), "private-secret")
	assert.NotContains(t, string(data), "private-host")
	assert.NotContains(t, string(data), "private-user")
}
