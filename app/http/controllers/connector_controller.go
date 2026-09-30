package controllers

import (
	"fmt"
	stdhttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/services"
)

type connectorRequest struct {
	Name                      string `form:"name" json:"name"`
	Kind                      string `form:"kind" json:"kind"`
	Host                      string `form:"host" json:"host"`
	Port                      *int   `form:"port" json:"port"`
	Database                  string `form:"database" json:"database"`
	Username                  string `form:"username" json:"username"`
	Password                  string `form:"password" json:"password"`
	Schema                    string `form:"schema" json:"schema"`
	SSLMode                   string `form:"ssl_mode" json:"sslMode"`
	SSLRootCert               string `form:"ssl_root_cert" json:"sslRootCert"`
	SSLCertificate            string `form:"ssl_certificate" json:"sslCertificate"`
	SSLKey                    string `form:"ssl_key" json:"sslKey"`
	ConnectTimeout            int    `form:"connect_timeout" json:"connectTimeout"`
	ClearPassword             bool   `form:"clear_password" json:"clearPassword"`
	ClearSSLRootCert          bool   `form:"clear_ssl_root_cert" json:"clearSSLRootCert"`
	ClearSSLClientCertificate bool   `form:"clear_ssl_client_certificate" json:"clearSSLClientCertificate"`
}

type connectorResponse struct {
	ID            uint                           `json:"id"`
	Name          string                         `json:"name"`
	Kind          string                         `json:"kind"`
	Detail        string                         `json:"detail"`
	Status        string                         `json:"status"`
	LastSync      string                         `json:"lastSync"`
	Configuration connectorConfigurationResponse `json:"configuration"`
}

type connectorConfigurationResponse struct {
	Host                    string `json:"host"`
	Port                    *int   `json:"port"`
	Database                string `json:"database"`
	Username                string `json:"username"`
	Schema                  string `json:"schema"`
	SSLMode                 string `json:"sslMode"`
	ConnectTimeout          int    `json:"connectTimeout"`
	HasPassword             bool   `json:"hasPassword"`
	HasSSLRootCert          bool   `json:"hasSSLRootCert"`
	HasSSLClientCertificate bool   `json:"hasSSLClientCertificate"`
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
	input.Username = strings.TrimSpace(input.Username)
	input.Schema = strings.TrimSpace(input.Schema)
	input.SSLMode = strings.ToLower(strings.TrimSpace(input.SSLMode))
	if input.Kind == "postgresql" {
		if input.Schema == "" {
			input.Schema = "public"
		}
		if input.SSLMode == "" {
			input.SSLMode = "prefer"
		}
		if input.ConnectTimeout == 0 {
			input.ConnectTimeout = 10
		}
	}
	if message := validateConnector(input); message != "" {
		return ctx.Response().Json(stdhttp.StatusUnprocessableEntity, http.Json{"message": message})
	}

	passwordEncrypted, err := encryptConnectorSecret(input.Password)
	if err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível proteger a senha"})
	}
	sslRootCertEncrypted, err := encryptConnectorSecret(input.SSLRootCert)
	if err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível proteger o certificado raiz"})
	}
	sslCertEncrypted, err := encryptConnectorSecret(input.SSLCertificate)
	if err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível proteger o certificado de cliente"})
	}
	sslKeyEncrypted, err := encryptConnectorSecret(input.SSLKey)
	if err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível proteger a chave SSL"})
	}

	connector := models.Connector{
		UserID:               user.ID,
		Name:                 input.Name,
		Kind:                 input.Kind,
		Status:               "configured",
		Host:                 input.Host,
		Port:                 input.Port,
		DatabaseName:         input.Database,
		Username:             input.Username,
		PasswordEncrypted:    passwordEncrypted,
		SchemaName:           input.Schema,
		SSLMode:              input.SSLMode,
		SSLRootCertEncrypted: sslRootCertEncrypted,
		SSLCertEncrypted:     sslCertEncrypted,
		SSLKeyEncrypted:      sslKeyEncrypted,
		ConnectTimeout:       input.ConnectTimeout,
	}
	if err := facades.Orm().Query().Create(&connector); err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível salvar o conector"})
	}

	return ctx.Response().Json(stdhttp.StatusCreated, http.Json{"data": presentConnector(connector)})
}

