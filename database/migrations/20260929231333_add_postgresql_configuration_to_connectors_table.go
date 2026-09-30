package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260929231333AddPostgresqlConfigurationToConnectorsTable struct{}

// Signature The unique signature for the migration.
func (r *M20260929231333AddPostgresqlConfigurationToConnectorsTable) Signature() string {
	return "20260929231333_add_postgresql_configuration_to_connectors_table"
}

// Up Run the migrations.
func (r *M20260929231333AddPostgresqlConfigurationToConnectorsTable) Up() error {
	return facades.Schema().Table("connectors", func(table schema.Blueprint) {
		table.String("username").Nullable()
		table.Text("password_encrypted").Nullable()
		table.String("schema_name").Default("public")
		table.String("ssl_mode", 32).Default("prefer")
		table.Text("ssl_root_cert_encrypted").Nullable()
		table.Text("ssl_cert_encrypted").Nullable()
		table.Text("ssl_key_encrypted").Nullable()
		table.Integer("connect_timeout").Default(10)
	})
}

// Down Reverse the migrations.
func (r *M20260929231333AddPostgresqlConfigurationToConnectorsTable) Down() error {
	return facades.Schema().Table("connectors", func(table schema.Blueprint) {
		table.DropColumn("username", "password_encrypted", "schema_name", "ssl_mode", "ssl_root_cert_encrypted", "ssl_cert_encrypted", "ssl_key_encrypted", "connect_timeout")
	})
}
