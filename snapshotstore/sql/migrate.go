package sql

import "context"

const create_table = `create table snapshots (id VARCHAR NOT NULL, type VARCHAR, data BLOB);`

// Migrate the database
func (s *SQL) Migrate() error {
	sqlStmt := []string{
		create_table,
		`create unique index id_type on snapshots (id, type);`,
	}
	return s.migrate(sqlStmt)
}

// MigrateTest remove the index that the test sql driver does not support
func (s *SQL) MigrateTest() error {
	sqlStmt := []string{
		create_table,
	}

	return s.migrate(sqlStmt)
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
