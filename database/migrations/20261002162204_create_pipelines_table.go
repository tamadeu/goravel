package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20261002162204CreatePipelinesTable struct{}

// Signature The unique signature for the migration.
func (r *M20261002162204CreatePipelinesTable) Signature() string {
	return "20261002162204_create_pipelines_table"
}

// Up Run the migrations.
func (r *M20261002162204CreatePipelinesTable) Up() error {
	if !facades.Schema().HasTable("pipelines") {
		return facades.Schema().Create("pipelines", func(table schema.Blueprint) {
			table.ID()
			table.UnsignedBigInteger("user_id")
			table.String("name")
			table.Text("description").Nullable()
			table.UnsignedBigInteger("source_id")
			table.UnsignedBigInteger("destination_id")
			table.TimestampsTz()
			table.Index("user_id")
			table.Foreign("user_id").References("id").On("users").CascadeOnDelete()
			table.Foreign("source_id").References("id").On("connectors")
			table.Foreign("destination_id").References("id").On("connectors")
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20261002162204CreatePipelinesTable) Down() error {
	return facades.Schema().DropIfExists("pipelines")
}
