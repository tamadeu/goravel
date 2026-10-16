package services

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ClickHouseConfig struct {
	Host           string
	Port           int
	Database       string
	Username       string
	Password       string
	SSLMode        string
	SSLRootCert    string
	SSLCertificate string
	SSLKey         string
	ConnectTimeout int
}

type ClickHouseTestResult struct {
	Database      string `json:"database"`
	User          string `json:"user"`
	Version       string `json:"version"`
	LatencyMillis int64  `json:"latencyMs"`
}

func TestClickHouseConnection(parent context.Context, input ClickHouseConfig) (ClickHouseTestResult, error) {
	startedAt := time.Now()
	var payload struct {
		Data []ClickHouseTestResult `json:"data"`
	}
	if err := queryClickHouseJSON(parent, input, "SELECT currentDatabase() AS database, currentUser() AS user, version() AS version FORMAT JSON", nil, &payload); err != nil {
		return ClickHouseTestResult{}, err
	}
	if len(payload.Data) != 1 || payload.Data[0].Version == "" || payload.Data[0].Database == "" || payload.Data[0].User == "" {
		return ClickHouseTestResult{}, fmt.Errorf("resposta de metadados ClickHouse inválida")
	}
	result := payload.Data[0]
	result.LatencyMillis = time.Since(startedAt).Milliseconds()
	return result, nil
}

func queryClickHouseJSON(parent context.Context, input ClickHouseConfig, query string, parameters url.Values, output any) error {
	transport, err := clickHouseTransport(input)
	if err != nil {
		return err
	}
	defer transport.CloseIdleConnections()

	ctx, cancel := context.WithTimeout(parent, time.Duration(input.ConnectTimeout)*time.Second)
	defer cancel()
	scheme := "http"
	if input.SSLMode == "verify-full" {
		scheme = "https"
	}
	values := url.Values{"database": {input.Database}}
	for name, entries := range parameters {
		values[name] = entries
	}
	endpoint := url.URL{
		Scheme:   scheme,
		Host:     net.JoinHostPort(input.Host, strconv.Itoa(input.Port)),
		Path:     "/",
		RawQuery: values.Encode(),
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), strings.NewReader(query))
	if err != nil {
		return fmt.Errorf("configuração HTTP ClickHouse inválida")
	}
	request.SetBasicAuth(input.Username, input.Password)
	request.Header.Set("Content-Type", "text/plain; charset=utf-8")
	client := &http.Client{
		Transport: transport,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	response, err := client.Do(request)
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("tempo limite excedido ou consulta cancelada")
		}
		return fmt.Errorf("não foi possível conectar ao ClickHouse; verifique host, porta e TLS")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("ClickHouse retornou HTTP %d; verifique credenciais, banco e permissões", response.StatusCode)
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(output); err != nil {
		return fmt.Errorf("resposta de metadados ClickHouse inválida")
	}
	return nil
}

func clickHouseTransport(input ClickHouseConfig) (*http.Transport, error) {
	if input.Host == "" || strings.ContainsAny(input.Host, "/?#@") || input.Port < 1 || input.Port > 65535 || input.Database == "" || input.Username == "" || input.ConnectTimeout < 1 || input.ConnectTimeout > 60 {
		return nil, fmt.Errorf("configuração ClickHouse inválida")
	}
	if input.SSLMode != "disable" && input.SSLMode != "verify-full" {
		return nil, fmt.Errorf("use HTTP ou HTTPS com verificação de certificado")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	if input.SSLMode == "disable" {
		if input.SSLRootCert != "" || input.SSLCertificate != "" || input.SSLKey != "" {
			return nil, fmt.Errorf("certificados SSL exigem HTTPS")
		}
		return transport, nil
	}
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12, ServerName: input.Host}
	if input.SSLRootCert != "" {
		rootCAs := x509.NewCertPool()
		if !rootCAs.AppendCertsFromPEM([]byte(input.SSLRootCert)) {
			return nil, fmt.Errorf("certificado raiz SSL inválido")
		}
		transport.TLSClientConfig.RootCAs = rootCAs
	}
	if input.SSLCertificate != "" || input.SSLKey != "" {
		certificate, err := tls.X509KeyPair([]byte(input.SSLCertificate), []byte(input.SSLKey))
		if err != nil {
			return nil, fmt.Errorf("certificado ou chave de cliente inválidos")
		}
		transport.TLSClientConfig.Certificates = []tls.Certificate{certificate}
	}
	return transport, nil
}
