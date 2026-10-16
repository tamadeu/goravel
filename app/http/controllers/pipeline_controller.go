package controllers

import (
	"encoding/json"
	"fmt"
	stdhttp "net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services"
)

type pipelineRequest struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	SourceID         uint     `json:"sourceId"`
	DestinationID    uint     `json:"destinationId"`
	SourceSchema     string   `json:"sourceSchema"`
	SourceTable      string   `json:"sourceTable"`
	DestinationTable string   `json:"destinationTable"`
	Columns          []string `json:"columns"`
	Frequency        string   `json:"frequency"`
}

type pipelineEndpointResponse struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	Status string `json:"status"`
}

type pipelineResponse struct {
	ID               uint                     `json:"id"`
	Name             string                   `json:"name"`
	Description      string                   `json:"description"`
	Source           pipelineEndpointResponse `json:"source"`
	Destination      pipelineEndpointResponse `json:"destination"`
	SourceSchema     string                   `json:"sourceSchema"`
	SourceTable      string                   `json:"sourceTable"`
	DestinationTable string                   `json:"destinationTable"`
	Columns          []string                 `json:"columns"`
	Frequency        string                   `json:"frequency"`
}

var destinationTablePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type PipelineController struct{}

func NewPipelineController() *PipelineController {
	return &PipelineController{}
}

func (r *PipelineController) Index(ctx http.Context) http.Response {
	user, response := authenticatedUser(ctx)
	if response != nil {
		return response
	}
	var pipelines []models.Pipeline
	if err := facades.Orm().Query().Where("user_id", user.ID).OrderByDesc("id").Find(&pipelines); err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível carregar os pipelines"})
	}
	data := make([]pipelineResponse, 0, len(pipelines))
	if len(pipelines) == 0 {
		return ctx.Response().Success().Json(http.Json{"data": data})
	}
	endpointIDs := make([]uint, 0, len(pipelines)*2)
	for _, pipeline := range pipelines {
		endpointIDs = append(endpointIDs, pipeline.SourceID, pipeline.DestinationID)
	}
	var connectors []models.Connector
	if err := facades.Orm().Query().Where("user_id", user.ID).Where("id IN ?", endpointIDs).Find(&connectors); err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível carregar as conexões dos pipelines"})
	}
	endpoints := make(map[uint]models.Connector, len(connectors))
	for _, connector := range connectors {
		endpoints[connector.ID] = connector
	}
	for _, pipeline := range pipelines {
		data = append(data, presentPipeline(pipeline, endpoints[pipeline.SourceID], endpoints[pipeline.DestinationID]))
	}
	return ctx.Response().Success().Json(http.Json{"data": data})
}

func (r *PipelineController) Store(ctx http.Context) http.Response {
	user, response := authenticatedUser(ctx)
	if response != nil {
		return response
	}
	input, response := bindPipelineInput(ctx)
	if response != nil {
		return response
	}
	source, destination, message := resolvePipelineEndpoints(user.ID, input)
	if message != "" {
		return ctx.Response().Json(stdhttp.StatusUnprocessableEntity, http.Json{"message": message})
	}
	if message := validateSourceSelection(ctx, source, input); message != "" {
		return ctx.Response().Json(stdhttp.StatusUnprocessableEntity, http.Json{"message": message})
	}
	columnsJSON, _ := json.Marshal(input.Columns)
	pipeline := models.Pipeline{
		UserID: user.ID, Name: input.Name, Description: input.Description, SourceID: source.ID, DestinationID: destination.ID,
		SourceSchema: input.SourceSchema, SourceTable: input.SourceTable, DestinationTable: input.DestinationTable, ColumnsJSON: string(columnsJSON), Frequency: input.Frequency,
	}
	if err := facades.Orm().Query().Create(&pipeline); err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível salvar o pipeline"})
	}
	return ctx.Response().Json(stdhttp.StatusCreated, http.Json{"data": presentPipeline(pipeline, source, destination)})
}

func (r *PipelineController) Update(ctx http.Context) http.Response {
	user, response := authenticatedUser(ctx)
	if response != nil {
		return response
	}
	pipeline, response := ownedPipeline(ctx, user.ID)
	if response != nil {
		return response
	}
	input, response := bindPipelineInput(ctx)
	if response != nil {
		return response
	}
	source, destination, message := resolvePipelineEndpoints(user.ID, input)
	if message != "" {
		return ctx.Response().Json(stdhttp.StatusUnprocessableEntity, http.Json{"message": message})
	}
	pipeline.Name, pipeline.Description = input.Name, input.Description
	if message := validateSourceSelection(ctx, source, input); message != "" {
		return ctx.Response().Json(stdhttp.StatusUnprocessableEntity, http.Json{"message": message})
	}
	columnsJSON, _ := json.Marshal(input.Columns)
	pipeline.SourceID, pipeline.DestinationID = source.ID, destination.ID
	pipeline.SourceSchema, pipeline.SourceTable = input.SourceSchema, input.SourceTable
	pipeline.DestinationTable, pipeline.ColumnsJSON, pipeline.Frequency = input.DestinationTable, string(columnsJSON), input.Frequency
	if err := facades.Orm().Query().Save(&pipeline); err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível atualizar o pipeline"})
	}
	return ctx.Response().Success().Json(http.Json{"data": presentPipeline(pipeline, source, destination)})
}

func (r *PipelineController) Destroy(ctx http.Context) http.Response {
	user, response := authenticatedUser(ctx)
	if response != nil {
		return response
	}
	pipeline, response := ownedPipeline(ctx, user.ID)
	if response != nil {
		return response
	}
	if _, err := facades.Orm().Query().Delete(&pipeline); err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível excluir o pipeline"})
	}
	return ctx.Response().NoContent()
}

