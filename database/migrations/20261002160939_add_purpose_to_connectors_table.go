package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20261002160939AddPurposeToConnectorsTable struct{}

// Signature The unique signature for the migration.
func (r *M20261002160939AddPurposeToConnectorsTable) Signature() string {
	return "20261002160939_add_purpose_to_connectors_table"
}

// Up Run the migrations.
func (r *M20261002160939AddPurposeToConnectorsTable) Up() error {
	return facades.Schema().Table("connectors", func(table schema.Blueprint) {
		table.String("purpose", 16).Default("source")
	})
}

// Down Reverse the migrations.
func (r *M20261002160939AddPurposeToConnectorsTable) Down() error {
	return facades.Schema().Table("connectors", func(table schema.Blueprint) {
		table.DropColumn("purpose")
	})
}
