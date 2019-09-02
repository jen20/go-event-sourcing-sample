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

}

func (sql *SQL) Save(events []eventsourcing.Event) error {
	return nil
}

func (sql *SQL) Get(id string, aggregateType string) ([]eventsourcing.Event, error) {
	return []eventsourcing.Event{}, nil
}