func (r *ConnectorController) Update(ctx http.Context) http.Response {
	user, response := authenticatedUser(ctx)
	if response != nil {
		return response
	}

	connector, response := ownedConnector(ctx, user.ID)
	if response != nil {
		return response
	}

	var input connectorRequest
	if err := ctx.Request().Bind(&input); err != nil {
		return ctx.Response().Json(stdhttp.StatusBadRequest, http.Json{"message": "Dados do conector inválidos"})
	}
	input.Kind = connector.Kind
	normalizeConnectorInput(&input)
	if message := validateConnector(input); message != "" {
		return ctx.Response().Json(stdhttp.StatusUnprocessableEntity, http.Json{"message": message})
	}

	passwordEncrypted, err := updatedConnectorSecret(connector.PasswordEncrypted, input.Password, input.ClearPassword)
	if err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível proteger a senha"})
	}
	sslRootCertEncrypted, err := updatedConnectorSecret(connector.SSLRootCertEncrypted, input.SSLRootCert, input.ClearSSLRootCert)
	if err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível proteger o certificado raiz"})
	}
	sslCertEncrypted, sslKeyEncrypted := connector.SSLCertEncrypted, connector.SSLKeyEncrypted
	if input.ClearSSLClientCertificate {
		sslCertEncrypted, sslKeyEncrypted = "", ""
	} else if input.SSLCertificate != "" || input.SSLKey != "" {
		sslCertEncrypted, err = encryptConnectorSecret(input.SSLCertificate)
		if err != nil {
			return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível proteger o certificado de cliente"})
		}
		sslKeyEncrypted, err = encryptConnectorSecret(input.SSLKey)
		if err != nil {
			return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível proteger a chave SSL"})
		}
	}

	connector.Name = input.Name
	connector.Status = "configured"
	connector.Host = input.Host
	connector.Port = input.Port
	connector.DatabaseName = input.Database
	connector.Username = input.Username
	connector.PasswordEncrypted = passwordEncrypted
	connector.SchemaName = input.Schema
	connector.SSLMode = input.SSLMode
	connector.SSLRootCertEncrypted = sslRootCertEncrypted
	connector.SSLCertEncrypted = sslCertEncrypted
	connector.SSLKeyEncrypted = sslKeyEncrypted
	connector.ConnectTimeout = input.ConnectTimeout
	connector.LastSyncAt = nil
	if err := facades.Orm().Query().Save(&connector); err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível atualizar o conector"})
	}

	return ctx.Response().Success().Json(http.Json{"data": presentConnector(connector)})
}

func (r *ConnectorController) Destroy(ctx http.Context) http.Response {
	user, response := authenticatedUser(ctx)
	if response != nil {
		return response
	}

	connector, response := ownedConnector(ctx, user.ID)
	if response != nil {
		return response
	}
	if _, err := facades.Orm().Query().Delete(&connector); err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível excluir o conector"})
	}

	return ctx.Response().NoContent()
}

func (r *ConnectorController) Test(ctx http.Context) http.Response {
	user, response := authenticatedUser(ctx)
	if response != nil {
		return response
	}

	connectorID, err := strconv.ParseUint(ctx.Request().Route("id"), 10, 64)
	if err != nil {
		return ctx.Response().Json(stdhttp.StatusBadRequest, http.Json{"message": "Conector inválido"})
	}

	var connector models.Connector
	if err := facades.Orm().Query().Where("id", connectorID).Where("user_id", user.ID).First(&connector); err != nil {
		return ctx.Response().Json(stdhttp.StatusNotFound, http.Json{"message": "Conector não encontrado"})
	}
	if connector.Kind != "postgresql" {
		return ctx.Response().Json(stdhttp.StatusUnprocessableEntity, http.Json{"message": "Teste ainda não disponível para este tipo de conector"})
	}

	configuration, err := decryptPostgreSQLConfiguration(connector)
	if err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Não foi possível ler as credenciais do conector"})
	}
	result, err := services.TestPostgreSQLConnection(ctx.Context(), configuration)
	if err != nil {
		connector.Status = "attention"
		_ = facades.Orm().Query().Save(&connector)
		return ctx.Response().Json(stdhttp.StatusUnprocessableEntity, http.Json{"message": err.Error()})
	}

	testedAt := time.Now()
	connector.Status = "connected"
	connector.LastSyncAt = &testedAt
	if err := facades.Orm().Query().Save(&connector); err != nil {
		return ctx.Response().Json(stdhttp.StatusInternalServerError, http.Json{"message": "Conexão válida, mas não foi possível atualizar o conector"})
	}

	return ctx.Response().Success().Json(http.Json{"data": http.Json{
		"result":    result,
		"connector": presentConnector(connector),
	}})
}

