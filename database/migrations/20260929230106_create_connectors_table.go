package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260929230106CreateConnectorsTable struct{}

// Signature The unique signature for the migration.
func (r *M20260929230106CreateConnectorsTable) Signature() string {
	return "20260929230106_create_connectors_table"
}

// Up Run the migrations.
func (r *M20260929230106CreateConnectorsTable) Up() error {
	if !facades.Schema().HasTable("connectors") {
		return facades.Schema().Create("connectors", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("user_id")
			table.String("name")
			table.String("kind", 32)
			table.String("status", 32).Default("configured")
			table.String("host")
			table.Integer("port").Nullable()
			table.String("database_name").Nullable()
			table.DateTimeTz("last_sync_at").Nullable()
			table.TimestampsTz()
			table.Index("user_id")
			table.Foreign("user_id").References("id").On("users").CascadeOnDelete()
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260929230106CreateConnectorsTable) Down() error {
	return facades.Schema().DropIfExists("connectors")
}
