package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20261002164403AddWizardConfigurationToPipelinesTable struct{}

// Signature The unique signature for the migration.
func (r *M20261002164403AddWizardConfigurationToPipelinesTable) Signature() string {
	return "20261002164403_add_wizard_configuration_to_pipelines_table"
}

// Up Run the migrations.
func (r *M20261002164403AddWizardConfigurationToPipelinesTable) Up() error {
	return facades.Schema().Table("pipelines", func(table schema.Blueprint) {
		table.String("source_schema").Default("")
		table.String("source_table").Default("")
		table.String("destination_table", 63).Default("")
		table.Text("columns_json").Nullable()
		table.String("frequency", 32).Default("manual")
	})
}

// Down Reverse the migrations.
func (r *M20261002164403AddWizardConfigurationToPipelinesTable) Down() error {
	return facades.Schema().Table("pipelines", func(table schema.Blueprint) {
		table.DropColumn("source_schema", "source_table", "destination_table", "columns_json", "frequency")
	})
}
