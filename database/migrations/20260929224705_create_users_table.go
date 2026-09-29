package migrations

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/app/facades"
)

type M20260929224705CreateUsersTable struct{}

// Signature The unique signature for the migration.
func (r *M20260929224705CreateUsersTable) Signature() string {
	return "20260929224705_create_users_table"
}

// Up Run the migrations.
func (r *M20260929224705CreateUsersTable) Up() error {
	if !facades.Schema().HasTable("users") {
		return facades.Schema().Create("users", func(table schema.Blueprint) {
			table.ID()
			table.String("name")
			table.String("email")
			table.String("password")
			table.TimestampsTz()
			table.Unique("email")
		})
	}

	return nil
}

// Down Reverse the migrations.
func (r *M20260929224705CreateUsersTable) Down() error {
	return facades.Schema().DropIfExists("users")
}
