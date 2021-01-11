package sql

// Migrate the database
func (sql *SQL) Migrate() error {
	sqlStmt := `
	create table snapshots (id VARCHAR NOT NULL, type VARCHAR, data BLOB);
	create unique index id_type on snapshots (id, type);
	delete from snapshots;
	`
	_, err := sql.db.Exec(sqlStmt)
	return err
}

// MigrateTest remove the index that the test sql driver does not support
func (sql *SQL) MigrateTest() error {
	sqlStmt := []string{
		`create table snapshots (id VARCHAR NOT NULL, type VARCHAR, data BLOB);`,
		`delete from snapshots;`,
	}

	for _, b := range sqlStmt {
		_, err := sql.db.Exec(b)
		if err != nil {
			return err
		}
	}
	return nil
}
