package bootstrap

import (
	"github.com/goravel/framework/contracts/database/schema"

	"goravel/database/migrations"
)

func Migrations() []schema.Migration {
	return []schema.Migration{
		&migrations.M20210101000001CreateJobsTable{},
		&migrations.M20260929224705CreateUsersTable{},
		&migrations.M20260929230106CreateConnectorsTable{},
		&migrations.M20260929231333AddPostgresqlConfigurationToConnectorsTable{},
		&migrations.M20261002160939AddPurposeToConnectorsTable{},
		&migrations.M20261002162204CreatePipelinesTable{},
		&migrations.M20261002164403AddWizardConfigurationToPipelinesTable{},
	}
}
