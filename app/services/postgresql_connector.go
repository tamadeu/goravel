package services

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

type PostgreSQLConfig struct {
	Host           string
	Port           int
	Database       string
	Username       string
	Password       string
	Schema         string
	SSLMode        string
	SSLRootCert    string
	SSLCertificate string
	SSLKey         string
	ConnectTimeout int
}

type PostgreSQLTestResult struct {
	Database      string `json:"database"`
	User          string `json:"user"`
	Version       string `json:"version"`
	LatencyMillis int64  `json:"latencyMs"`
}

func TestPostgreSQLConnection(parent context.Context, input PostgreSQLConfig) (PostgreSQLTestResult, error) {
	config, err := postgreSQLConnectionConfig(input)
	if err != nil {
		return PostgreSQLTestResult{}, err
	}

	timeout := time.Duration(input.ConnectTimeout) * time.Second
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	startedAt := time.Now()
	connection, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return PostgreSQLTestResult{}, fmt.Errorf("não foi possível abrir a conexão: %w", err)
	}
	defer connection.Close(context.Background())

	if err := connection.Ping(ctx); err != nil {
		return PostgreSQLTestResult{}, fmt.Errorf("o servidor não respondeu: %w", err)
	}

	var result PostgreSQLTestResult
	if err := connection.QueryRow(ctx, "SELECT current_database(), current_user, version()").Scan(&result.Database, &result.User, &result.Version); err != nil {
		return PostgreSQLTestResult{}, fmt.Errorf("não foi possível consultar metadados: %w", err)
	}
	result.LatencyMillis = time.Since(startedAt).Milliseconds()

	return result, nil
}

func postgreSQLConnectionConfig(input PostgreSQLConfig) (*pgx.ConnConfig, error) {
	query := url.Values{}
	query.Set("sslmode", input.SSLMode)
	query.Set("connect_timeout", strconv.Itoa(input.ConnectTimeout))
	query.Set("application_name", "goravel-data")
	if input.Schema != "" {
		query.Set("search_path", input.Schema)
	}

	connectionURL := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(input.Username, input.Password),
		Host:     net.JoinHostPort(input.Host, strconv.Itoa(input.Port)),
		Path:     input.Database,
		RawQuery: query.Encode(),
	}
	config, err := pgx.ParseConfig(connectionURL.String())
	if err != nil {
		return nil, fmt.Errorf("configuração PostgreSQL inválida: %w", err)
	}

	if input.SSLMode == "disable" {
		return config, nil
	}
	if config.TLSConfig == nil {
		config.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12, ServerName: input.Host}
	} else {
		config.TLSConfig.MinVersion = tls.VersionTLS12
	}

	if input.SSLRootCert != "" {
		rootCAs := x509.NewCertPool()
		if !rootCAs.AppendCertsFromPEM([]byte(input.SSLRootCert)) {
			return nil, fmt.Errorf("certificado raiz SSL inválido")
		}
		config.TLSConfig.RootCAs = rootCAs
	}

	if input.SSLCertificate != "" || input.SSLKey != "" {
		certificate, err := tls.X509KeyPair([]byte(input.SSLCertificate), []byte(input.SSLKey))
		if err != nil {
			return nil, fmt.Errorf("certificado ou chave de cliente inválidos: %w", err)
		}
		config.TLSConfig.Certificates = []tls.Certificate{certificate}
	}

	return config, nil
}
