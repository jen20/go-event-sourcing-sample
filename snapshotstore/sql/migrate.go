package sql

import "context"

const createTable = `create table snapshots (id VARCHAR NOT NULL, type VARCHAR, version INTEGER, global_version INTEGER, state BLOB);`

// Migrate the database
func (s *SQL) Migrate() error {
	sqlStmt := []string{
		createTable,
		`create unique index id_type on snapshots (id, type);`,
	}
	return s.migrate(sqlStmt)
}

// MigrateTest remove the index that the test sql driver does not support
func (s *SQL) MigrateTest() error {
	return s.migrate([]string{createTable})
}

func (s *SQL) migrate(stm []string) error {
	tx, err := s.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil
	}
	defer tx.Rollback()
	for _, b := range stm {
		_, err := tx.Exec(b)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
