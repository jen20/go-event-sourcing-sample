package sql

func (sql *SQL) Migrate() error {
	sqlStmt := `
	create table events (id varchar not null, version integer, reason varchar, aggregate_type varchar, data varchar max, meta_data varchar max);
	create unique index primary_key on events (id, aggregate_type, version);
	create index get_index on events (id, aggregate_type);
	delete from events;
	`
	_, err := sql.db.Exec(sqlStmt)
	return err
}
