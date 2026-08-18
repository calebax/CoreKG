package migrator

func init() {
	Register(ExampleMigrate, &exampleMigrator{})
	Register(DepartmentOrgMigrate, &DepartmentMigrator{})
	// Register(FileNameOrgMigrate, &FileNameMigrator{})
	// Register(GraphTagNodeMigrate, &TagNodeMigrator{})
	Register(CozeModelSyncMigrate, &CozeModelSyncMigrator{})
}
