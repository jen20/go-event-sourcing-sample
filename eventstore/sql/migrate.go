package sql

// Migrate the database
func (sql *SQL) Migrate() error {
	sqlStmt := `
	create table events (id INTEGER PRIMARY KEY AUTOINCREMENT, aggregate_id varchar not null, version integer, reason varchar, aggregate_type varchar, data varchar max, meta_data varchar max);
	create unique index aggregate_id_type_version on events (aggregate_id, aggregate_type, version);
	create index aggregate_id_type on events (aggregate_id, aggregate_type);
	delete from events;
	`
	_, err := sql.db.Exec(sqlStmt)
	return err
}