func ownedConnector(ctx http.Context, userID uint) (models.Connector, http.Response) {
	connectorID, err := strconv.ParseUint(ctx.Request().Route("id"), 10, 64)
	if err != nil {
		return models.Connector{}, ctx.Response().Json(stdhttp.StatusBadRequest, http.Json{"message": "Conector inválido"})
	}

	var connector models.Connector
	if err := facades.Orm().Query().Where("id", connectorID).Where("user_id", userID).FindOrFail(&connector); err != nil {
		return connector, ctx.Response().Json(stdhttp.StatusNotFound, http.Json{"message": "Conector não encontrado"})
	}

	return connector, nil
}

func normalizeConnectorInput(input *connectorRequest) {
	input.Name = strings.TrimSpace(input.Name)
	input.Kind = strings.ToLower(strings.TrimSpace(input.Kind))
	input.Host = strings.TrimSpace(input.Host)
	input.Database = strings.TrimSpace(input.Database)
	input.Username = strings.TrimSpace(input.Username)
	input.Schema = strings.TrimSpace(input.Schema)
	input.SSLMode = strings.ToLower(strings.TrimSpace(input.SSLMode))
	if input.Kind == "postgresql" {
		if input.Schema == "" {
			input.Schema = "public"
		}
		if input.SSLMode == "" {
			input.SSLMode = "prefer"
		}
		if input.ConnectTimeout == 0 {
			input.ConnectTimeout = 10
		}
	}
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
	if input.Kind == "postgresql" {
		validSSLModes := map[string]bool{"disable": true, "allow": true, "prefer": true, "require": true, "verify-ca": true, "verify-full": true}
		if input.Username == "" {
			return "Informe o usuário PostgreSQL"
		}
		if !validSSLModes[input.SSLMode] {
			return "Modo SSL inválido"
		}
		if input.ConnectTimeout < 1 || input.ConnectTimeout > 60 {
			return "O timeout deve estar entre 1 e 60 segundos"
		}
		if (input.SSLCertificate == "") != (input.SSLKey == "") {
			return "Certificado e chave SSL de cliente devem ser informados juntos"
		}
		if input.SSLMode == "disable" && (input.SSLRootCert != "" || input.SSLCertificate != "") {
			return "Certificados SSL não podem ser usados com SSL desabilitado"
		}
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
		lastSync = "agora"
		if time.Since(*connector.LastSyncAt) >= time.Minute {
			lastSync = connector.LastSyncAt.Format(time.DateTime)
		}
	}

	return connectorResponse{
		ID: connector.ID, Name: connector.Name, Kind: connector.Kind, Detail: detail, Status: connector.Status, LastSync: lastSync,
		Configuration: connectorConfigurationResponse{
			Host: connector.Host, Port: connector.Port, Database: connector.DatabaseName,
			Username: connector.Username, Schema: connector.SchemaName, SSLMode: connector.SSLMode, ConnectTimeout: connector.ConnectTimeout,
			HasPassword: connector.PasswordEncrypted != "", HasSSLRootCert: connector.SSLRootCertEncrypted != "", HasSSLClientCertificate: connector.SSLCertEncrypted != "" && connector.SSLKeyEncrypted != "",
		},
	}
}

func encryptConnectorSecret(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	return facades.Crypt().EncryptString(value)
}

func updatedConnectorSecret(current, replacement string, clear bool) (string, error) {
	if clear {
		return "", nil
	}
	if replacement == "" {
		return current, nil
	}
	return encryptConnectorSecret(replacement)
}

func decryptConnectorSecret(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	return facades.Crypt().DecryptString(value)
}

func decryptPostgreSQLConfiguration(connector models.Connector) (services.PostgreSQLConfig, error) {
	password, err := decryptConnectorSecret(connector.PasswordEncrypted)
	if err != nil {
		return services.PostgreSQLConfig{}, err
	}
	rootCert, err := decryptConnectorSecret(connector.SSLRootCertEncrypted)
	if err != nil {
		return services.PostgreSQLConfig{}, err
	}
	certificate, err := decryptConnectorSecret(connector.SSLCertEncrypted)
	if err != nil {
		return services.PostgreSQLConfig{}, err
	}
	key, err := decryptConnectorSecret(connector.SSLKeyEncrypted)
	if err != nil {
		return services.PostgreSQLConfig{}, err
	}

	port := 5432
	if connector.Port != nil {
		port = *connector.Port
	}
	return services.PostgreSQLConfig{
		Host: connector.Host, Port: port, Database: connector.DatabaseName, Username: connector.Username, Password: password,
		Schema: connector.SchemaName, SSLMode: connector.SSLMode, SSLRootCert: rootCert, SSLCertificate: certificate, SSLKey: key, ConnectTimeout: connector.ConnectTimeout,
	}, nil
}
