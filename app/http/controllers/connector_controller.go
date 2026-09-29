package controllers

import (
	"fmt"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/models"
)

type connectorRequest struct {
	Name     string `form:"name" json:"name"`
	Kind     string `form:"kind" json:"kind"`
	Host     string `form:"host" json:"host"`
	Port     *int   `form:"port" json:"port"`
	Database string `form:"database" json:"database"`
}

type connectorResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Detail   string `json:"detail"`
	Status   string `json:"status"`
	LastSync string `json:"lastSync"`
}

type ConnectorController struct{}

func NewConnectorController() *ConnectorController {
	return &ConnectorController{}
}

func (r *ConnectorController) Index(ctx http.Context) http.Response {
	user, response := authenticatedUser(ctx)
	if response != nil {
		return response
	}

	var connectors []models.Connector
	if err := facades.Orm().Query().Where("user_id", user.ID).OrderByDesc("id").Find(&connectors); err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível carregar os conectores"})
	}

	data := make([]connectorResponse, 0, len(connectors))
	for _, connector := range connectors {
		data = append(data, presentConnector(connector))
	}

	return ctx.Response().Success().Json(http.Json{"data": data})
}

func (r *ConnectorController) Store(ctx http.Context) http.Response {
	user, response := authenticatedUser(ctx)
	if response != nil {
		return response
	}

	var input connectorRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return ctx.Response().Json(stdhttp.StatusBadRequest, http.Json{"message": "Dados do conector inválidos"})
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Kind = strings.ToLower(strings.TrimSpace(input.Kind))
	input.Host = strings.TrimSpace(input.Host)
	input.Database = strings.TrimSpace(input.Database)
	if message := validateConnector(input); message != "" {
		return ctx.Response().Json(stdhttp.StatusUnprocessableEntity, http.Json{"message": message})
	}

	connector := models.Connector{
		UserID:       user.ID,
		Name:         input.Name,
		Kind:         input.Kind,
		Status:       "configured",
		Host:         input.Host,
		Port:         input.Port,
		DatabaseName: input.Database,
	}
	if err := facades.Orm().Query().Create(&connector); err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível salvar o conector"})
	}

	return ctx.Response().Json(stdhttp.StatusCreated, http.Json{"data": presentConnector(connector)})
}

func authenticatedUser(ctx http.Context) (models.User, http.Response) {
	var user models.User
	if err := facades.Auth(ctx).User(&user); err != nil {
		return user, ctx.Response().Json(stdhttp.StatusUnauthorized, http.Json{"message": "Sessão inválida"})
	}

	return user, nil
}

func validateConnector(input connectorRequest) string {
	validKinds := map[string]bool{"bigquery": true, "mysql": true, "postgresql": true, "sqlserver": true, "s3": true, "rest": true}
	if len(input.Name) < 2 {
		return "Informe um nome para o conector"
	}
	if !validKinds[input.Kind] {
		return "Tipo de conector não suportado"
	}
	if input.Host == "" {
		return "Informe o host ou endpoint da origem"
	}
	if input.Port != nil && (*input.Port < 1 || *input.Port > 65535) {
		return "Informe uma porta válida"
	}
	if input.Kind != "rest" && input.Database == "" {
		return "Informe o banco, dataset ou bucket"
	}

	return ""
}

func presentConnector(connector models.Connector) connectorResponse {
	detail := connector.Host
	if connector.Port != nil {
		detail += fmt.Sprintf(":%d", *connector.Port)
	}
	if connector.DatabaseName != "" {
		detail += " / " + connector.DatabaseName
	}

	lastSync := "ainda não executado"
	if connector.LastSyncAt != nil {
		lastSync = connector.LastSyncAt.Format(time.DateTime)
	}

	return connectorResponse{ID: connector.ID, Name: connector.Name, Kind: connector.Kind, Detail: detail, Status: connector.Status, LastSync: lastSync}
}
