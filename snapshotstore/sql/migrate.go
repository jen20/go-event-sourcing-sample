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
	create table snapshots (id VARCHAR NOT NULL, type VARCHAR, data BLOB);
	create unique index id_type on snapshots (id, type);
	delete from snapshots;
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
		`create table snapshots (id VARCHAR NOT NULL, type VARCHAR, data BLOB);`,
		`delete from snapshots;`,
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