func bindPipelineInput(ctx http.Context) (pipelineRequest, http.Response) {
	var input pipelineRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return input, ctx.Response().Json(stdhttp.StatusBadRequest, http.Json{"message": "Dados do pipeline inválidos"})
	}
	input.Name, input.Description = strings.TrimSpace(input.Name), strings.TrimSpace(input.Description)
	input.DestinationTable = strings.TrimSpace(input.DestinationTable)
	input.Frequency = strings.ToLower(strings.TrimSpace(input.Frequency))
	if message := validatePipeline(input); message != "" {
		return input, ctx.Response().Json(stdhttp.StatusUnprocessableEntity, http.Json{"message": message})
	}
	return input, nil
}

func validatePipeline(input pipelineRequest) string {
	if length := utf8.RuneCountInString(input.Name); length < 2 || length > 120 {
		return "Informe um nome entre 2 e 120 caracteres"
	}
	if utf8.RuneCountInString(input.Description) > 2000 {
		return "A descrição deve ter no máximo 2000 caracteres"
	}
	if input.SourceID == 0 {
		return "Selecione um conector de origem"
	}
	if input.DestinationID == 0 {
		return "Selecione um destino"
	}
	if input.SourceID == input.DestinationID {
		return "Origem e destino devem ser cadastros diferentes"
	}
	if input.SourceSchema == "" || len(input.SourceSchema) > 255 || input.SourceTable == "" || len(input.SourceTable) > 255 {
		return "Selecione uma tabela válida da origem"
	}
	if len(input.DestinationTable) > 63 || !destinationTablePattern.MatchString(input.DestinationTable) {
		return "Informe uma tabela de destino com até 63 caracteres, usando letras, números e underscores, sem iniciar por número"
	}
	if len(input.Columns) == 0 || len(input.Columns) > 1000 {
		return "Selecione entre 1 e 1000 colunas da origem"
	}
	seen := make(map[string]bool, len(input.Columns))
	for _, column := range input.Columns {
		if strings.TrimSpace(column) == "" || len(column) > 255 || seen[column] {
			return "Selecione colunas válidas, sem duplicações"
		}
		seen[column] = true
	}
	frequencies := map[string]bool{"manual": true, "every_5_minutes": true, "every_15_minutes": true, "hourly": true, "daily": true}
	if !frequencies[input.Frequency] {
		return "Selecione uma frequência válida para o pipeline"
	}
	return ""
}

func validateSourceSelection(ctx http.Context, source models.Connector, input pipelineRequest) string {
	columns, err := loadSourceColumns(ctx.Context(), source, input.SourceSchema, input.SourceTable)
	if err != nil {
		return err.Error()
	}
	return validatePipelineColumns(input.Columns, columns)
}

func validatePipelineColumns(selected []string, available []services.SourceColumn) string {
	if len(available) == 0 {
		return "A tabela não foi encontrada ou não possui colunas acessíveis"
	}
	names := make(map[string]bool, len(available))
	for _, column := range available {
		names[column.Name] = true
	}
	for _, name := range selected {
		if !names[name] {
			return fmt.Sprintf("A coluna %q não está disponível na tabela da origem", name)
		}
	}
	return ""
}

func resolvePipelineEndpoints(userID uint, input pipelineRequest) (models.Connector, models.Connector, string) {
	var source, destination models.Connector
	if err := facades.Orm().Query().Where("id", input.SourceID).Where("user_id", userID).Where("purpose", "source").FindOrFail(&source); err != nil {
		return source, destination, "Conector de origem não encontrado"
	}
	if err := facades.Orm().Query().Where("id", input.DestinationID).Where("user_id", userID).Where("purpose", "destination").FindOrFail(&destination); err != nil {
		return source, destination, "Destino não encontrado"
	}
	return source, destination, ""
}

func ownedPipeline(ctx http.Context, userID uint) (models.Pipeline, http.Response) {
	pipelineID, err := strconv.ParseUint(ctx.Request().Route("id"), 10, 64)
	if err != nil || pipelineID == 0 {
		return models.Pipeline{}, ctx.Response().Json(stdhttp.StatusBadRequest, http.Json{"message": "Pipeline inválido"})
	}
	var pipeline models.Pipeline
	if err := facades.Orm().Query().Where("id", pipelineID).Where("user_id", userID).FindOrFail(&pipeline); err != nil {
		return pipeline, ctx.Response().Json(stdhttp.StatusNotFound, http.Json{"message": "Pipeline não encontrado"})
	}
	return pipeline, nil
}

func presentPipeline(pipeline models.Pipeline, source, destination models.Connector) pipelineResponse {
	columns := make([]string, 0)
	if pipeline.ColumnsJSON != "" {
		if err := json.Unmarshal([]byte(pipeline.ColumnsJSON), &columns); err != nil || columns == nil {
			columns = make([]string, 0)
		}
	}
	frequency := pipeline.Frequency
	if frequency == "" {
		frequency = "manual"
	}
	return pipelineResponse{
		ID: pipeline.ID, Name: pipeline.Name, Description: pipeline.Description,
		SourceSchema: pipeline.SourceSchema, SourceTable: pipeline.SourceTable, DestinationTable: pipeline.DestinationTable, Columns: columns, Frequency: frequency,
		Source:      pipelineEndpointResponse{ID: source.ID, Name: source.Name, Kind: source.Kind, Status: source.Status},
		Destination: pipelineEndpointResponse{ID: destination.ID, Name: destination.Name, Kind: destination.Kind, Status: destination.Status},
	}
}
