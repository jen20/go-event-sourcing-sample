package sql

// Migrate the database
func (s *SQL) Migrate() error {
	sqlStmt := `
	create table events (seq INTEGER PRIMARY KEY AUTOINCREMENT, id VARCHAR NOT NULL, version INTEGER, reason VARCHAR, type VARCHAR, timestamp VARCHAR, data BLOB, metadata BLOB);
	create unique index id_type_version on events (id, type, version);
	create index id_type on events (id, type);
	delete from events;
	`
	_, err := s.db.Exec(sqlStmt)
	return err
}

// MigrateTest remove the index that the test sql driver does not support
func (s *SQL) MigrateTest() error {
	sqlStmt := []string{
		`create table events (seq INTEGER PRIMARY KEY AUTOINCREMENT, id VARCHAR NOT NULL, version INTEGER, reason VARCHAR, type VARCHAR, timestamp VARCHAR, data BLOB, metadata BLOB);`,
		`delete from events;`,
	}

	for _, b := range sqlStmt {
		_, err := s.db.Exec(b)
		if err != nil {
			return err
		}
	}
	return nil
}
