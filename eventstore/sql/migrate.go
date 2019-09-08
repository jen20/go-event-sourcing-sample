package sql

func (sql *SQL) Migrate() error {
	sqlStmt := `
	create table events (id varchar not null, version integer, reason varchar, aggregate_type varchar, data varchar max, meta_data varchar max);
	delete from events;
	`
	_, err := sql.db.Exec(sqlStmt)
	return err
}
