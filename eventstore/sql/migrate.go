package sql

import "context"

// Migrate the database
func (s *SQL) Migrate() error {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil
	}
	defer tx.Rollback()
	sqlStmt := `
	create table events (seq INTEGER PRIMARY KEY AUTOINCREMENT, id VARCHAR NOT NULL, version INTEGER, reason VARCHAR, type VARCHAR, timestamp VARCHAR, data BLOB, metadata BLOB);
	create unique index id_type_version on events (id, type, version);
	create index id_type on events (id, type);
	delete from events;
	`
	_, err = tx.Exec(sqlStmt)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// MigrateTest remove the index that the test sql driver does not support
func (s *SQL) MigrateTest() error {
	sqlStmt := []string{
		`create table events (seq INTEGER PRIMARY KEY AUTOINCREMENT, id VARCHAR NOT NULL, version INTEGER, reason VARCHAR, type VARCHAR, timestamp VARCHAR, data BLOB, metadata BLOB);`,
		`delete from events;`,
	}

	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil
	}
	defer tx.Rollback()
	for _, b := range sqlStmt {
		_, err := tx.Exec(b)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
