package services

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type SourceTable struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
}

type SourceColumn struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}

func ListPostgreSQLTables(parent context.Context, input PostgreSQLConfig) ([]SourceTable, error) {
	tables := make([]SourceTable, 0)
	err := withPostgreSQLMetadata(parent, input, func(ctx context.Context, connection *pgx.Conn) error {
		rows, err := connection.Query(ctx, `SELECT table_schema, table_name FROM information_schema.tables
			WHERE table_schema = $1 AND table_type = 'BASE TABLE'
			AND has_table_privilege(format('%I.%I', table_schema, table_name), 'SELECT') ORDER BY table_name`, postgreSQLSourceSchema(input))
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var table SourceTable
			if err := rows.Scan(&table.Schema, &table.Name); err != nil {
				return err
			}
			tables = append(tables, table)
		}
		return rows.Err()
	})
	return tables, err
}

func ListPostgreSQLColumns(parent context.Context, input PostgreSQLConfig, schema, table string) ([]SourceColumn, error) {
	if schema != postgreSQLSourceSchema(input) || table == "" {
		return nil, fmt.Errorf("tabela fora do schema configurado no conector")
	}
	columns := make([]SourceColumn, 0)
	err := withPostgreSQLMetadata(parent, input, func(ctx context.Context, connection *pgx.Conn) error {
		rows, err := connection.Query(ctx, `SELECT column_name, data_type, is_nullable = 'YES'
			FROM information_schema.columns AS columns
			WHERE table_schema = $1 AND table_name = $2
			AND EXISTS (SELECT 1 FROM information_schema.tables AS tables
				WHERE tables.table_schema = columns.table_schema AND tables.table_name = columns.table_name AND tables.table_type = 'BASE TABLE')
			AND has_column_privilege(format('%I.%I', table_schema, table_name), column_name, 'SELECT') ORDER BY ordinal_position`, schema, table)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var column SourceColumn
			if err := rows.Scan(&column.Name, &column.Type, &column.Nullable); err != nil {
				return err
			}
			columns = append(columns, column)
		}
		return rows.Err()
	})
	return columns, err
}

func postgreSQLSourceSchema(input PostgreSQLConfig) string {
	if input.Schema == "" {
		return "public"
	}
	return input.Schema
}

func withPostgreSQLMetadata(parent context.Context, input PostgreSQLConfig, query func(context.Context, *pgx.Conn) error) error {
	config, err := postgreSQLConnectionConfig(input)
	if err != nil {
		return fmt.Errorf("configuração PostgreSQL inválida")
	}
	ctx, cancel := context.WithTimeout(parent, time.Duration(input.ConnectTimeout)*time.Second)
	defer cancel()
	connection, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("não foi possível consultar o PostgreSQL; verifique conexão e credenciais")
	}
	defer connection.Close(context.Background())
	if err := query(ctx, connection); err != nil {
		return fmt.Errorf("não foi possível consultar os metadados PostgreSQL; verifique permissões e timeout")
	}
	return nil
}

func ListClickHouseTables(parent context.Context, input ClickHouseConfig) ([]SourceTable, error) {
	payload := struct {
		Data []SourceTable `json:"data"`
	}{Data: make([]SourceTable, 0)}
	err := queryClickHouseJSON(parent, input, "SELECT database AS schema, name FROM system.tables WHERE database = {source_database:String} AND is_temporary = 0 ORDER BY name FORMAT JSON", url.Values{"param_source_database": {input.Database}}, &payload)
	return payload.Data, err
}

func ListClickHouseColumns(parent context.Context, input ClickHouseConfig, schema, table string) ([]SourceColumn, error) {
	if schema != input.Database || table == "" {
		return nil, fmt.Errorf("tabela fora do banco configurado no conector")
	}
	var payload struct {
		Data []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"data"`
	}
	if err := queryClickHouseJSON(parent, input, "SELECT name, type FROM system.columns WHERE database = {source_database:String} AND table = {source_table:String} ORDER BY position FORMAT JSON", url.Values{"param_source_database": {schema}, "param_source_table": {table}}, &payload); err != nil {
		return nil, err
	}
	columns := make([]SourceColumn, 0, len(payload.Data))
	for _, column := range payload.Data {
		nullable := strings.HasPrefix(column.Type, "Nullable(") || strings.HasPrefix(column.Type, "LowCardinality(Nullable(")
		columns = append(columns, SourceColumn{Name: column.Name, Type: column.Type, Nullable: nullable})
	}
	return columns, nil
}
