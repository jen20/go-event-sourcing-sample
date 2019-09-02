package sql

import (
	"database/sql"
	"github.com/hallgren/eventsourcing"
	"github.com/hallgren/eventsourcing/eventstore"
)

type SQL struct {
	db sql.DB
	serializer eventstore.Serializer
}

func Open(db sql.DB, serializer eventstore.Serializer) *SQL {
	return &SQL{
		db: db,
		serializer: serializer,
	}
}

func (sql *SQL) Close() {
	sql.db.Close()
}

func (sql *SQL) Save(events []eventsourcing.Event) error {
	insert := `Insert into events (id, version, reason, aggregate_type, data, meta_data) values ($1, $2, $3, $4, $5, $6)`
	for _, event := range events {
		d, err := sql.serializer.Serialize(event)
		if err != nil {
			return err
		}
		_, err = sql.db.Exec(insert, event.AggregateRootID, event.Version, event.Reason, event.AggregateType, string(d), "")
		if err != nil {
			return err
		}
	}
	return nil
}

func (sql *SQL) Get(id string, aggregateType string) (events []eventsourcing.Event, err error) {
	selectStm := `Select id from events where id='$1' and aggregate_type='$2'`
	rows, err := sql.db.Query(selectStm, id, aggregateType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var aggregateID string
		if err := rows.Scan(&aggregateID); err != nil {
			return nil, err
		}
		events = append(events, eventsourcing.Event{AggregateRootID: eventsourcing.AggregateRootID(aggregateID)})
	}
	return
}